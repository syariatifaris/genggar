package util

import (
	"os/exec"
	"reflect"
	"strings"
)

func GetV4UUID() (string, error) {
	//use satori uuid library
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

func InArrayStr(data string, slice []string) bool {
	for _, s := range slice {
		if s == data {
			return true
		}
	}

	return false
}

//InArray return true if element exist in array
func InArray(array interface{}, val interface{}) bool {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			//DeepEqual return true if two var(value and type) are exact identical
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				return true
			}
		}
	}

	return false
}
