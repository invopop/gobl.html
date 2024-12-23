package pdf

import (
	"context"

	"github.com/invopop/princepdf"
)

// The princeConvertor is by far the best option for converting HTML to PDF for GOBL documents.
// Response times are fast and the quality of the PDFs is excellent.
type princeConvertor struct {
	client *princepdf.Client
}

func newPrinceConvertor() (*princeConvertor, error) {
	c := &princeConvertor{client: princepdf.New()}

	if err := c.client.Start(); err != nil {
		return nil, err
	}

	return c, nil
}

func (pc *princeConvertor) HTML(_ context.Context, data []byte, opts ...Option) ([]byte, error) {
	o := prepareOptions(opts)
	j := new(princepdf.Job)
	j.Input = &princepdf.Input{
		Src: "data.html",
	}
	j.Files = map[string][]byte{
		"data.html": data,
	}

	if len(o.styles) > 0 {
		for _, ss := range o.styles {
			j.Input.Styles = append(j.Input.Styles, ss.Filename)
			j.Files[ss.Filename] = ss.Data
		}
	}
	if len(o.attachments) > 0 {
		for _, a := range o.attachments {
			j.Files[a.Filename] = a.Data
			if j.PDF == nil {
				j.PDF = new(princepdf.PDF)
			}
			j.PDF.Attach = append(j.PDF.Attach, &princepdf.Attachment{
				URL:         a.Filename,
				Filename:    a.Filename,
				Description: a.Description,
			})
		}
	}

	if o.metadata != nil {
		j.Metadata = &princepdf.Metadata{
			Title:    o.metadata.Title,
			Subject:  o.metadata.Subject,
			Author:   o.metadata.Author,
			Keywords: o.metadata.Keywords,
			Creator:  o.metadata.Creator,
		}
	}

	if o.xmpMetadata != nil {
		xmpFilename := "metadata.xmp"
		j.Files[xmpFilename] = o.xmpMetadata.Data
		if j.PDF == nil {
			j.PDF = new(princepdf.PDF)
		}
		j.PDF.PDFXMP = xmpFilename
		// j.PDF.Attach = append(j.PDF.Attach, &princepdf.Attachment{
		// 	URL:         xmpFilename,
		// 	Filename:    xmpFilename,
		// 	Description: "XMP Metadata File",
		// })
	}

	return pc.client.Run(j)
}
