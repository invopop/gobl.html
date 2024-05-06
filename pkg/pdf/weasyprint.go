package pdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// weasyprintConvertor connects to the provided URL and waits for a response. We've not seen
// good results with weasyprint, the GOBL HTML convertor will need a lot of work on the
// stylesheets to get the level of results expected.
type weasyprintConvertor struct {
	client *resty.Client
}

func newWeasyprintConvertor(opts ...Config) (*weasyprintConvertor, error) {
	conf := new(config)
	for _, opt := range opts {
		opt(conf)
	}
	if conf.url == "" {
		return nil, errors.New("weasyprint requires url parameter")
	}

	gc := new(weasyprintConvertor)
	gc.client = resty.New().SetBaseURL(conf.url)

	fmt.Printf("prepared weasyprint convertor with url: %s\n", conf.url)

	return gc, nil
}

func (gc *weasyprintConvertor) HTML(_ context.Context, data []byte, _ ...Option) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	resp, err := gc.client.R().
		SetFileReader("html", "index.html", buf).
		Post("")

	if err != nil {
		fmt.Printf("error sending to weasyprint: %s\n", err.Error())
		return nil, err
	}

	return resp.Body(), nil
}
