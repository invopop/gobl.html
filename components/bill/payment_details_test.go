package bill

import (
	"testing"

	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/pay"
	"github.com/stretchr/testify/assert"
)

const paypalURL = "https://www.paypal.com/pay"

func TestShowPaymentOnlineQRs(t *testing.T) {
	online := func() *pay.Instructions {
		return &pay.Instructions{
			Key: pay.MeansKeyOnline,
			Online: []*pay.Online{
				{Label: "Pay by card", URL: "https://pay.example.com/card"},
				{Label: "PayPal", URL: paypalURL},
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

	t.Run("shown when payment is pending", func(t *testing.T) {
		assert.True(t, showPaymentOnlineQRs(online(), payable(10000)))
	})

	t.Run("accepts combined online keys", func(t *testing.T) {
		inst := online()
		inst.Key = pay.MeansKeyOnline.With(pay.MeansKeyCard)
		assert.True(t, showPaymentOnlineQRs(inst, payable(10000)))
	})

	t.Run("uses the due amount after advances", func(t *testing.T) {
		assert.True(t, showPaymentOnlineQRs(online(), due(4000)))
	})

	t.Run("hidden when already paid", func(t *testing.T) {
		assert.False(t, showPaymentOnlineQRs(online(), due(0)))
	})

	t.Run("hidden when the total due is zero or negative", func(t *testing.T) {
		assert.False(t, showPaymentOnlineQRs(online(), payable(0)))
		assert.False(t, showPaymentOnlineQRs(online(), payable(-5000)))
	})

	t.Run("hidden without totals", func(t *testing.T) {
		assert.False(t, showPaymentOnlineQRs(online(), nil))
	})

	t.Run("hidden for other payment means", func(t *testing.T) {
		inst := online()
		inst.Key = pay.MeansKeyCreditTransfer
		assert.False(t, showPaymentOnlineQRs(inst, payable(10000)))
	})

	t.Run("hidden without instructions", func(t *testing.T) {
		assert.False(t, showPaymentOnlineQRs(nil, payable(10000)))
	})

	t.Run("hidden whenever the invoice is shown as paid", func(t *testing.T) {
		inv := &bill.Invoice{Totals: due(0)}
		assert.True(t, paid(internal.DocumentFor(inv)))
		assert.False(t, showPaymentOnlineQRs(online(), inv.Totals))
	})

	t.Run("nothing to pay is not shown as paid", func(t *testing.T) {
		for _, totals := range []*bill.Totals{payable(0), payable(-5000), due(-1000)} {
			inv := &bill.Invoice{Totals: totals}
			assert.False(t, paid(internal.DocumentFor(inv)))
			assert.False(t, showPaymentOnlineQRs(online(), inv.Totals))
		}
	})
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
	assert.Equal(t, "PayPal", paymentOnlineName(&pay.Online{Label: "PayPal", URL: paypalURL}))
	assert.Equal(t, paypalURL, paymentOnlineName(&pay.Online{URL: paypalURL}))
}
