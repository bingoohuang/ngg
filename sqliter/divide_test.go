package sqliter

import (
	"testing"

	"github.com/golang-module/carbon/v2"
	"github.com/stretchr/testify/assert"
)

func TestDivided(t *testing.T) {
	assert.Equal(t, "month.202407", DividedByMonth.CutoffDays(carbon.CreateFromDate(2024, 8, 1).StdTime(), UnitMonth.Of(1)))
	assert.Equal(t, "month.202407", DividedByMonth.CutoffDays(carbon.CreateFromDate(2024, 8, 31).StdTime(), UnitMonth.Of(1)))
	assert.Equal(t, "month.202408", DividedByMonth.CutoffDays(carbon.CreateFromDate(2024, 9, 1).StdTime(), UnitMonth.Of(1)))
}

func TestParseTimeSpan(t *testing.T) {
	span, err := ParseTimeSpan("3 months")
	assert.Nil(t, err)
	assert.Equal(t, TimeSpan{Value: 3, Unit: UnitMonth}, span)
}
