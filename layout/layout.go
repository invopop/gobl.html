// Package layout provides all the possible layouts for a document.
package layout

// Code is an ENUM representing the possible layouts.
type Code string

const (
	// A4 is used for rendering a document with an A4 layout
	A4 Code = "A4"
	// Letter is used for rendering a document with an US letter layout
	Letter Code = "Letter"
	// DIN5008 is used for rendering a document with a german DIN5008 Form-B layout
	DIN5008 Code = "DIN5008"
)

// IsValid checks if a given Layout Code is valid.
func (c Code) IsValid() bool {
	switch c {
	case A4, Letter, DIN5008:
		return true
	default:
		return false
	}
}
