package catalog_test

import (
	"context"
	"testing"

	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/invopop/gobl.html/internal/catalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan(t *testing.T) {
	render := func(ctx context.Context, env *gobl.Envelope) ([]byte, error) {
		return goblhtml.Render(ctx, env)
	}
	entries, errs := catalog.Scan(context.Background(), "../../examples", render)
	require.Empty(t, errs)
	require.NotEmpty(t, entries)

	byFile := make(map[string]catalog.Entry, len(entries))
	for _, e := range entries {
		byFile[e.File] = e
	}

	inv := byFile["es-ticketbai-invoice.json"]
	assert.Equal(t, "ticketbai invoice", inv.Title)
	assert.Equal(t, "es", inv.Country)
	assert.Equal(t, "invoice", inv.Kind)
	assert.Empty(t, inv.Addons)
	assert.Contains(t, inv.Tags, "qr")
	assert.Equal(t, "EUR", inv.Currency)
	assert.NotZero(t, inv.Lines)

	cn := byFile["es-ticketbai-credit-note.adjustment.es.json"]
	assert.Equal(t, "credit-note", cn.Kind)
	assert.Equal(t, "es", cn.Locale)
	assert.True(t, cn.AdjustmentMode)
	assert.Contains(t, cn.Tags, "adjustment")
	assert.Contains(t, cn.Tags, "preceding")
	assert.Equal(t, []string{"es-tbai-v1"}, cn.Addons)

	party := byFile["es-party.json"]
	assert.Equal(t, "party", party.Kind)
	assert.Equal(t, "es", party.Country)

	pay := byFile["payment-multi-method.json"]
	assert.Equal(t, "payment", pay.Kind)
	assert.NotEmpty(t, pay.Country, "country comes from the regime, not the file name")

	logo := byFile["ar-arca-invoice-logo.json"]
	assert.Contains(t, logo.Tags, "logo")

	usd := byFile["es-verifactu-invoice-usd.json"]
	assert.Contains(t, usd.Tags, "currency", "foreign currency with exchange rates")
	assert.NotContains(t, byFile["es-verifactu-invoice.json"].Tags, "currency", "home currency")
	assert.Contains(t, byFile["pt-at-order-full.json"].Tags, "currency", "exchange rates in the regime's currency")

	void := byFile["pt-at-invoice.void.json"]
	assert.True(t, void.Void)
	assert.Contains(t, void.Tags, "void")
}
