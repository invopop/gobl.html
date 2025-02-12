package pdf

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type chromedpConverter struct {
}

const timeout = 30 * time.Second

func newChromedpConverter() (*chromedpConverter, error) {
	cdp := new(chromedpConverter)
	return cdp, nil
}

func (cdp *chromedpConverter) HTML(ctx context.Context, htmlBytes []byte, _ ...Option) ([]byte, error) {
	// create chromedp context
	browserCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// navigate to intial blank page
	// prevents ERROR: received DOM.documentUpdated when there's no top-level frame
	err := chromedp.Run(browserCtx, chromedp.Navigate("about:blank"))
	if err != nil {
		return nil, fmt.Errorf("error starting chromedp browser instance: %w", err)
	}

	dataChan := make(chan []byte, 1)
	errorChan := make(chan error, 1)

	// Setup ListenTarget to execute Print function when page is loaded
	chromedp.ListenTarget(browserCtx, func(ev interface{}) {
		switch ev.(type) {
		case *page.EventLoadEventFired:
			// page loaded here
			// alternatively consider waiting for lifecycle event NetworkIdle
			// note ListenTarget should not block, use a separate goroutine
			go func() {
				err := chromedp.Run(browserCtx, chromedp.ActionFunc(func(ctx context.Context) error {
					buf, _, err := page.PrintToPDF().WithPrintBackground(true).WithPreferCSSPageSize(true).Do(ctx)
					if err != nil {
						return err
					}
					dataChan <- buf
					return nil
				}))

				if err != nil {
					errorChan <- err
				}

			}()
		}
	})

	err = chromedp.Run(browserCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		frameTree, err := page.GetFrameTree().Do(ctx)
		if err != nil {
			return err
		}

		// Note SetDocumentContent does not wait on page load event
		// in contrast to e.g. Navigate
		return page.SetDocumentContent(frameTree.Frame.ID, string(htmlBytes)).Do(ctx)
	}))
	if err != nil {
		return nil, fmt.Errorf("error setting page HTML contents: %w", err)
	}

	select {
	case pdfBytes := <-dataChan:
		return pdfBytes, nil
	case err := <-errorChan:
		return nil, fmt.Errorf("error printing to PDF: %w", err)
	case <-time.After(timeout):
		return nil, fmt.Errorf("error: chromedp operation timed out after %s seconds", timeout)
	}

}
