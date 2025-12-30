package utils

import (
	"errors"
	"os"
	"reflect"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	// An error can also be returned if a file exists but the user lacks permissions,
	// so it's not enough to check err == nil.
	if err != nil {
		// Log a different error here if needed.
		return false
	}

	// We also check if it's a directory
	return !info.IsDir()
}

func IsEmptyStruct[T any](data T) bool {
	return reflect.ValueOf(data).IsZero()
}
