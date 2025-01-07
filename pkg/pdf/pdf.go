// Package pdf provides a wrapper around different PDF generation tools to provide
// single conversion process regardless of implementation.
package pdf

import (
	"context"
	_ "embed"
	"fmt"
	"io/fs"
	"path/filepath"
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
	styles      []*Stylesheet
	attachments []*Attachment
	xmpMetadata *XMPMetadata
}

// Metadata contains additional information to add to the PDF
type Metadata struct {
	Title    string
	Subject  string
	Author   string
	Keywords string
	Creator  string
}

// Stylesheet descriptions a document to upload with the HTML for styles.
type Stylesheet struct {
	Data     []byte
	Filename string
}

// Attachment is used to embed files inside the PDF when supported by
// the implementation.
type Attachment struct {
	Data        []byte
	Filename    string
	Description string
}

// XMPMetadata is used to add XMP metadata to the PDF.
type XMPMetadata struct {
	Data     []byte
	Filename string
}

//go:embed assets/zugferd.xmp
var zugferdXMPData []byte

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

// WithStylesheets prepares the stylesheets to be included in the PDF generation
// request.
func WithStylesheets(src fs.FS) Option {
	return func(o *options) {
		err := fs.WalkDir(src, "styles", func(path string, _ fs.DirEntry, _ error) error {
			if filepath.Ext(path) != ".css" {
				return nil
			}
			data, err := fs.ReadFile(src, path)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}
			o.styles = append(o.styles, &Stylesheet{
				Filename: path,
				Data:     data,
			})
			return nil
		})
		if err != nil {
			panic(err)
		}
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

// WithZugferd adds the Zugferd XMP metadata to the conversion request.
func WithZugferd() Option {
	return func(o *options) {
		o.xmpMetadata = &XMPMetadata{
			Data:     loadXMP(),
			Filename: "zugferd.xmp",
		}
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

func loadXMP() []byte {
	fmt.Printf("loading XMP metadata: %d bytes\n", len(zugferdXMPData))
	return zugferdXMPData
}
