package redisgosearch

import (
	"strings"
)

func Segment(p string) (r []string) {
	p = strings.ToLower(p)
	r = strings.Split(p, " ")
	return
}
