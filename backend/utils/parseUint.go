package utils

import "strconv"

func ParseUint(str string) (uint, error) {
	parsed, err := strconv.ParseUint(str, 10, 64)
	return uint(parsed), err
}
