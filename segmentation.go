package redisgosearch

import (
	"strings"
	"regexp"
)

var tailNonWordRegexp = regexp.MustCompile(`(.*)[^\w]+$`)

func Segment(p string) (r []string) {
	p = strings.ToLower(p)
	r1 := strings.Split(p, " ")
	for _, word := range r1 {
		word = tailNonWordRegexp.ReplaceAllString(word, "$1")
		r = append(r, word)
	}
	return
}
