package goblhtml

import (
	"bytes"
	"context"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl.html/components"
	"github.com/invopop/gobl.html/components/t"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/currency"
)

// Render takes the GOBL envelope and attempts to render an HTML document
// from it.
func Render(ctx context.Context, env *gobl.Envelope) ([]byte, error) {
	out := components.Envelope(env)

	cur := currency.EUR
	// Extract the currency to use for formatting
	switch doc := env.Extract().(type) {
	case *bill.Invoice:
		cur = doc.Currency
	}
	ctx = t.WithFormatter(ctx, cur.Def().Formatter())

	buf := new(bytes.Buffer)
	if err := out.Render(ctx, buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
