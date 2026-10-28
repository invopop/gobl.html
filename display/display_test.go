package display_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"

	"github.com/invopop/gobl.html/display"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeysComplete(t *testing.T) {
	// Collect every constant of type Key declared in the package source.
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "display.go", nil, 0)
	require.NoError(t, err)
	var declared []display.Key
	ast.Inspect(f, func(n ast.Node) bool {
		vs, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		if id, ok := vs.Type.(*ast.Ident); !ok || id.Name != "Key" {
			return true
		}
		for _, v := range vs.Values {
			lit, ok := v.(*ast.BasicLit)
			require.True(t, ok)
			s, err := strconv.Unquote(lit.Value)
			require.NoError(t, err)
			declared = append(declared, display.Key(s))
		}
		return true
	})
	require.NotEmpty(t, declared)
	assert.ElementsMatch(t, declared, display.Keys())
	for _, k := range display.Keys() {
		assert.True(t, k.Valid(), k)
	}
}

func TestOptionsNil(t *testing.T) {
	var o *display.Options
	assert.NoError(t, o.Validate())
	assert.False(t, o.Hidden(display.KeyPayment))
	assert.False(t, o.AncestorHidden(display.KeyPayment))
	assert.False(t, o.Suppressed(display.KeyPayment))
	assert.False(t, o.HasSnippets(display.KeyPayment))
	_, ok := o.Replacement(display.KeyPayment)
	assert.False(t, ok)
	assert.Empty(t, o.Before(display.KeyPayment))
	assert.Empty(t, o.After(display.KeyPayment))
}

func TestHiddenCascade(t *testing.T) {
	o := &display.Options{Hide: []display.Key{display.KeyPayment}}
	assert.True(t, o.Hidden(display.KeyPayment))
	assert.False(t, o.AncestorHidden(display.KeyPayment))
	assert.True(t, o.Hidden(display.KeyPaymentInstructions))
	assert.True(t, o.AncestorHidden(display.KeyPaymentInstructions))
	assert.True(t, o.Hidden(display.KeyPaymentBankData))
	assert.True(t, o.AncestorHidden(display.KeyPaymentBankData))
	assert.False(t, o.Hidden(display.KeyTaxes))

	o = &display.Options{Hide: []display.Key{display.KeyPaymentInstructions}}
	assert.False(t, o.Hidden(display.KeyPayment))
	assert.True(t, o.Hidden(display.KeyPaymentOnline))
	assert.False(t, o.Hidden(display.KeyPaymentTerms))
}

func TestHiddenPrefixRequiresDot(t *testing.T) {
	o := &display.Options{Hide: []display.Key{"pay"}}
	assert.False(t, o.Hidden(display.KeyPayment))
	assert.False(t, o.Hidden(display.KeyPaymentTerms))
}

func TestHiddenUnknownKeysIgnored(t *testing.T) {
	o := &display.Options{
		Hide: []display.Key{"foo.bar"},
		HTML: []*display.HTML{{At: "foo", Mode: display.ModeReplace, Content: "x"}},
	}
	assert.False(t, o.Hidden(display.KeyPayment))
	assert.True(t, o.Hidden("foo.bar.baz"))
	assert.True(t, o.Hidden("foo.bar"))
}

func TestReplaceCascade(t *testing.T) {
	o := &display.Options{
		HTML: []*display.HTML{
			{At: display.KeyPayment, Mode: display.ModeReplace, Content: "<p>custom</p>"},
		},
	}
	s, ok := o.Replacement(display.KeyPayment)
	assert.True(t, ok)
	assert.Equal(t, "<p>custom</p>", s)
	assert.False(t, o.Hidden(display.KeyPayment))
	assert.True(t, o.Suppressed(display.KeyPayment))
	assert.True(t, o.Hidden(display.KeyPaymentInstructions))
	assert.True(t, o.AncestorHidden(display.KeyPaymentBankData))
	_, ok = o.Replacement(display.KeyPaymentInstructions)
	assert.False(t, ok)
}

func TestBeforeAfter(t *testing.T) {
	o := &display.Options{
		Hide: []display.Key{display.KeyTaxes},
		HTML: []*display.HTML{
			{At: display.KeyTaxes, Mode: display.ModeAfter, Content: "a1"},
			{At: display.KeyTaxes, Mode: display.ModeBefore, Content: "b1"},
			{At: display.KeyTaxes, Mode: display.ModeAfter, Content: "a2"},
			{At: display.KeyPayment, Mode: display.ModeBefore, Content: "p"},
			nil,
		},
	}
	assert.Equal(t, []string{"b1"}, o.Before(display.KeyTaxes))
	assert.Equal(t, []string{"a1", "a2"}, o.After(display.KeyTaxes))
	assert.True(t, o.HasSnippets(display.KeyTaxes))
	assert.False(t, o.HasSnippets(display.KeyFooterBranding))
}

func TestHasSnippetsAncestorHidden(t *testing.T) {
	o := &display.Options{
		Hide: []display.Key{display.KeyPayment},
		HTML: []*display.HTML{
			{At: display.KeyPaymentTerms, Mode: display.ModeAfter, Content: "x"},
		},
	}
	assert.False(t, o.HasSnippets(display.KeyPaymentTerms))
}

func TestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		o := &display.Options{
			Hide: []display.Key{display.KeyPaymentBankData, display.KeyPaymentQR},
			HTML: []*display.HTML{
				{At: display.KeyPayment, Mode: display.ModeReplace, Content: "<p>x</p>"},
				{At: display.KeyPayment, Mode: display.ModeBefore, Content: "<p>y</p>"},
				{At: display.KeyTaxes, Mode: display.ModeReplace, Content: "<p>z</p>"},
			},
		}
		assert.NoError(t, o.Validate())
	})
	t.Run("unknown hide key", func(t *testing.T) {
		o := &display.Options{Hide: []display.Key{"totals"}}
		assert.ErrorContains(t, o.Validate(), "hide: 0: unknown key 'totals'")
	})
	t.Run("unknown html key", func(t *testing.T) {
		o := &display.Options{HTML: []*display.HTML{{At: "lines", Mode: display.ModeAfter}}}
		assert.ErrorContains(t, o.Validate(), "html: 0: at: unknown key 'lines'")
	})
	t.Run("bad mode", func(t *testing.T) {
		o := &display.Options{HTML: []*display.HTML{{At: display.KeyTaxes, Mode: "inside"}}}
		assert.ErrorContains(t, o.Validate(), "html: 0: mode: unknown mode 'inside'")
	})
	t.Run("missing snippet", func(t *testing.T) {
		o := &display.Options{HTML: []*display.HTML{nil}}
		assert.ErrorContains(t, o.Validate(), "html: 0: missing")
	})
	t.Run("duplicate replace", func(t *testing.T) {
		o := &display.Options{HTML: []*display.HTML{
			{At: display.KeyTaxes, Mode: display.ModeReplace, Content: "a"},
			{At: display.KeyTaxes, Mode: display.ModeReplace, Content: "b"},
		}}
		assert.ErrorContains(t, o.Validate(), "html: 1: duplicate replace for 'taxes'")
	})
	t.Run("snippet too large", func(t *testing.T) {
		o := &display.Options{HTML: []*display.HTML{
			{At: display.KeyTaxes, Mode: display.ModeAfter, Content: strings.Repeat("x", display.MaxSnippetSize+1)},
		}}
		assert.ErrorContains(t, o.Validate(), "html: 0: content: exceeds 32768 bytes")
	})
	t.Run("total too large", func(t *testing.T) {
		o := new(display.Options)
		for range 5 {
			o.HTML = append(o.HTML, &display.HTML{
				At: display.KeyTaxes, Mode: display.ModeAfter, Content: strings.Repeat("x", display.MaxSnippetSize),
			})
		}
		err := o.Validate()
		assert.ErrorContains(t, err, "html: total content exceeds 131072 bytes")
		assert.NotContains(t, err.Error(), "html: 0: content")
	})
	t.Run("multiple errors", func(t *testing.T) {
		o := &display.Options{
			Hide: []display.Key{"x"},
			HTML: []*display.HTML{{At: "y", Mode: "z"}},
		}
		err := o.Validate()
		assert.ErrorContains(t, err, "hide: 0")
		assert.ErrorContains(t, err, "html: 0: at")
		assert.ErrorContains(t, err, "html: 0: mode")
	})
}
