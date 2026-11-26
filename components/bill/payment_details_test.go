package bill

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/pay"
	"github.com/stretchr/testify/assert"
)

const (
	paypalLabel = "PayPal"
	paypalURL   = "https://www.paypal.com/pay"
)

func TestPaymentOnlineQRCodes(t *testing.T) {
	online := func() *pay.Instructions {
		return &pay.Instructions{
			Key: pay.MeansKeyOnline,
			Online: []*pay.Online{
				{Label: "Pay by card", URL: "https://pay.example.com/card"},
				{Label: paypalLabel, URL: paypalURL},
			},
		}
	}
	payable := func(cents int64) *bill.Totals {
		return &bill.Totals{Payable: num.MakeAmount(cents, 2)}
	}
	due := func(cents int64) *bill.Totals {
		d := num.MakeAmount(cents, 2)
		return &bill.Totals{Payable: num.MakeAmount(10000, 2), Due: &d}
	}

	t.Run("one QR per online link when payment is pending", func(t *testing.T) {
		qrs := paymentOnlineQRCodes(context.Background(), online(), payable(10000))
		if assert.Len(t, qrs, 2) {
			assert.Equal(t, "Pay by card", qrs[0].online.Label)
			assert.Equal(t, paypalLabel, qrs[1].online.Label)
			assert.Contains(t, qrs[0].src, "data:image/svg+xml;base64,")
		}
	})

	t.Run("shown for any payment means with online links", func(t *testing.T) {
		inst := online()
		inst.Key = pay.MeansKeyCreditTransfer
		assert.Len(t, paymentOnlineQRCodes(context.Background(), inst, payable(10000)), 2)
	})

	t.Run("uses the due amount after advances", func(t *testing.T) {
		assert.Len(t, paymentOnlineQRCodes(context.Background(), online(), due(4000)), 2)
	})

	t.Run("skips links that aren't web URLs", func(t *testing.T) {
		inst := online()
		inst.Online[0].URL = "javascript:alert(1)"
		inst.Online = append(inst.Online, nil, &pay.Online{Label: "Empty"})
		qrs := paymentOnlineQRCodes(context.Background(), inst, payable(10000))
		if assert.Len(t, qrs, 1) {
			assert.Equal(t, paypalLabel, qrs[0].online.Label)
		}
	})

	t.Run("uses the default size without qr-size meta", func(t *testing.T) {
		qrs := paymentOnlineQRCodes(context.Background(), online(), payable(10000))
		if assert.Len(t, qrs, 2) {
			assert.Equal(t, float64(PaymentOnlineQRDefaultSize), qrs[0].size)
			assert.Equal(t, templ.SafeCSS("width: 20mm; height: 20mm;"), qrs[0].style())
		}
	})

	t.Run("skips links with a qr-size of zero", func(t *testing.T) {
		inst := online()
		inst.Online[0].Meta = cbc.Meta{PaymentOnlineMetaKeyQRSize: "0"}
		qrs := paymentOnlineQRCodes(context.Background(), inst, payable(10000))
		if assert.Len(t, qrs, 1) {
			assert.Equal(t, paypalLabel, qrs[0].online.Label)
		}
		assert.Equal(t, []*pay.Online{inst.Online[0]}, paymentOnlineLinks(inst, qrs))
	})

	t.Run("sizes the QR code from qr-size meta", func(t *testing.T) {
		inst := online()
		inst.Online[0].Meta = cbc.Meta{PaymentOnlineMetaKeyQRSize: "32.5"}
		qrs := paymentOnlineQRCodes(context.Background(), inst, payable(10000))
		if assert.Len(t, qrs, 2) {
			assert.Equal(t, 32.5, qrs[0].size)
			assert.Equal(t, templ.SafeCSS("width: 32.5mm; height: 32.5mm;"), qrs[0].style())
			assert.Equal(t, float64(PaymentOnlineQRDefaultSize), qrs[1].size)
		}
	})

	t.Run("none without online links", func(t *testing.T) {
		inst := online()
		inst.Online = nil
		assert.Empty(t, paymentOnlineQRCodes(context.Background(), inst, payable(10000)))
	})

	t.Run("none without instructions", func(t *testing.T) {
		assert.Empty(t, paymentOnlineQRCodes(context.Background(), nil, payable(10000)))
	})

	t.Run("none when already paid", func(t *testing.T) {
		assert.Empty(t, paymentOnlineQRCodes(context.Background(), online(), due(0)))
	})

	t.Run("none when the total due is zero or negative", func(t *testing.T) {
		assert.Empty(t, paymentOnlineQRCodes(context.Background(), online(), payable(0)))
		assert.Empty(t, paymentOnlineQRCodes(context.Background(), online(), payable(-5000)))
	})

	t.Run("none without totals", func(t *testing.T) {
		assert.Empty(t, paymentOnlineQRCodes(context.Background(), online(), nil))
	})

	t.Run("none whenever the invoice is shown as paid", func(t *testing.T) {
		inv := &bill.Invoice{Totals: due(0)}
		assert.True(t, paid(internal.DocumentFor(inv)))
		assert.Empty(t, paymentOnlineQRCodes(context.Background(), online(), inv.Totals))
	})

	t.Run("nothing to pay is not shown as paid", func(t *testing.T) {
		for _, totals := range []*bill.Totals{payable(0), payable(-5000), due(-1000)} {
			inv := &bill.Invoice{Totals: totals}
			assert.False(t, paid(internal.DocumentFor(inv)))
			assert.Empty(t, paymentOnlineQRCodes(context.Background(), online(), inv.Totals))
		}
	})
}

func TestPaymentOnlineQRSize(t *testing.T) {
	tests := []struct {
		name string
		meta cbc.Meta
		want float64
	}{
		{"no meta", nil, PaymentOnlineQRDefaultSize},
		{"other meta", cbc.Meta{"foo": "bar"}, PaymentOnlineQRDefaultSize},
		{"empty value", cbc.Meta{PaymentOnlineMetaKeyQRSize: ""}, PaymentOnlineQRDefaultSize},
		{"not a number", cbc.Meta{PaymentOnlineMetaKeyQRSize: "big"}, PaymentOnlineQRDefaultSize},
		{"with units", cbc.Meta{PaymentOnlineMetaKeyQRSize: "25mm"}, PaymentOnlineQRDefaultSize},
		{"not a number (NaN)", cbc.Meta{PaymentOnlineMetaKeyQRSize: "NaN"}, PaymentOnlineQRDefaultSize},
		{"zero hides", cbc.Meta{PaymentOnlineMetaKeyQRSize: "0"}, 0},
		{"negative hides", cbc.Meta{PaymentOnlineMetaKeyQRSize: "-5"}, 0},
		{"whitespace around", cbc.Meta{PaymentOnlineMetaKeyQRSize: " 25 "}, 25},
		{"within range", cbc.Meta{PaymentOnlineMetaKeyQRSize: "30"}, 30},
		{"decimal", cbc.Meta{PaymentOnlineMetaKeyQRSize: "12.5"}, 12.5},
		{"too small is raised to the minimum", cbc.Meta{PaymentOnlineMetaKeyQRSize: "3"}, PaymentOnlineQRMinSize},
		{"too large is capped at the maximum", cbc.Meta{PaymentOnlineMetaKeyQRSize: "200"}, PaymentOnlineQRMaxSize},
		{"infinite is capped at the maximum", cbc.Meta{PaymentOnlineMetaKeyQRSize: "Inf"}, PaymentOnlineQRMaxSize},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ol := &pay.Online{URL: paypalURL, Meta: tt.meta}
			assert.Equal(t, tt.want, paymentOnlineQRSize(ol))
		})
	}
}

func TestPaymentOnlineQRURL(t *testing.T) {
	assert.True(t, paymentOnlineQRURL("https://pay.example.com/i/1?method=card"))
	assert.True(t, paymentOnlineQRURL("http://pay.example.com/i/1"))
	assert.True(t, paymentOnlineQRURL("HTTPS://pay.example.com/i/1"), "schemes are case-insensitive")
	assert.True(t, paymentOnlineQRURL("Http://pay.example.com/i/1"), "schemes are case-insensitive")
	assert.False(t, paymentOnlineQRURL(""))
	assert.False(t, paymentOnlineQRURL("javascript:alert(1)"))
	assert.False(t, paymentOnlineQRURL("data:text/html,hello"))
	assert.False(t, paymentOnlineQRURL("mailto:billing@example.com"))
	assert.False(t, paymentOnlineQRURL("/relative/path"))
	assert.False(t, paymentOnlineQRURL("https://"))
	assert.False(t, paymentOnlineQRURL("https://:443/pay"), "port without a hostname")
	assert.False(t, paymentOnlineQRURL("https://user@/pay"), "user info without a hostname")
	assert.False(t, paymentOnlineQRURL("https://pay example.com/"), "invalid hostname")
}

func TestPaymentOnlineName(t *testing.T) {
	assert.Equal(t, paypalLabel, paymentOnlineName(&pay.Online{Label: paypalLabel, URL: paypalURL}))
	assert.Equal(t, "www.paypal.com", paymentOnlineName(&pay.Online{URL: paypalURL + "?ref=" + strings.Repeat("x", 200)}))
	assert.Equal(t, "not a url", paymentOnlineName(&pay.Online{URL: "not a url"}))
}

func TestPaymentOnlineLinks(t *testing.T) {
	card := &pay.Online{Label: "Pay by card", URL: "https://pay.example.com/card"}
	paypal := &pay.Online{Label: paypalLabel, URL: paypalURL}
	other := &pay.Online{Label: "Bank app", URL: "bank://pay/123"}
	inst := &pay.Instructions{Online: []*pay.Online{card, nil, paypal, other}}
	pending := &bill.Totals{Payable: num.MakeAmount(10000, 2)}

	t.Run("leaves out links shown with a QR code", func(t *testing.T) {
		qrs := paymentOnlineQRCodes(context.Background(), inst, pending)
		assert.Len(t, qrs, 2)
		assert.Equal(t, []*pay.Online{other}, paymentOnlineLinks(inst, qrs))
	})

	t.Run("lists every link without QR codes", func(t *testing.T) {
		assert.Equal(t, []*pay.Online{card, paypal, other}, paymentOnlineLinks(inst, nil))
	})
}
