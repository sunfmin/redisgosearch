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

type Piece struct {
	IndexContent string
	Type         string
}

type Indexable interface {
	IndexKey() string
	IndexPieces() []*Piece
	IndexEntity() interface{}
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
	c, err := json.Marshal(i.IndexEntity())
	if err != nil {
		return
	}
	k := i.IndexKey()
	this.redisConn.Do("SET", this.withnamespace("entity", k), c)

	for _, piece := range i.IndexPieces() {
		words := Segment(piece.IndexContent)
		for _, word := range words {
			this.redisConn.Do("SADD", this.withnamespace("keywords", word, piece.Type), this.withnamespace("entity", k))
		}
	}
	return
}

func (this *Client) withnamespace(keys ...string) (r string) {
	keys = append([]string{this.namespace}, keys...)
	r = strings.Join(keys, ":")
	return
}
