package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	goblhtml "github.com/invopop/gobl.html"
	"github.com/invopop/gobl.html/display"
	"github.com/invopop/gobl.html/internal/catalog"
	"github.com/invopop/gobl.html/layout"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOptionsBindQuery checks Echo binds the query into the embedded Params,
// which is what the previewer and the benchmark rely on.
func TestOptionsBindQuery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/x.html?locale=es&layout=Letter&embed_assets=true&void=false&watermark=DRAFT&thousands_separator=.", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	c.SetParamNames("filename")
	c.SetParamValues("x.html")

	o := new(options)
	require.NoError(t, c.Bind(o))
	assert.Equal(t, "x.html", o.Filename)
	assert.EqualValues(t, "es", o.Locale)
	assert.Equal(t, layout.Letter, o.Layout)
	assert.True(t, o.EmbedAssets)
	assert.False(t, o.Void)
	assert.Equal(t, "DRAFT", o.Watermark)
	assert.Equal(t, ".", o.ThousandsSeparator)
}

// TestOptionsBindSnippets checks the snippets arrive through a query
// parameter as a JSON string, the way the previewer sends them.
func TestOptionsBindSnippets(t *testing.T) {
	e := echo.New()
	q := url.Values{"snippets": {`[{"at": "notes", "mode": "after", "html": "<p>{{ .Number }}</p>"}]`}}
	req := httptest.NewRequest(http.MethodGet, "/x.html?"+q.Encode(), nil)
	c := e.NewContext(req, httptest.NewRecorder())
	c.SetParamNames("filename")
	c.SetParamValues("x.html")

	o := new(options)
	require.NoError(t, c.Bind(o))
	assert.Equal(t, display.Snippets{{At: display.KeyNotes, Mode: display.ModeAfter, HTML: "<p>{{ .Number }}</p>"}}, o.Snippets)

	req = httptest.NewRequest(http.MethodGet, "/x.html?snippets=nope", nil)
	c = e.NewContext(req, httptest.NewRecorder())
	assert.Error(t, c.Bind(new(options)))
}

func TestApplyExampleParams(t *testing.T) {
	none := func(string) bool { return false }

	p := new(goblhtml.Params)
	applyExampleParams("pt-at-invoice.void.json", p, none)
	assert.True(t, p.Void, "file name implies void")
	assert.False(t, p.Sandbox)

	p = new(goblhtml.Params) // void=false in the query
	applyExampleParams("pt-at-invoice.void.json", p, func(k string) bool { return k == "void" })
	assert.False(t, p.Void, "an explicit false keeps the file name from forcing void")

	p = new(goblhtml.Params)
	applyExampleParams("es-verifactu-invoice-full.hidden.json", p, none)
	assert.Equal(t, strings.Join(catalog.HiddenKeys, ","), p.Hide, "file name implies the hidden sections")

	p = new(goblhtml.Params) // hide= in the query
	applyExampleParams("es-verifactu-invoice-full.hidden.json", p, func(k string) bool { return k == "hide" })
	assert.Empty(t, p.Hide, "an explicit empty hide keeps the file name from hiding anything")

	p = new(goblhtml.Params)
	applyExampleParams("es-verifactu-invoice-full.snippets.json", p, none)
	assert.Equal(t, catalog.ExampleSnippets, p.Snippets, "file name implies the example snippets")

	p = new(goblhtml.Params) // snippets= in the query
	applyExampleParams("es-verifactu-invoice-full.snippets.json", p, func(k string) bool { return k == "snippets" })
	assert.Empty(t, p.Snippets, "an explicit empty snippets list keeps the file name from adding any")

	p = &goblhtml.Params{AdjustmentMode: true}
	applyExampleParams("plain.json", p, none)
	assert.True(t, p.AdjustmentMode, "request values survive when the file name implies nothing")
}
