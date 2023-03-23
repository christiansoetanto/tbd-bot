package util

import (
	"github.com/christiansoetanto/tbd-bot/domain"
	"reflect"
)

func IsInterfaceNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Slice:
		return reflect.ValueOf(i).Len() == 0
	case reflect.Ptr:
		return reflect.ValueOf(i).IsNil() || IsInterfaceNil(reflect.ValueOf(i).Elem().Interface())
	default:
		return reflect.ValueOf(i).IsZero()
	}
}

func IntRemoveIndex(s []int, index int) []int {
	return append(s[:index], s[index+1:]...)
}

func StringRemoveIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
func VotersRemoveIndex(s []domain.Voter, index int) []domain.Voter {
	return append(s[:index], s[index+1:]...)
}
