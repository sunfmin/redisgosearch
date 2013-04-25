package redisgosearch

import (
	"strings"
	"unicode"
)

func nonWordOrNumbers(w rune) (r bool) {
	r = !unicode.IsLetter(w) && !unicode.IsDigit(w)
	return
}

func Segment(p string) (r []string) {
	p = strings.ToLower(p)
	r1 := strings.Fields(p)
	for _, word := range r1 {
		// r = append(r, strings.TrimFunc(word, nonWordOrNumbers))

		deepSplitWords := strings.FieldsFunc(word, nonWordOrNumbers)
		if len(deepSplitWords) >= 1 {
			for _, w := range deepSplitWords {
				r = append(r, w)
			}
		}
	}
	return
}
