package goblhtml_test

import (
	"context"
	"encoding/json"
	"testing"

	goblhtml "github.com/invopop/gobl.html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamsOptions(t *testing.T) {
	env, err := loadExample("es-ticketbai-invoice.json")
	require.NoError(t, err)

	t.Run("nil params render defaults", func(t *testing.T) {
		var p *goblhtml.Params
		out, err := goblhtml.Render(context.Background(), env, p.Options()...)
		require.NoError(t, err)
		assert.Contains(t, string(out), `href="styles/envelope.css"`)
	})

	t.Run("json params match functional options", func(t *testing.T) {
		p := new(goblhtml.Params)
		require.NoError(t, json.Unmarshal([]byte(`{
			"locale": "es",
			"date_format": "%d/%m/%Y",
			"label": "COPY",
			"watermark": "DRAFT",
			"thousands_separator": ".",
			"decimal_mark": ",",
			"void": true,
			"embed_assets": true
		}`), p))

		got, err := goblhtml.Render(context.Background(), env, p.Options()...)
		require.NoError(t, err)

		want, err := goblhtml.Render(context.Background(), env,
			goblhtml.WithLocale("es"),
			goblhtml.WithCalFormatter("02/01/2006", "", nil),
			goblhtml.WithLabel("COPY"),
			goblhtml.WithWatermark("DRAFT"),
			goblhtml.WithThousandsSeparator("."),
			goblhtml.WithDecimalMark(","),
			goblhtml.WithVoid(true),
			goblhtml.WithEmbeddedAssets(),
		)
		require.NoError(t, err)

		assert.Equal(t, string(want), string(got))
		assert.Contains(t, string(got), "<style")
		assert.NotContains(t, string(got), `href="styles/`)
	})
}
