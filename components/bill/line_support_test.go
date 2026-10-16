package bill

import (
	"testing"

	gbill "github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
)

func TestPrepareLineSupportUnits(t *testing.T) {
	reg := new(tax.RegimeDef)

	t.Run("no unit", func(t *testing.T) {
		lines := []*gbill.Line{{Item: &org.Item{}}}
		assert.False(t, prepareLineSupport(reg, lines, nil, nil).units)
	})

	t.Run("GOBL unit", func(t *testing.T) {
		lines := []*gbill.Line{{Item: &org.Item{Unit: org.UnitHour}}}
		assert.True(t, prepareLineSupport(reg, lines, nil, nil).units)
	})

	t.Run("UNTDID unit extension only", func(t *testing.T) {
		lines := []*gbill.Line{{
			Item: &org.Item{Ext: tax.MakeExtensions().Set("untdid-unit", "E48")},
		}}
		assert.True(t, prepareLineSupport(reg, lines, nil, nil).units)
	})
}
