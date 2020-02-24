package fmt

import "strings"

// SRepeat just repeats a character or string times <repeat>.
// DEPRICATED, just use strings.Repeat.
func SRepeat(char string, repeat int) string {
	return strings.Repeat(char, repeat)
}
