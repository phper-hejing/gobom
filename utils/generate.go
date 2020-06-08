package utils

import "strconv"

func GenerateId() string {
	return strconv.Itoa(int(Now()))
}
