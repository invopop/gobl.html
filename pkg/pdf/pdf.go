// Package pdf provides a wrapper around different PDF generation tools to provide
// single conversion process regardless of implementation.
package pdf

import (
	"context"
)

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

// New creates a new PDF convertor using the provider.
func New(provider string) (Convertor, error) {
	switch provider {
	case "prince":
		return newPrinceConvertor()
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
