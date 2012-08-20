package redisgosearch

type Extractor interface {
	Extract(content string)
}
