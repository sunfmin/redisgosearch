package redisgosearch

import (
	"github.com/garyburd/redigo/redis"
	"encoding/json"
	"strings"
)

type Client struct {
	namespace string
	redisConn redis.Conn
}

type Indexable interface {
	IndexPieces() (pieces []string, relatedPieces []Indexable)
	IndexEntity() (key string, indexType string, entity interface{})
}

func NewClient(address string, namespace string) (r *Client) {
	r = &Client{namespace: namespace}
	var err error
	r.redisConn, err = redis.Dial("tcp", address)
	if err != nil {
		panic(err)
	}
	return
}

func (this *Client) Index(i Indexable) (err error) {
	key, indexType, entity := i.IndexEntity()

	c, err := json.Marshal(entity)
	if err != nil {
		return
	}

	pieces, relatedIndexables := i.IndexPieces()

	this.redisConn.Do("SET", this.withnamespace("entity", key), c)

	for _, piece := range pieces {
		words := Segment(piece)
		for _, word := range words {
			this.redisConn.Do("SADD", this.withnamespace("keywords", word, indexType), this.withnamespace("entity", key))
		}
	}

	if len(relatedIndexables) > 0 {
		for _, i1 := range relatedIndexables {
			this.Index(i1)
		}
	}

	return
}

func (this *Client) withnamespace(keys ...string) (r string) {
	keys = append([]string{this.namespace}, keys...)
	r = strings.Join(keys, ":")
	return
}
