package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	goblhtml "github.com/invopop/gobl.html"
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

func TestApplyExampleParams(t *testing.T) {
	none := func(string) bool { return false }

	p := new(goblhtml.Params)
	applyExampleParams("pt-at-invoice.void.json", p, none)
	assert.True(t, p.Void, "file name implies void")
	assert.False(t, p.Sandbox)

	p = new(goblhtml.Params) // void=false in the query
	applyExampleParams("pt-at-invoice.void.json", p, func(k string) bool { return k == "void" })
	assert.False(t, p.Void, "an explicit false keeps the file name from forcing void")

	p = &goblhtml.Params{AdjustmentMode: true}
	applyExampleParams("plain.json", p, none)
	assert.True(t, p.AdjustmentMode, "request values survive when the file name implies nothing")
}
