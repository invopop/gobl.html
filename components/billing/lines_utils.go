package billing

// flatLine represents a single GOBL Bill Line in a flat-ish format
// prepared to easily generate HTML.
type flatLine struct {
	Quantity string
}

/*
func flattern_lines(inv *bill.Invoice) [][]*flatLine {
	var lines [][]string
	for _, l := range inv.Lines {
		lines = append(lines, []string{l.Quantity.String()})
	}
	return nil
}
*/
