package bill

import (
	"testing"

	gbill "github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
)

func TestPrepareLineSupportUNTDIDUnit(t *testing.T) {
	lines := []*gbill.Line{
		{
			Item: &org.Item{
				Ext: tax.MakeExtensions().Set("untdid-unit", "E48"),
			},
		},
	}

	assert.True(t, prepareLineSupport(new(tax.RegimeDef), lines, nil, nil).units)
}
