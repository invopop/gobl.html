package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/invopop/gobl.html/internal/catalog"
	"github.com/invopop/gobl.html/internal/previewer"
	"github.com/labstack/echo/v4"
)

const examplesDir = "examples"

// previewerHandler serves the previewer at /: the page, the example catalog,
// the raw example files and the WebAssembly renderer. The .wasm is compiled
// on first request and kept for the life of the process, so restarting the
// server (as air does) picks up changes.
type previewerHandler struct {
	wasmOnce sync.Once
	wasmDir  string
	wasmErr  error

	catalogOnce sync.Once
	entries     []catalog.Entry
}

func newPreviewer() *previewerHandler {
	return &previewerHandler{}
}

func (p *previewerHandler) register(e *echo.Echo) {
	e.GET("/", p.index)
	e.GET("/api/catalog", p.catalog)
	e.GET("/api/examples/:filename", p.example)
	e.GET("/wasm/gobl.html.wasm", p.wasm)
	e.GET("/wasm/wasm_exec.js", p.execJS)
}

func (p *previewerHandler) index(c echo.Context) error {
	return c.HTMLBlob(http.StatusOK, previewer.IndexHTML)
}

// catalog describes the examples once per process; rendering each one to
// detect QR codes takes well under a second for the whole set.
func (p *previewerHandler) catalog(c echo.Context) error {
	p.catalogOnce.Do(func() {
		render := func(ctx context.Context, env *gobl.Envelope) ([]byte, error) {
			return goblhtml.Render(ctx, env)
		}
		entries, errs := catalog.Scan(c.Request().Context(), examplesDir, render)
		for _, err := range errs {
			c.Logger().Warnf("catalog: %v", err)
		}
		p.entries = entries
	})
	return c.JSON(http.StatusOK, p.entries)
}

func (p *previewerHandler) example(c echo.Context) error {
	fn := filepath.Base(c.Param("filename"))
	if !strings.HasSuffix(fn, ".json") {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	return c.File(filepath.Join(examplesDir, fn))
}

// build compiles the renderer to a temporary directory once.
func (p *previewerHandler) build() (string, error) {
	p.wasmOnce.Do(func() {
		dir, err := os.MkdirTemp("", "gobl.html-wasm-")
		if err != nil {
			p.wasmErr = err
			return
		}
		cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", filepath.Join(dir, "gobl.html.wasm"), "./wasm")
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
		if out, err := cmd.CombinedOutput(); err != nil {
			p.wasmErr = fmt.Errorf("building wasm: %w\n%s", err, out)
			return
		}
		p.wasmDir = dir
	})
	return p.wasmDir, p.wasmErr
}

func (p *previewerHandler) wasm(c echo.Context) error {
	dir, err := p.build()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/wasm")
	return c.File(filepath.Join(dir, "gobl.html.wasm"))
}

func (p *previewerHandler) execJS(c echo.Context) error {
	root, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "locating GOROOT: "+err.Error())
	}
	return c.File(filepath.Join(strings.TrimSpace(string(root)), "lib", "wasm", "wasm_exec.js"))
}
