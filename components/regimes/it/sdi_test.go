package it

import (
	"testing"

	"github.com/invopop/gobl.html/internal"
	sdi "github.com/invopop/gobl.it.sdi/addon"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
)

func TestIsSDICorrection(t *testing.T) {
	invoice := func(code cbc.Code, total string) internal.Document {
		amount, err := num.AmountFromString(total)
		if err != nil {
			t.Fatal(err)
		}
		inv := &bill.Invoice{Totals: &bill.Totals{Total: amount}}
		if code != "" {
			inv.Tax = &bill.Tax{Ext: tax.ExtensionsOf(cbc.CodeMap{sdi.ExtKeyDocumentType: code})}
		}
		return internal.DocumentFor(inv)
	}

	tests := []struct {
		name string
		doc  internal.Document
		want bool
	}{
		{"negative TD17", invoice("TD17", "-200.00"), true},
		{"negative TD28", invoice("TD28", "-50.00"), true},
		{"positive TD17", invoice("TD17", "1000.00"), false},
		{"negative TD24, corrected with a TD04", invoice("TD24", "-200.00"), false},
		{"negative TD01", invoice("TD01", "-200.00"), false},
		{"no SDI document type", invoice("", "-200.00"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isSDICorrection(tt.doc))
		})
	}
}
