package utils

import (
	"encoding/json"
	"fmt"
)

func StructToMap(i interface{}) (error, map[string]interface{}) {
	j, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
		return err, nil
	}
	var res map[string]interface{}
	err = json.Unmarshal(j, &res)

	if err != nil {
		fmt.Println(err)
		return err, nil
	}

	return nil, res
}

func IsKeyExist(data map[string]interface{}, key string) bool {
	isExist := false
	for k, _ := range data {
		if k == key {
			isExist = true
		}
	}

	return isExist
}
