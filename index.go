package redisgosearch

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"strings"
)

type Client struct {
	namespace string
	redisConn redis.Conn
}

type Indexable interface {
	IndexPieces() (pieces []string, relatedPieces []Indexable)
	IndexEntity() (indexType string, key string, entity interface{}, rank int64)
	IndexFilters() (r map[string]string)
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
	indexType, key, entity, rank := i.IndexEntity()

	c, err := json.Marshal(entity)
	if err != nil {
		return
	}

	pieces, relatedIndexables := i.IndexPieces()

	entityKey := this.withnamespace(indexType, "entity", key)
	this.redisConn.Do("SET", entityKey, c)
	this.redisConn.Do("SET", "rank_"+entityKey, rank)

	filters := i.IndexFilters()

	for k, v := range filters {
		this.redisConn.Do("SADD", this.withnamespace(indexType, "filters", k, v), entityKey)
	}

	for _, piece := range pieces {
		words := Segment(piece)
		for _, word := range words {
			this.redisConn.Do("SADD", this.withnamespace(indexType, "keywords", word), entityKey)
		}
	}

	if len(relatedIndexables) > 0 {
		for _, i1 := range relatedIndexables {
			this.Index(i1)
		}
	}

	return
}

func (this *Client) RemoveIndex(i Indexable) (err error) {
	indexType, key, entity, rank := i.IndexEntity()

	c, err := json.Marshal(entity)
	if err != nil {
		return
	}

	pieces, relatedIndexables := i.IndexPieces()

	entityKey := this.withnamespace(indexType, "entity", key)
	this.redisConn.Do("DEL", entityKey, c)
	this.redisConn.Do("DEL", "rank_"+entityKey, rank)

	filters := i.IndexFilters()

	for k, v := range filters {
		this.redisConn.Do("SREM", this.withnamespace(indexType, "filters", k, v), entityKey)
	}

	for _, piece := range pieces {
		words := Segment(piece)
		for _, word := range words {
			this.redisConn.Do("SREM", this.withnamespace(indexType, "keywords", word), entityKey)
		}
	}

	if len(relatedIndexables) > 0 {
		for _, i1 := range relatedIndexables {
			this.RemoveIndex(i1)
		}
	}

	return
}

func (this *Client) withnamespace(keys ...string) (r string) {
	keys = append([]string{this.namespace}, keys...)
	r = strings.Join(keys, ":")
	return
}
