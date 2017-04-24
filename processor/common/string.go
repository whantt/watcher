package common

import (
	"bytes"
)

func equal(a, b string) bool {
	return a == b
}

func larger(a, b string) bool {
	return bytes.Compare([]byte(a), []byte(b)) == 1
}

func lesser(a, b string) bool {
	return bytes.Compare([]byte(a), []byte(b)) == -1
}

func contains(a, b string) bool {
	return bytes.Contains([]byte(a), []byte(b))
}
