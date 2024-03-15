// Package billing provides components for building "bill" related
// documents, like Invoices.
package billing

import (
	"fmt"
)

func invoiceCode(series, code string) string {
	if series == "" {
		return code
	}
	return fmt.Sprintf("%s-%s", series, code)
}
