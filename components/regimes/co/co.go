// Package co provides additional output for Colombian invoices.
package co

// FormatTaxIDCode formats a Colombian tax ID code.
func FormatTaxIDCode(taxID string) string {
	dv := taxID[len(taxID)-1:]
	nitChars := taxID[:len(taxID)-1]
	nit := ""
	for i, c := range nitChars {
		if i%3 == 0 && i != 0 {
			nit += "."
		}
		nit += string(c)
	}
	return nit + "-" + dv
}
