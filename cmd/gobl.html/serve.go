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
	"strconv"
	"strings"
	"time"

	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/invopop/gobl.html/assets"
	"github.com/invopop/gobl.html/pkg/pdf"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/ziflex/lecho/v3"
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

	e := prepareEcho()
	e.StaticFS("/styles", echo.MustSubFS(assets.Content, "styles"))
	e.StaticFS("/scripts", echo.MustSubFS(assets.Content, "scripts"))
	e.GET("/:filename", s.generate)
	newPreviewer().register(e)

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

func prepareEcho() *echo.Echo {
	e := echo.New()
	zl := log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05",
	})
	b := zl.With().Timestamp()
	log.Logger = b.Logger()
	lvl, _ := lecho.MatchZeroLevel(zerolog.DebugLevel)
	logger := lecho.From(
		log.Logger,
		lecho.WithLevel(lvl),
		lecho.WithTimestamp(),
		// lecho.WithCaller(), // useful for debugging
	)
	e.Logger = logger
	e.Use(lecho.Middleware(lecho.Config{Logger: logger}))
	e.Use(middleware.Recover())
	return e
}

func (s *serveOpts) render(c echo.Context, req *options, env *gobl.Envelope, opts []goblhtml.Option) ([]byte, error) {
	ctx := c.Request().Context()
	opts = append(opts, req.Options()...)
	start := time.Now()
	out, err := goblhtml.Render(ctx, env, opts...)
	if err != nil {
		return nil, fmt.Errorf("generating html: %w", err)
	}
	// Report the native render time, excluding parsing and transfer, so the
	// WASM playground can compare it with the in-browser render.
	ms := float64(time.Since(start).Microseconds()) / 1000
	c.Response().Header().Set("X-Render-Ms", strconv.FormatFloat(ms, 'f', 3, 64))
	c.Response().Header().Set("Server-Timing", "render;dur="+strconv.FormatFloat(ms, 'f', 3, 64))
	return out, nil
}

// options binds the path and query parameters of a render request. The
// rendering options share their names with pdf.go's task configuration.
type options struct {
	goblhtml.Params
	Filename string `param:"filename"`
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
		req.EmbedAssets = true
	}
	applyExampleParams(fn, &req.Params)
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

// applyExampleParams sets the parameters that an example file's name implies,
// such as ".void." or ".adjustment.", unless the request already set them.
func applyExampleParams(fn string, p *goblhtml.Params) {
	if strings.Contains(fn, ".void.") {
		p.Void = true
	}
	if strings.Contains(fn, ".sandbox.") {
		p.Sandbox = true
	}
	if strings.Contains(fn, ".adjustment.") {
		p.AdjustmentMode = true
	}
}
