package redisgosearch

import (
	"strings"
	// "fmt"
	"encoding/json"
)

func (this *Client) Search(indexType string, keywords string, filters map[string]string, skip int, limit int, result interface{}) (count int, err error) {
	words := Segment(keywords)
	var args []interface{}
	for _, word := range words {
		args = append(args, this.withnamespace(indexType, "keywords", word))
	}

	if filters != nil {
		for k, v := range filters {
			args = append(args, this.withnamespace(indexType, "filters", k, v))
		}
	}

	rawKeyRs, err := this.redisConn.Do("SINTER", args...)
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
