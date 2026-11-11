package display

// Vars are the variables a snippet can refer to, as in {{ .PaymentLink }}.
// Every value is already formatted the way the document shows it, with the
// locale, date format and number format of the render, and is escaped for
// wherever it lands in the snippet: as text, inside an attribute, or as a
// URL in an href. A variable that does not apply to the document, such as
// the due date of an invoice without payment terms, is an empty string, so
// {{ if .DueDate }} guards a block that needs it.
//
// Only an invoice fills DueDate, TotalPayable, TotalDue and PaymentLink;
// orders, deliveries and payment documents leave them empty.
type Vars struct {
	// Code is the document's code, such as "0042".
	Code string
	// Series is the document's series, such as "SAMPLE", or empty.
	Series string
	// Number is the series and code joined the way the document's title
	// shows them, such as "SAMPLE-0042", or just the code without a series.
	Number string
	// IssueDate is the formatted issue date.
	IssueDate string
	// DueDate is the formatted first due date in the invoice's payment
	// terms, or empty.
	DueDate string
	// TotalPayable is the formatted total payable with its currency, after
	// taxes and before advances.
	TotalPayable string
	// TotalDue is the formatted amount still to pay with its currency: the
	// total payable less any advances. It equals TotalPayable when the
	// invoice records no advances.
	TotalDue string
	// Currency is the document's currency code, such as "EUR".
	Currency string
	// SupplierName is the supplier's name.
	SupplierName string
	// CustomerName is the customer's name, or empty on a simplified invoice
	// without one.
	CustomerName string
	// PaymentLink is the first online payment URL in the payment
	// instructions, or empty.
	PaymentLink string
}
