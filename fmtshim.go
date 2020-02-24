package fmt

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

// Stringer is implemented by any value that has a String method,
// which defines the ``native'' format for that value.
// The String method is used to print values passed as an operand
// to any format that accepts a string or to an unformatted printer
// such as Print.
type Stringer interface {
	String() string
}

// Errorf returns an error object for the given format and values.
func Errorf(format string, a ...interface{}) error {
	return errors.New(Sprintf(format, a...))
}

// Printf prints a format string filled with the given values.
func Printf(format string, a ...interface{}) (n int, err error) {
	out := Sprintf(format, a...)
	print(out)
	return len(out), nil
}

// Println formats using the default formats for its
// operands and writes to standard output. Spaces are
// always added between operands and a newline is appended.
// It returns the number of bytes written and any write
// error encountered.
func Println(a ...interface{}) (n int, err error) {
	out := Sprintln(a...)
	println(out)
	return len(out), nil
}

// Print formats using the default formats for its
// operands and writes to standard output. Spaces
// are added between operands when neither is a string.
// It returns the number of bytes written and any
// write error encountered.
func Print(a ...interface{}) (n int, err error) {
	out := Sprint(a...)
	print(out)
	return len(out), nil
}

// Sprint renders arguments in their default format
// ("%s"/"%d"/"%f"/"%U" for string/int/float/rune, respectively)
func Sprint(a ...interface{}) string {
	return sprint(false, a...)
}
func sprint(ln bool, a ...interface{}) string {
	var sb strings.Builder
	sb.Grow(len(a) * 2) // XXX
	for i, v := range a {
		if ln && i != 0 {
			sb.WriteByte(' ')
		}
		fmtI(&sb, "%s", v)
	}
	if ln {
		sb.WriteByte('\n')
	}
	return sb.String()
}

// Sprintf is a fmtless alternative to fmt.Sprintf that supports some of
// the most common subset of fmt usage.
func Sprintf(fmts string, args ...interface{}) string {
	var sb strings.Builder
	fmlist := splitFmtSpecs(fmts)
	for idx, sm := range fmlist {
		var i interface{}
		i = nil
		if idx < len(args) {
			i = args[idx]
		}
		sm.render(&sb, i)
	}
	return sb.String()
}

// Sprintln is just like Sprint but with a trailing newline.
func Sprintln(a ...interface{}) string {
	return sprint(true, a...)
}

type sprintMatch struct {
	before string
	spec   string
}

func (spm sprintMatch) render(sb *strings.Builder, a interface{}) {
	sb.WriteString(spm.before)
	if spm.spec != "" {
		fmtI(sb, spm.spec, a)
	}
}

func splitFmtSpecs(fmts string) []sprintMatch {
	var (
		lastMatch int
		window    []byte
		matches   []sprintMatch
	)
	percent := byte('%')
	// strings.Index uses a search algo that's
	// probably slower than straight iteration for
	// the length of most format strings.
	for i := 0; i < len(fmts); i++ {
		bound := i + 3
		if bound >= len(fmts) {
			bound = len(fmts)
		}
		window = []byte(fmts)[i:bound]
		if len(window) < 2 || window[0] != percent {
			continue
		}
		spec, ok := getSpec(window)
		if !ok {
			continue
		}
		sm := sprintMatch{before: fmts[lastMatch:i], spec: spec}
		matches = append(matches, sm)
		lastMatch = i + len(sm.spec)
		i += len(sm.spec) - 1
	}
	if lastMatch < len(fmts) {
		lm := sprintMatch{before: fmts[lastMatch:], spec: ""}
		matches = append(matches, lm)
	}
	return matches
}

func getSpec(window []byte) (string, bool) {
	if window[0] != '%' {
		return "", false
	}
	var speclen int
	if window[1] == '+' || window[1] == '#' {
		speclen = 3
	} else {
		speclen = 2
	}
	switch window[speclen-1] {
	case 'v', 's', 'q', 'd', 'b', 'f', 'F', 'g', 'G', 'e', 'E', 'o', 'x', 'X', 'U':
		return string(window[:speclen]), true
	default:
		return "", false
	}
}

func fmtI(sb *strings.Builder, spec string, i interface{}) {
	switch i := i.(type) {
	case Stringer:
		fmtString(sb, spec, i.String())
	case reflect.Type:
		fmtString(sb, spec, i.String())
	case error:
		fmtString(sb, spec, i.Error())
	case string:
		fmtString(sb, spec, i)
	case []byte:
		fmtBytes(sb, spec, i)
	case rune:
		fmtUEscape(sb, i)
	case int, int64:
		fmtInt(sb, spec, i)
	case float32, float64:
		fmtFloat(sb, spec, i)
	default:
		panic("Unsupported interface for fmtless.Sprintf")
	}
}

func fmtBytes(sb *strings.Builder, spec string, b []byte) {
	switch spec {
	case "%v", "%s":
		sb.Write(b)
	case "%q":
		fmtString(sb, spec, string(b))
	case "%x", "%X":
		{
			tmp := make([]byte, 0, 2)
			for _, c := range b {
				tmp = strconv.AppendInt(tmp[:0], int64(c), 16)
				if len(tmp) == 1 {
					sb.WriteByte('0')
				}
				if spec == "%X" {
					for i, r := range tmp {
						if 'a' <= r && r <= 'z' {
							tmp[i] -= 'a' - 'A'
						}
					}
				}
				sb.Write(tmp)
			}
		}
	default:
		panic("Unsupported spec for []byte: " + spec)
	}
}

func fmtString(sb *strings.Builder, spec, s string) {
	switch spec {
	case "%s", "%v", "%#v":
		sb.WriteString(s)
	case "%q":
		sb.WriteString(strconv.Quote(s))
	default:
		panic("Unsupported spec for string: " + spec)
	}
}

func fmtUEscape(sb *strings.Builder, r rune) {
	tmp := make([]byte, 0, 8)
	tmp = strconv.AppendInt(tmp, int64(r), 16)
	for i, r := range tmp {
		if 'a' <= r && r <= 'z' {
			tmp[i] -= 'a' - 'A'
		}
	}
	sb.Grow(6)
	sb.WriteString("U+")
	if len(tmp) < 4 {
		sb.WriteString("0000"[:4-len(tmp)])
	}
	sb.Write(tmp)
}

func fmtInt(sb *strings.Builder, spec string, i interface{}) {
	var i64 int64
	var base int
	switch i := i.(type) {
	case int:
		i64 = int64(i)
	case int32:
		i64 = int64(i)
	case int64:
		i64 = i
	}
	switch spec {
	case "%s", "%d", "%v":
		base = 10
	case "%o":
		base = 8
	case "%b":
		base = 2
	case "%X", "%x":
		base = 16
	}
	tmp := make([]byte, 0, 64)
	tmp = strconv.AppendInt(tmp, i64, base)
	if spec == "%X" {
		for i, r := range tmp {
			if 'a' <= r && r <= 'z' {
				tmp[i] -= 'a' - 'A'
			}
		}
	}
	sb.Write(tmp)
}

func fmtFloat(sb *strings.Builder, spec string, f interface{}) {
	var bitSize int
	var f64 float64
	switch f := f.(type) {
	case float32:
		f64 = float64(f)
		bitSize = 32
	case float64:
		f64 = f
		bitSize = 64
	default:
		panic("unreachable")
	}
	tmp := make([]byte, 0, 32)
	switch spec {
	case "%b", "%f", "%F", "%g", "%G", "%e", "%E":
		tmp = strconv.AppendFloat(tmp, f64, spec[1], -1, bitSize)
	case "%s", "%v":
		tmp = strconv.AppendFloat(tmp, f64, 'f', -1, bitSize)
	default:
		panic("Unsupported specifier for floats: " + spec)
	}
	sb.Write(tmp)
}
