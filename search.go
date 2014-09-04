package redisgosearch

import (
	"encoding/json"
	"strings"
)

func (client *Client) search(indexType string, keywords string, filters map[string]string, skip int, limit int, segmentFn SegmentFn, result interface{}) (count int, err error) {
	words := DefaultSegment(keywords)
	if len(words) == 0 {
		return
	}
	keywordsKey := client.withnamespace(indexType, "search", strings.Join(words, "+"))
	var args []interface{}
	for _, word := range words {
		args = append(args, client.withnamespace(indexType, "keywords", word))
	}

	if filters != nil {
		for k, v := range filters {
			args = append(args, client.withnamespace(indexType, "filters", k, v))
		}
	}

	args = append([]interface{}{keywordsKey}, args...)

	_, err = client.redisConn.Do("SINTERSTORE", args...)

	if err != nil {
		return
	}

	sortArgs := []interface{}{keywordsKey, "BY", "rank_*", "DESC"}

	rawKeyRs, err := client.redisConn.Do("SORT", sortArgs...)
	if err != nil {
		return
	}

	iKeyRs := rawKeyRs.([]interface{})
	if len(iKeyRs) == 0 {
		return
	}

	count = len(iKeyRs)
	end := skip + limit
	if end > count {
		end = count
	}
	iKeyRs = iKeyRs[skip:end]

	rawRs, err := client.redisConn.Do("MGET", iKeyRs...)
	if err != nil {
		return
	}

	iRs := rawRs.([]interface{})

	var stringRs []string
	for _, row := range iRs {
		if row == nil {
			continue
		}
		stringRs = append(stringRs, string(row.([]byte)))
	}

	jsonData := "[" + strings.Join(stringRs, ", ") + "]"
	err = json.Unmarshal([]byte(jsonData), result)

	return
}

// Search returns the Redis-stored marshalled JSON struct of type indexType, that was originally indexed, filtered by the given parameters.
func (client *Client) Search(indexType string, keywords string, filters map[string]string, skip int, limit int, result interface{}) (count int, err error) {
	return client.search(indexType, keywords, filters, skip, limit, DefaultSegment, result)
}

// SearchCustom does the same as Search, but with a custom keyword segmentation function instead of the default one. See SegmentFn.
func (client *Client) SearchCustom(indexType string, keywords string, filters map[string]string, skip int, limit int, segmentFn SegmentFn, result interface{}) (count int, err error) {
	return client.search(indexType, keywords, filters, skip, limit, segmentFn, result)
}
