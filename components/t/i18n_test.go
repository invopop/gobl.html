package t_test

import (
	"context"
	"testing"
	"time"

	calT "github.com/invopop/gobl.html/components/t"
	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/num"
	"github.com/stretchr/testify/assert"
)

func TestLocalizeCurrency(t *testing.T) {
	amount := num.MakeAmount(123456, 2)

	t.Run("no options in context", func(t *testing.T) {
		got := calT.LocalizeCurrency(context.Background(), amount, currency.EUR)
		assert.Equal(t, "€1.234,56", got)
	})

	t.Run("no formatter in options", func(t *testing.T) {
		ctx := internal.WithOptions(context.Background(), &internal.Opts{})
		got := calT.LocalizeCurrency(ctx, amount, currency.EUR)
		assert.Equal(t, "€1.234,56", got)
	})

	t.Run("uses document layout with target currency symbol", func(t *testing.T) {
		// A GBP document formatted with a custom layout, e.g. "3,99 £",
		// converting an amount to EUR should keep the same layout: "1 234,56 €".
		nf := currency.GBP.Def().Formatter()
		nf.Template = "%n %u"
		nf.ThousandsSeparator = " "
		nf.DecimalMark = ","
		ctx := internal.WithOptions(context.Background(), &internal.Opts{
			NumFormatter: &nf,
		})
		got := calT.LocalizeCurrency(ctx, amount, currency.EUR)
		assert.Equal(t, "1 234,56 €", got)
	})

	t.Run("negative amounts follow the document layout", func(t *testing.T) {
		nf := currency.GBP.Def().Formatter()
		nf.NegativeTemplate = "(%u%n)"
		ctx := internal.WithOptions(context.Background(), &internal.Opts{
			NumFormatter: &nf,
		})
		got := calT.LocalizeCurrency(ctx, amount.Negate(), currency.EUR)
		assert.Equal(t, "(€1,234.56)", got)
	})

	t.Run("disambiguate symbol is preserved", func(t *testing.T) {
		nf := currency.EUR.Def().Formatter()
		ctx := internal.WithOptions(context.Background(), &internal.Opts{
			NumFormatter: &nf,
		})
		got := calT.LocalizeCurrency(ctx, amount, currency.USD, currency.WithDisambiguateSymbol())
		assert.Equal(t, "US$1.234,56", got)
	})
}

func TestLocalizeDateTime(t *testing.T) {
	// 2026-03-25T12:34 – matches the co-dian-invoice example that has both
	// issue_date and issue_time, which is the scenario fixed by APP-472.
	dt := cal.MakeDateTime(2026, time.March, 25, 12, 34, 0)

	nf := num.Formatter{}
	tests := []struct {
		name     string
		date     string
		timeF    string
		expected string
	}{
		{
			name:     "ISO defaults",
			date:     "",
			timeF:    "",
			expected: "2026-03-25 · 12:34",
		},
		{
			name:     "custom date format keeps date_format when issue_time present",
			date:     "02/01/2006",
			timeF:    "",
			expected: "25/03/2026 · 12:34",
		},
		{
			name:     "custom date and time format",
			date:     "02/01/2006",
			timeF:    "3:04 PM",
			expected: "25/03/2026 · 12:34 PM",
		},
		{
			name:     "default date with custom time format",
			date:     "",
			timeF:    "3:04 PM",
			expected: "2026-03-25 · 12:34 PM",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cf := internal.CalFormatterISO
			if tc.date != "" {
				cf.Date = tc.date
			}
			if tc.timeF != "" {
				cf.Time = tc.timeF
			}

			opts := &internal.Opts{
				CalFormatter: &cf,
				NumFormatter: &nf,
			}
			ctx := internal.WithOptions(context.Background(), opts)

			got := calT.Localize(ctx, dt)
			assert.Equal(t, tc.expected, got)
		})
	}
}
