// Package invoice is used to generated invoices
package invoice

import (
	"fmt"
)

func code(series, code string) string {
	if series == "" {
		return code
	}
	return fmt.Sprintf("%s-%s", series, code)
}
