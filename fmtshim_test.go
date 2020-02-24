package fmt

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestGetSpec(t *testing.T) {
	cases := [...]struct {
		spec string
		s    string
		ok   bool
	}{
		{"%d", "%d", true},
		{"%df", "%d", true},
		{"%f", "%f", true},
		{"%n", "", false},
		{"ff", "", false},
		{"%+f", "%+f", true},
	}
	for _, tc := range cases {
		s, ok := getSpec([]byte(tc.spec))
		if ok != tc.ok || s != tc.s {
			t.Errorf("getSpec(%q) gave (%q,%t); want (%q,%t)",
				tc.spec, s, ok, tc.s, tc.ok,
			)
		}
	}
}

func TestSplitSpecs(t *testing.T) {
	cases := [...]struct {
		spec string
		want []sprintMatch
	}{
		{
			"This %s that %d these %f those %g",
			[]sprintMatch{
				{"This ", "%s"},
				{" that ", "%d"},
				{" these ", "%f"},
				{" those ", "%g"},
			},
		},
	}
	for _, tc := range cases {
		specs := splitFmtSpecs(tc.spec)
		if !reflect.DeepEqual(specs, tc.want) {
			t.Errorf("splitFmtSpecs(%q)\n\tgave: %q\n\twant: %q",
				tc.spec, specs, tc.want,
			)
		}
	}
}

func TestSprintf(t *testing.T) {
	cases := [...]struct {
		format string
		args   []interface{}
		want   string
	}{
		{
			format: "This: '%s' is stringier than: %q ",
			args:   []interface{}{`"string"`, `"string"`},
		}, {
			format: "This: '%s' is stringier than: %q",
			args:   []interface{}{`"string"`, `"string"`},
		}, {
			format: "There are %d ways to kill someone who rounds pi to %f",
			args:   []interface{}{3, 3.1},
			want:   "There are 3 ways to kill someone who rounds pi to 3.1",
		}, {
			format: "%U + %U != %U",
			args:   []interface{}{'a', 'í', '쎭'},
		}, {
			format: "%v == %s",
			args:   []interface{}{Errorf("error %v", 1), fmt.Errorf("error %d", 2)},
		}, {
			format: "%X",
			args:   []interface{}{[]byte{1, 2, 3, 4}},
		}, {
			format: "%X",
			args:   []interface{}{[]byte{1, 2, 4, 8, 16, 32, 64, 128, 255}},
		}, {
			format: "%x",
			args:   []interface{}{[]byte{1, 2, 4, 8, 16, 32, 64, 128, 255}},
		}, {
			// Similar to a problematic error in json decode.go
			format: "failed to unmarshal %q into %v",
			args:   []interface{}{"1", reflect.ValueOf("").Type()},
		},
	}
	for _, tc := range cases {
		got := Sprintf(tc.format, tc.args...)
		errgot := Errorf(tc.format, tc.args...)
		want := tc.want
		var errwant error
		if want == "" {
			want = fmt.Sprintf(tc.format, tc.args...)
			errwant = fmt.Errorf(tc.format, tc.args...)
		} else {
			errwant = errors.New(want)
		}
		if got != want {
			t.Errorf("Sprintf(%q, %v)\n\tgave %#q\n\twant %#q",
				tc.format, tc.args, got, want,
			)
		}
		if !reflect.DeepEqual(errgot, errwant) {
			t.Errorf("Errorf(%q, %v)\n\tgave %#q\n\twant %#q",
				tc.format, tc.args, errgot, errwant,
			)
		}

	}
}

func TestSprintCodepoint(t *testing.T) {
	cases := [...]struct {
		r rune
		s string
	}{
		{'\x12', "U+0012"},
		{18, "U+0012"},
		{'í', "U+00ED"},
		{'쎭', "U+C3AD"},
	}
	var sb strings.Builder
	for _, tc := range cases {
		sb.Reset()
		fmtUEscape(&sb, tc.r)
		if s := sb.String(); s != tc.s {
			t.Errorf("fmtUEscape(%q) gave %q, want %q",
				tc.r, s, tc.s,
			)
		}
	}
}
