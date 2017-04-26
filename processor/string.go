package processor

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

func match(method string, exist bool, a, b string) bool {

	switch method {
	case "exist":
		return exist

	case "empty":
		return a == ""

	case "equal":
		return equal(a, b)

	case "lesser":
		return lesser(a, b)

	case "larger":
		return larger(a, b)

	case "contains":
		return contains(a, b)
	}

	return false
}
