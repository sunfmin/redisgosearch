package redisgosearch

import (
	"strings"
	// "fmt"
	"encoding/json"
)

func (this *Client) Search(indexType string, keywords string, limit int, result interface{}) (err error) {
	words := Segment(keywords)
	targetKey := strings.Join(words, "+")
	args := []interface{}{this.withnamespace("keywords", targetKey, indexType)}
	for _, word := range words {
		args = append(args, this.withnamespace("keywords", word, indexType))
	}

	_, err = this.redisConn.Do("SINTERSTORE", args...)
	if err != nil {
		return
	}

	rawKeyRs, err := this.redisConn.Do("SMEMBERS", this.withnamespace("keywords", targetKey, indexType))
	if err != nil {
		return
	}

	iKeyRs := rawKeyRs.([]interface{})
	rawRs, err := this.redisConn.Do("MGET", iKeyRs...)
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
