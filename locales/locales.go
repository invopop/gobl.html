// Package locales provides all the content for translations.
package locales

import "embed"

//go:embed de en es fr gr it pl pt

// Content is the embedded content for the locales.
var Content embed.FS

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
