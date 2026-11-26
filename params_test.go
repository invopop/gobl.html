package goblhtml_test

import (
	"context"
	"encoding/json"
	"testing"

	goblhtml "github.com/invopop/gobl.html"
	"github.com/invopop/gobl.html/display"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamsOptions(t *testing.T) {
	env, err := loadExample("es-tbai-invoice.json")
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
			"hide": "payment,notes",
			"snippets": [{"at": "taxes", "mode": "after", "html": "<p>{{ .Number }}</p>"}],
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
			goblhtml.WithHide("payment", "notes"),
			goblhtml.WithSnippets(display.Snippets{{At: display.KeyTaxes, Mode: display.ModeAfter, HTML: "<p>{{ .Number }}</p>"}}),
			goblhtml.WithThousandsSeparator("."),
			goblhtml.WithDecimalMark(","),
			goblhtml.WithVoid(true),
			goblhtml.WithEmbeddedAssets(),
		)
		require.NoError(t, err)

		assert.Equal(t, string(want), string(got))
		assert.Contains(t, string(got), `<div class="snippet snippet-after" data-at="taxes"><p>`)
		assert.Contains(t, string(got), "<style")
		assert.NotContains(t, string(got), `href="styles/`)
	})
}

func TestParamsEmbedAssetsFalseSurvivesJSON(t *testing.T) {
	// The WebAssembly entry point defaults embed_assets to true when the key
	// is absent, so a marshalled false must keep the key.
	b, err := json.Marshal(&goblhtml.Params{EmbedAssets: false})
	require.NoError(t, err)
	assert.Contains(t, string(b), `"embed_assets":false`)
}

func TestParamsSnippetsAsString(t *testing.T) {
	// A query parameter or form field carries the list as a JSON string,
	// which decodes the same as the array.
	p := new(goblhtml.Params)
	require.NoError(t, json.Unmarshal([]byte(`{"snippets": "[{\"at\": \"notes\", \"mode\": \"before\", \"html\": \"<p>x</p>\"}]"}`), p))
	assert.Equal(t, display.Snippets{{At: display.KeyNotes, Mode: display.ModeBefore, HTML: "<p>x</p>"}}, p.Snippets)

	b, err := json.Marshal(&goblhtml.Params{})
	require.NoError(t, err)
	assert.NotContains(t, string(b), "snippets", "an empty list is omitted")
}
