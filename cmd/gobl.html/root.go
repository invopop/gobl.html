package main

import (
	"io"
	"os"

	"github.com/invopop/gobl.html/internal/gallery"
	"github.com/spf13/cobra"
)

type rootOpts struct {
}

func root() *rootOpts {
	return &rootOpts{}
}

func (o *rootOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "gobl.html",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(versionCmd())
	cmd.AddCommand(convert(o).cmd())
	cmd.AddCommand(serve(o).cmd())
	cmd.AddCommand(&cobra.Command{
		Use:   "generate-index",
		Short: "Write examples/index.html from internal/gallery/catalog.html and examples/*.json",
		RunE: func(*cobra.Command, []string) error {
			return gallery.WriteIndexHTML(gallery.ExamplesDir, gallery.IndexHTMLPath())
		},
	})

	return cmd
}

func (o *rootOpts) outputFilename(args []string) string {
	if len(args) >= 2 && args[1] != "-" {
		return args[1]
	}
	return ""
}

func openInput(cmd *cobra.Command, args []string) (io.ReadCloser, error) {
	if inFile := inputFilename(args); inFile != "" {
		return os.Open(inFile)
	}
	return io.NopCloser(cmd.InOrStdin()), nil
}

func (o *rootOpts) openOutput(cmd *cobra.Command, args []string) (io.WriteCloser, error) {
	if outFile := o.outputFilename(args); outFile != "" {
		flags := os.O_CREATE | os.O_WRONLY
		return os.OpenFile(outFile, flags, os.ModePerm)
	}
	return writeCloser{cmd.OutOrStdout()}, nil
}

type writeCloser struct {
	io.Writer
}

func (writeCloser) Close() error { return nil }
