// Package locales provides all the content for translations.
package locales

import (
	"embed"
	"io/fs"
)

//go:embed ar ca da de el en es eu fi fr gl it nl no pl pt

// Content is the embedded content for the locales.
var Content embed.FS

// Codes returns the language codes of every embedded locale, sorted
// alphabetically. Keeping this derived from the embedded content means the
// list of supported languages only has to be maintained in the embed
// directive above.
func Codes() []string {
	entries, err := fs.ReadDir(Content, ".")
	if err != nil {
		// Content is embedded at compile time, so this cannot happen.
		panic(err)
	}
	codes := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			codes = append(codes, e.Name())
		}
	}
	return codes
}

// Config determines the configuration options for numbers and currency
// which don't necessarily coincide with those expected from the locale.
type Config struct {
	DateFormat        string `json:"date_format"`
	DecimalMark       string `json:"decimal_mark"`
	ThousandSeparator string `json:"thousand_separator"`
	CurrencyFormat    string `json:"currency_format"`
}

// NewConfig prepares the default configuration.
func NewConfig() *Config {
	return &Config{
		DateFormat:        "2006-01-02", // Go formatting
		DecimalMark:       ".",
		ThousandSeparator: ",",
		CurrencyFormat:    "%{symbol}%{amount}",
	}
}
