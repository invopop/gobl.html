package layout

// Code is an ENUM representing the possible layouts.
type Code string

const (
	A4      Code = "A4"
	Letter  Code = "Letter"
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
