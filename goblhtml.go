package goblhtml

import (
	"bytes"
	"context"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl.html/components"
)

// Render takes the GOBL envelope an attempts to render an HTML document
// from it.
func Render(ctx context.Context, env *gobl.Envelope) ([]byte, error) {
	out := components.Envelope(env)

	buf := new(bytes.Buffer)
	if err := out.Render(ctx, buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
