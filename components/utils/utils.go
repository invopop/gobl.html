// Package utils provides utility functions for the components package.
package utils

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

// Coalesce renders only the first component that is not empty. This is useful for
// rendering variants of a component (e.g. regime-specific), falling back to a default
// version. This allows to encapsulate the conditional logic to render a variant in the
// variant component itself.
func Coalesce(components ...templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for _, c := range components {
			pr := &proxy{w: w}
			err := c.Render(ctx, pr)
			if err != nil {
				return err
			}
			if pr.n > 0 {
				// the component wrote to the buffer, we stop here
				return nil
			}
		}
		// no component wrote to the buffer
		return nil
	})
}

type proxy struct {
	w io.Writer
	n int
}

func (p *proxy) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	if err != nil {
		return n, err
	}
	p.n += n
	return n, nil
}
