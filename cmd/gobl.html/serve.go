package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/invopop/gobl.html/assets"
	"github.com/invopop/gobl.html/pkg/pdf"
	"github.com/invopop/gobl/org"
	"github.com/labstack/echo/v4"
	echopprof "github.com/sevenNt/echo-pprof"
	"github.com/spf13/cobra"
)

type serveOpts struct {
	*rootOpts
	port   string
	pdf    string
	pdfURL string

	convertor pdf.Convertor
}

func serve(o *rootOpts) *serveOpts {
	return &serveOpts{rootOpts: o}
}

func (s *serveOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve serves GOBL files from the examples directory as HTML",
		RunE:  s.runE,
	}

	f := cmd.Flags()
	f.StringVarP(&s.port, "port", "p", "3000", "port to listen on")
	f.StringVarP(&s.pdf, "pdf", "", "", "PDF convertor to use")
	f.StringVarP(&s.pdfURL, "pdf-url", "", "", "URL of the PDF convertor to use (if needed)")

	return cmd
}

func (s *serveOpts) runE(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	var err error
	opts := make([]pdf.Config, 0)
	if s.pdfURL != "" {
		opts = append(opts, pdf.WithURL(s.pdfURL))
	}
	if s.convertor, err = pdf.New(s.pdf, opts...); err != nil {
		return fmt.Errorf("preparing PDF convertor: %w", err)
	}

	e := echo.New()

	e.StaticFS("/styles", echo.MustSubFS(assets.Content, "styles"))
	e.GET("/:filename", s.generate)

	echopprof.Wrap(e)

	var startErr error
	go func() {
		err := e.Start(":" + s.port)
		if !errors.Is(err, http.ErrServerClosed) {
			startErr = err
		}
		cancel()
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return startErr
}

func (s *serveOpts) render(c echo.Context, req *options, env *gobl.Envelope, opts []goblhtml.Option) ([]byte, error) {
	ctx := c.Request().Context()
	var err error

	// Prepare the request options
	if req.DateFormat != "" {
		opts = append(opts, goblhtml.WithCalFormatter(req.DateFormat, "", time.UTC))
	}
	opts = append(opts, goblhtml.WithLocale(req.Locale))

	// Add some of the extras to the output
	if req.LogoURL != "" {
		logo := &org.Image{
			URL:    req.LogoURL,
			Height: req.LogoHeight,
		}
		opts = append(opts, goblhtml.WithLogo(logo))
	}
	if req.Notes != "" {
		opts = append(opts, goblhtml.WithNotes(req.Notes))
	}

	out, err := goblhtml.Render(ctx, env, opts...)
	if err != nil {
		return nil, fmt.Errorf("generating html: %w", err)
	}

	return out, nil
}

type options struct {
	Filename   string    `param:"filename"`
	Locale     i18n.Code `query:"locale"`
	DateFormat string    `query:"date_format"`
	LogoURL    string    `query:"logo_url"`
	LogoHeight int32     `query:"logo_height"`
	Notes      string    `query:"notes"`
}

func (s *serveOpts) generate(c echo.Context) error {
	req := new(options)
	if err := c.Bind(req); err != nil {
		return fmt.Errorf("binding options: %w", err)
	}
	fn := filepath.Base(path.Clean(req.Filename))
	ext := filepath.Ext(fn)
	fn = strings.TrimSuffix(fn, ext) + ".json"

	ed, err := os.ReadFile(filepath.Join("./examples", fn))
	if err != nil {
		return fmt.Errorf("loading file: %w", err)
	}

	env := new(gobl.Envelope)
	if err := json.Unmarshal(ed, env); err != nil {
		return fmt.Errorf("unmarshalling file: %w", err)
	}

	opts := make([]goblhtml.Option, 0)
	if ext == ".pdf" {
		opts = append(opts, goblhtml.WithEmbeddedStylesheets())
	}
	data, err := s.render(c, req, env, opts)
	if err != nil {
		return err
	}

	switch ext {
	case ".pdf":
		return s.renderPDF(c, ed, data)
	default:
		return c.HTML(http.StatusOK, string(data))
	}
}

func (s *serveOpts) renderPDF(c echo.Context, ed, data []byte) error {
	if s.convertor == nil {
		return errors.New("no PDF convertor available")
	}

	// prepare the GOBL attachment
	opts := []pdf.Option{
		pdf.WithAttachment(&pdf.Attachment{
			Data:     ed,
			Filename: "gobl.json",
		}),
	}

	out, err := s.convertor.HTML(c.Request().Context(), data, opts...)
	if err != nil {
		return fmt.Errorf("converting to PDF: %w", err)
	}

	return c.Blob(http.StatusOK, "application/pdf", out)
}
