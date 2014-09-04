package redisgosearch

import (
	"strings"
	"unicode"
)

// SegmentFn breaks the given string into
// keywords to be indexed.
type SegmentFn func(string) []string

func nonWordOrNumbers(w rune) (r bool) {
	r = !unicode.IsLetter(w) && !unicode.IsDigit(w)
	return
}

// DefaultSegment splits strings at any non-letter, non-digit char.
func DefaultSegment(p string) (r []string) {
	p = strings.ToLower(p)
	r1 := strings.Fields(p)
	for _, word := range r1 {
		deepSplitWords := strings.FieldsFunc(word, nonWordOrNumbers)
		if len(deepSplitWords) >= 1 {
			for _, w := range deepSplitWords {
				r = append(r, w)
			}
		}
	}
	return
}
