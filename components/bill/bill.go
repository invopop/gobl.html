// Package bill provides templates for rendering objects from GOBL's `bill` package.
package bill

import "github.com/invopop/gobl/num"

// subtotal is a helper to accumulate amounts.
type subtotal struct {
	amount num.Amount
}

// newSubtotal creates a new subtotal.
func newSubtotal() *subtotal {
	return &subtotal{}
}

// Add adds an amount to the subtotal.
func (s *subtotal) Add(a *num.Amount) num.Amount {
	if a == nil {
		return s.amount
	}
	s.amount = a.Add(s.amount)
	return s.amount
}

// Substract substracts an amount from the subtotal.
func (s *subtotal) Substract(a *num.Amount) num.Amount {
	if a == nil {
		return s.amount
	}
	s.amount = a.Subtract(s.amount).Invert()
	return s.amount
}
