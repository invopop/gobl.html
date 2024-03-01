package organizing

import "github.com/invopop/gobl/org"

func buildAddressLines(addr *org.Address) []string {
	lines := []string{}
	if addr.Street != "" {
		lines = append(lines, buildStreetWithNumbers(addr))
	}
	if addr.StreetExtra != "" {
		lines = append(lines, addr.StreetExtra)
	}
	lines = append(lines, addr.Locality)
	if addr.Region != "" {
		lines = append(lines, addr.Region)
	}
	return lines
}

func buildStreetWithNumbers(addr *org.Address) string {
	str := addr.Street
	if addr.Number != "" {
		str = str + " " + addr.Number
	}
	return str
}
