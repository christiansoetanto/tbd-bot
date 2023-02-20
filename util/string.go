package util

import (
	"context"
	"encoding/json"
	"regexp"
)

const alphanumOnlyPattern = "[^a-zA-Z0-9]+"
const alphanumAndSpaceOnlyPattern = "[^a-zA-Z0-9 ]+"

func replace(input, pattern string) string {
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	return reg.ReplaceAllString(input, "")
}

func ToAlphanum(ctx context.Context, input string) string {
	return replace(input, alphanumOnlyPattern)
}
func ToAlphanumAndSpace(ctx context.Context, input string) string {
	return replace(input, alphanumAndSpaceOnlyPattern)
}

func InterfaceToJSON(data interface{}) (result string) {
	switch data.(type) {
	case []uint8:
		result = string(data.([]byte))
	case string:
		result = data.(string)
	case nil:
		result = ""
	default:
		resultByte, _ := json.Marshal(data)
		result = string(resultByte)
	}
	return
}
