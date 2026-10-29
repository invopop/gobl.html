package bill

import "github.com/invopop/gobl/pay"

type methodSupport struct {
	dates bool
}

func prepareMethodSupport(methods []*pay.Record) *methodSupport {
	ms := new(methodSupport)
	for _, m := range methods {
		if m != nil && m.Date != nil {
			ms.dates = true
		}
	}
	return ms
}
