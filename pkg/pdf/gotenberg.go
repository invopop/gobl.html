package pdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	gotenbergHTMLPath = "/forms/chromium/convert/html"
)

type gotenbergConvertor struct {
	client *resty.Client
}

func newGotenbergConvertor(opts ...Config) (*gotenbergConvertor, error) {
	conf := new(config)
	for _, opt := range opts {
		opt(conf)
	}
	if conf.url == "" {
		return nil, errors.New("gotenberg requires url parameter")
	}

	gc := new(gotenbergConvertor)
	gc.client = resty.New().SetBaseURL(conf.url)

	fmt.Printf("prepared gotenberg convertor with url: %s\n", conf.url)

	return gc, nil
}

func (gc *gotenbergConvertor) HTML(_ context.Context, data []byte, _ ...Option) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	resp, err := gc.client.R().
		SetFileReader("files", "index.html", buf).
		Post(gotenbergHTMLPath)

	if err != nil {
		fmt.Printf("error sending to gotenberg: %s\n", err.Error())
		return nil, err
	}

	return resp.Body(), nil
}
