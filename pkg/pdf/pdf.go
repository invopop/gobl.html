// Package pdf provides a wrapper around different PDF generation tools to provide
// single conversion process regardless of implementation.
package pdf

import (
	"context"
)

// Config defines options used to configure the PDF convertor.
type Config func(any)

type config struct {
	url  string
	auth string
}

// Option defines a functional option to configure the PDF conversion
type Option func(*options)

type options struct {
	metadata    *Metadata
	attachments []*Attachment
}

// Metadata contains additional information to add to the PDF
type Metadata struct {
	Title    string
	Subject  string
	Author   string
	Keywords string
	Creator  string
}

// Attachment is used to embed files inside the PDF when supported by
// the implementation.
type Attachment struct {
	Data        []byte
	Filename    string
	Description string
}

// WithURL sets the URL to use for the connection to a remote server if needed.
func WithURL(url string) Config {
	return func(in any) {
		conf := in.(*config)
		conf.url = url
	}
}

// WithAuthToken defines an authentication token to use to connect to the remote
// server if required by the PDF service.
func WithAuthToken(token string) Config {
	return func(in any) {
		conf := in.(*config)
		conf.auth = token
	}
}

// WithMetadata adds the provided metadata to include in the conversion request.
func WithMetadata(md *Metadata) Option {
	return func(o *options) {
		o.metadata = md
	}
}

// WithAttachment adds the attachment to the conversion request.
func WithAttachment(a *Attachment) Option {
	return func(o *options) {
		o.attachments = append(o.attachments, a)
	}
}

// Convertor defines the interface expected to be able to convert sources into PDF
type Convertor interface {
	// HTML is the default implementation that takes a raw HTML file and converts it
	// into a PDF.
	HTML(ctx context.Context, data []byte, opts ...Option) ([]byte, error)
}

// New creates a new PDF convertor using the provider and any Config options that it might
// require. This method maps the provider to the correct implementation, or returns
// nil if no matching convertor was found.
func New(provider string, conf ...Config) (Convertor, error) {
	switch provider {
	case "prince":
		return newPrinceConvertor()
	case "gotenberg":
		return newGotenbergConvertor(conf...)
	case "weasyprint":
		return newWeasyprintConvertor(conf...)
	default:
		return nil, nil
	}
}

func prepareOptions(opts []Option) *options {
	o := new(options)
	for _, opt := range opts {
		opt(o)
	}
	return o
}
