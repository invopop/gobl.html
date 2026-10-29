package main

import (
	"testing"

	goblhtml "github.com/invopop/gobl.html"
	"github.com/stretchr/testify/assert"
)

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
