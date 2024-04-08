package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/invopop/ctxi18n"
	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/invopop/gobl.html/pkg/pdf"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

type serveOpts struct {
	*rootOpts
	port string
	pdf  string

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

	return cmd
}

func (s *serveOpts) runE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	var err error
	if s.convertor, err = pdf.New(s.pdf); err != nil {
		return fmt.Errorf("preparing PDF convertor: %w", err)
	}

	e := echo.New()

	e.GET("/:filename", s.generate)

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

func (s *serveOpts) render(c echo.Context, env *gobl.Envelope) ([]byte, error) {
	var err error
	// Set the locale to English to start with
	ctx := c.Request().Context()
	ctx, err = ctxi18n.WithLocale(ctx, "en")
	if err != nil {
		return nil, fmt.Errorf("setting locale: %w", err)
	}

	out, err := goblhtml.Render(ctx, env)
	if err != nil {
		return nil, fmt.Errorf("generating html: %w", err)
	}

	return out, nil
}

func (s *serveOpts) generate(c echo.Context) error {
	filename := c.Param("filename")
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext) + ".json"

	ed, err := os.ReadFile(filepath.Join("./examples", filename))
	if err != nil {
		return fmt.Errorf("loading file: %w", err)
	}

	env := new(gobl.Envelope)
	if err := json.Unmarshal(ed, env); err != nil {
		return fmt.Errorf("unmarshalling file: %w", err)
	}

	data, err := s.render(c, env)
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
