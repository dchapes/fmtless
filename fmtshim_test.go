package fmtless

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSpec(t *testing.T) {
	s, o := getSpec([]byte("%d"))
	assert.True(t, o)
	assert.Equal(t, s, "%d")

	s, o = getSpec([]byte("%df"))
	assert.True(t, o)
	assert.Equal(t, s, "%d")

	s, o = getSpec([]byte("%f"))
	assert.True(t, o)
	assert.Equal(t, s, "%f")

	s, o = getSpec([]byte("%n"))
	assert.False(t, o)
	assert.Equal(t, s, "")

	s, o = getSpec([]byte("ff"))
	assert.False(t, o)
	assert.Equal(t, s, "")

	s, o = getSpec([]byte("%+f"))
	assert.True(t, o)
	assert.Equal(t, s, "%+f")
}

func TestSplitSpecs(t *testing.T) {
	// splitFmtSpecs(fmts string) []sprintMatch {
	specs := splitFmtSpecs("This %s that %d these %f those %g")
	expected := []sprintMatch{
		sprintMatch{"This ", "%s"},
		sprintMatch{" that ", "%d"},
		sprintMatch{" these ", "%f"},
		sprintMatch{" those ", "%g"},
	}
	assert.EqualValues(t, expected, specs)
}

func TestSprintf(t *testing.T) {
	filled := Sprintf("This: '%s' is stringier than: %q ", "\"string\"", "\"string\"")
	assert.Equal(t, `This: '"string"' is stringier than: "\"string\"" `, filled)

	filled = Sprintf("This: '%s' is stringier than: %q", "\"string\"", "\"string\"")
	assert.Equal(t, `This: '"string"' is stringier than: "\"string\""`, filled)

	filled = Sprintf("There are %d ways to kill someone who rounds pi to %f", 3, 3.1)
	assert.Equal(t, "There are 3 ways to kill someone who rounds pi to 3.1", filled)
}
