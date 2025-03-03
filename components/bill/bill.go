// Package bill provides templates for rendering objects from GOBL's `bill` package.
package bill

import "github.com/invopop/gobl/num"

func amountOrZero(amount *num.Amount) num.Amount {
	if amount == nil {
		return num.AmountZero
	}
	return *amount
}
