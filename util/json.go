package util

import (
	"encoding/json"
)

func DecodeFirestore(v interface{}, target interface{}) error {
	marshal, err := json.Marshal(v)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshal, &target)
	if err != nil {
		return err
	}
	return nil
}
