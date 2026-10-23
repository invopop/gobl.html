package bill

import (
	"testing"

	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/pay"
	"github.com/stretchr/testify/assert"
)

func TestShowPaymentOnlineQRs(t *testing.T) {
	online := func() *pay.Instructions {
		return &pay.Instructions{
			Key: pay.MeansKeyOnline,
			Online: []*pay.Online{
				{Label: "Pay by card", URL: "https://pay.example.com/card"},
				{Label: "PayPal", URL: "https://www.paypal.com/pay"},
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
}
