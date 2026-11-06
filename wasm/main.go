//go:build js && wasm

// Command wasm exposes the gobl.html renderer to the browser:
//
//	goblRender(envelopeJSON, params) → { html } or { error }
//	goblVersion()                    → { version, gobl }
//
// envelopeJSON is a GOBL envelope as a string. The document is rendered as
// provided, without calculation or validation, so it must already be
// complete. params is optional: either an object or a JSON string in the
// shape of goblhtml.Params; stylesheets and scripts are embedded unless
// embed_assets is explicitly false.
//
// A "gobl:ready" event is dispatched on the global object once the functions
// are installed.
//
//	GOOS=js GOARCH=wasm go build -o gobl.html.wasm ./wasm
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
)

// version is set with -ldflags "-X main.version=v1.2.3" when publishing.
var version = "dev"

func main() {
	js.Global().Set("goblRender", js.FuncOf(render))
	js.Global().Set("goblVersion", js.FuncOf(func(js.Value, []js.Value) any {
		return map[string]any{
			"version": version,
			"gobl":    string(gobl.VERSION),
		}
	}))
	js.Global().Call("dispatchEvent", js.Global().Get("Event").New("gobl:ready"))

	select {} // keep the Go runtime alive for callbacks
}

func render(_ js.Value, args []js.Value) (result any) {
	defer func() {
		if r := recover(); r != nil {
			result = failure(fmt.Errorf("rendering: %v", r))
		}
	}()

	if len(args) < 1 || args[0].Type() != js.TypeString {
		return failure(fmt.Errorf("goblRender expects the envelope as a JSON string"))
	}
	env := new(gobl.Envelope)
	if err := json.Unmarshal([]byte(args[0].String()), env); err != nil {
		return failure(fmt.Errorf("parsing envelope: %w", err))
	}

	params, err := parseParams(args[1:])
	if err != nil {
		return failure(err)
	}

	out, err := goblhtml.Render(context.Background(), env, params.Options()...)
	if err != nil {
		return failure(fmt.Errorf("rendering: %w", err))
	}
	return map[string]any{"html": string(out)}
}

// parseParams reads the optional params argument, which may be a JS object
// or a JSON string. Assets are embedded unless the caller says otherwise.
func parseParams(args []js.Value) (*goblhtml.Params, error) {
	params := &goblhtml.Params{EmbedAssets: true}
	if len(args) == 0 {
		return params, nil
	}
	var src string
	switch args[0].Type() {
	case js.TypeUndefined, js.TypeNull:
		return params, nil
	case js.TypeString:
		src = args[0].String()
	case js.TypeObject:
		src = js.Global().Get("JSON").Call("stringify", args[0]).String()
	default:
		return nil, fmt.Errorf("goblRender expects params as an object or JSON string")
	}
	if src == "" {
		return params, nil
	}
	// Decode over a copy so an explicit "embed_assets": false is honoured
	// while a missing field keeps the default.
	if err := json.Unmarshal([]byte(src), params); err != nil {
		return nil, fmt.Errorf("parsing params: %w", err)
	}
	return params, nil
}

func failure(err error) map[string]any {
	return map[string]any{"error": err.Error()}
}
