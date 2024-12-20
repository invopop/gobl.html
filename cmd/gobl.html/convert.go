package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/spf13/cobra"
)

type convertOpts struct {
	*rootOpts
	Zugferd bool
}

func convert(o *rootOpts) *convertOpts {
	return &convertOpts{rootOpts: o}
}

func (c *convertOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert [infile] [outfile]",
		Short: "Convert a GOBL JSON into an HTML file",
		RunE:  c.runE,
	}
	cmd.Flags().BoolVar(&c.Zugferd, "zugferd", false, "Add Zugferd XMP metadata to the PDF")
	return cmd
}

func (c *convertOpts) runE(cmd *cobra.Command, args []string) error {
	input, err := openInput(cmd, args)
	if err != nil {
		return err
	}
	defer input.Close() // nolint:errcheck

	out, err := c.openOutput(cmd, args)
	if err != nil {
		return err
	}
	defer out.Close() // nolint:errcheck

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(input); err != nil {
		return err
	}

	env := new(gobl.Envelope)
	if err := json.Unmarshal(buf.Bytes(), env); err != nil {
		return err
	}

	data, err := goblhtml.Render(cmd.Context(), env)
	if err != nil {
		return fmt.Errorf("generating html: %w", err)
	}

	if _, err = out.Write(data); err != nil {
		return fmt.Errorf("writing html: %w", err)
	}

	return nil
}
