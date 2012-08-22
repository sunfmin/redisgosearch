package redisgosearch

import (
	"strings"
	// "fmt"
	"encoding/json"
)

type Result struct {
}

type TypeList struct {
	jsonData string
}

func (this Result) Type(t string) (r TypeList) {
	return
}

func (this TypeList) All(v interface{}) (err error) {
	err = json.Unmarshal([]byte(this.jsonData), v)
	return
}

func (this *Client) SearchInType(keywords string, indexType string) (r TypeList, err error) {
	words := Segment(keywords)
	targetKey := strings.Join(words, "+")
	args := []interface{}{this.withnamespace("keywords", targetKey, indexType)}
	for _, word := range words {
		args = append(args, this.withnamespace("keywords", word, indexType))
	}

	// fmt.Printf("%+v", args)

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

	r.jsonData = "[" + strings.Join(stringRs, ", ") + "]"
	return
}
