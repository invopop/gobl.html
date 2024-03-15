package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/invopop/ctxi18n"
	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

type serveOpts struct {
	*rootOpts
	port string
}

func serve(o *rootOpts) *serveOpts {
	return &serveOpts{rootOpts: o}
}

func (c *serveOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve serves GOBL files from the examples directory as HTML",
		RunE:  c.runE,
	}
	return cmd
}

func (c *serveOpts) runE(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile("./examples/invoice.env.json")
	if err != nil {
		return fmt.Errorf("loading file: %w", err)
	}

	env := new(gobl.Envelope)
	if err := json.Unmarshal(data, env); err != nil {
		return fmt.Errorf("unmarshalling file: %w", err)
	}

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		// Set the locale to English to start with
		ctx, err := ctxi18n.WithLocale(c.Request().Context(), "en")
		if err != nil {
			return fmt.Errorf("setting locale: %w", err)
		}
		out, err := goblhtml.Render(ctx, env)
		if err != nil {
			return fmt.Errorf("generating html: %w", err)
		}
		return c.HTML(http.StatusOK, string(out))
	})

	var startErr error
	go func() {
		err = e.Start(":3000")
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
