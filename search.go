package redisgosearch

import (
	"strings"
	"fmt"
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

	rawRs, err := this.redisConn.Do("SMEMBERS", this.withnamespace("keywords", targetKey, indexType))
	r.jsonData = "["
	var stringRs []string
	iRs := rawRs.([]interface{})
	for _, row := range iRs {
		stringRs = append(stringRs, string(row.([]byte)))
	}
	fmt.Printf("%+v", stringRs)
	return
}
