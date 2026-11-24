package goblhtml_test

import (
	"testing"

	goblhtml "github.com/invopop/gobl.html"
	"github.com/invopop/gobl/cbc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWarnings(t *testing.T) {
	tests := map[string][]cbc.Key{
		"au-gst-tax-invoice.json":                      nil,
		"au-gst-tax-invoice-inclusive.json":            nil,
		"au-gst-tax-invoice-no-customer-identity.json": {"au-customer-identity"},
		"au-gst-tax-invoice-mixed.json":                nil,
		"au-gst-tax-invoice-usd.json":                  nil,
		"au-gst-adjustment-note-credit.json":           nil,
		"au-gst-rcti.json":                             nil,
		"es-verifactu-invoice.json":                    nil,
	}
	for name, want := range tests {
		t.Run(name, func(t *testing.T) {
			env, err := loadExample(name)
			require.NoError(t, err)
			var got []cbc.Key
			for _, w := range goblhtml.Warnings(env) {
				assert.NotEmpty(t, w.Message)
				got = append(got, w.Code)
			}
			assert.Equal(t, want, got)
		})
	}
}
