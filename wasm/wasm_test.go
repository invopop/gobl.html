//go:build !js

package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const examplesDir = "../examples"

// TestWASMRendersExamples compiles the renderer to WebAssembly, renders every
// example through it in Node, and checks the output is byte-identical to the
// native renderer with the same parameters. It needs node on the PATH.
func TestWASMRendersExamples(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping WebAssembly build in short mode")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not found on PATH")
	}

	dir := t.TempDir()
	wasmPath := filepath.Join(dir, "gobl.html.wasm")
	outDir := filepath.Join(dir, "out")

	build := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", wasmPath, ".")
	build.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building wasm: %v\n%s", err, out)
	}
	root, err := exec.Command("go", "env", "GOROOT").Output()
	require.NoError(t, err, "locating GOROOT")
	execJS, err := os.ReadFile(filepath.Join(strings.TrimSpace(string(root)), "lib", "wasm", "wasm_exec.js"))
	require.NoError(t, err, "reading wasm_exec.js from GOROOT")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "wasm_exec.js"), execJS, 0o644))

	run := exec.Command(node, "render.mjs", wasmPath, examplesDir, outDir)
	out, err := run.CombinedOutput()
	t.Logf("%s", out)
	require.NoError(t, err, "rendering through wasm")

	examples, err := filepath.Glob(filepath.Join(examplesDir, "*.json"))
	require.NoError(t, err)
	require.NotEmpty(t, examples)

	for _, example := range examples {
		name := filepath.Base(example)
		t.Run(name, func(t *testing.T) {
			want := renderNative(t, example)
			got, err := os.ReadFile(filepath.Join(outDir, strings.TrimSuffix(name, ".json")+".html"))
			require.NoError(t, err, "wasm output missing")
			assert.Equal(t, string(want), string(got))
		})
	}
}

// renderNative renders an example the way wasm/render.mjs does: no
// calculation, embedded assets, and the state implied by the file name.
func renderNative(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	env := new(gobl.Envelope)
	require.NoError(t, json.Unmarshal(data, env))

	name := filepath.Base(path)
	p := &goblhtml.Params{EmbedAssets: true}
	p.Void = strings.Contains(name, ".void.")
	p.Sandbox = strings.Contains(name, ".sandbox.")
	p.AdjustmentMode = strings.Contains(name, ".adjustment.")
	for _, l := range []string{"es", "fr", "pl"} {
		if strings.Contains(name, "."+l+".") {
			p.Locale = i18n.Code(l)
		}
	}

	out, err := goblhtml.Render(context.Background(), env, p.Options()...)
	require.NoError(t, err)
	return bytes.Clone(out)
}
