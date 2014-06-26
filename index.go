package redisgosearch

import (
	"encoding/json"
	"strings"

	"github.com/garyburd/redigo/redis"
)

// Client wraps a namespace (Redis-key prefix) and internal connection.
type Client struct {
	namespace string
	redisConn redis.Conn
}

// Indexable is satisfied by any struct that can be indexed
// and searched in Redis by this package.
type Indexable interface {
	IndexPieces() (pieces []string, relatedPieces []Indexable)
	IndexEntity() (indexType string, key string, entity interface{}, rank int64)
	IndexFilters() (r map[string]string)
}

// NewClient returns a Client given the redis address and namespace,
// or an error if a connection couldn't be made.
func NewClient(address string, namespace string) (r *Client, err error) {
	r = &Client{namespace: namespace}
	r.redisConn, err = redis.Dial("tcp", address)
	return
}

func (client *Client) index(i Indexable, segmentFn SegmentFn) (err error) {
	indexType, key, entity, rank := i.IndexEntity()

	c, err := json.Marshal(entity)
	if err != nil {
		return
	}

	pieces, relatedIndexables := i.IndexPieces()

	entityKey := client.withnamespace(indexType, "entity", key)
	client.redisConn.Do("SET", entityKey, c)
	client.redisConn.Do("SET", "rank_"+entityKey, rank)

	filters := i.IndexFilters()

	for k, v := range filters {
		client.redisConn.Do("SADD", client.withnamespace(indexType, "filters", k, v), entityKey)
	}

	for _, piece := range pieces {
		words := segmentFn(piece)
		for _, word := range words {
			client.redisConn.Do("SADD", client.withnamespace(indexType, "keywords", word), entityKey)
		}
	}

	if len(relatedIndexables) > 0 {
		for _, i1 := range relatedIndexables {
			client.Index(i1)
		}
	}

	return
}

// Index marshals the given Indexable and stores
// it in the Redis database, using the default keyword segmentation function.
func (client *Client) Index(i Indexable) (err error) {
	return client.index(i, DefaultSegment)
}

// IndexCustom does the same as Index, with a custom keyword segmentation function.
func (client *Client) IndexCustom(i Indexable, segmentFn SegmentFn) (err error) {
	return client.index(i, segmentFn)
}

func (client *Client) removeIndex(i Indexable, segmentFn SegmentFn) (err error) {
	indexType, key, entity, rank := i.IndexEntity()

	c, err := json.Marshal(entity)
	if err != nil {
		return
	}

	pieces, relatedIndexables := i.IndexPieces()

	entityKey := client.withnamespace(indexType, "entity", key)
	client.redisConn.Do("DEL", entityKey, c)
	client.redisConn.Do("DEL", "rank_"+entityKey, rank)

	filters := i.IndexFilters()

	for k, v := range filters {
		client.redisConn.Do("SREM", client.withnamespace(indexType, "filters", k, v), entityKey)
	}

	for _, piece := range pieces {
		words := segmentFn(piece)
		for _, word := range words {
			client.redisConn.Do("SREM", client.withnamespace(indexType, "keywords", word), entityKey)
		}
	}

	if len(relatedIndexables) > 0 {
		for _, i1 := range relatedIndexables {
			client.RemoveIndex(i1)
		}
	}

	return
}

// RemoveIndex deletes the Redis keys and data for the given
// Indexable (the opposite of Index)
func (client *Client) RemoveIndex(i Indexable) (err error) {
	return client.removeIndex(i, DefaultSegment)
}

// RemoveIndexCustom does the same as RemoveIndex, with a custom keyword segmentation function.
func (client *Client) RemoveIndexCustom(i Indexable, segmentFn SegmentFn) (err error) {
	return client.removeIndex(i, segmentFn)
}

func (client *Client) withnamespace(keys ...string) (r string) {
	keys = append([]string{client.namespace}, keys...)
	r = strings.Join(keys, ":")
	return
}
