package tc

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// TrimSpace 删除头尾的空字符，空格，换行之类的东西, 具体在unicode.IsSpac.
func TrimSpace(raw string) string {
	s := strings.TrimLeftFunc(raw, unicode.IsSpace)
	if s != "" {
		s = strings.TrimRightFunc(s, unicode.IsSpace)
	}
	return s
}

// TrimSplit 按sep拆分，并去掉空字符.
func TrimSplit(raw, sep string) []string {
	var ss []string

	s := TrimSpace(raw)
	if s == "" {
		return ss
	}

	ss = strings.Split(s, sep)
	i := 0
	for _, s := range ss {
		s = TrimSpace(s)
		if s != "" {
			ss[i] = s
			i++
		}
	}
	return ss[:i]
}

// TruncateID returns a shorthand version of a string identifier for convenience.
// A collision with other shorthands is very unlikely, but possible.
// In case of a collision a lookup with TruncIndex.Get() will fail, and the caller
// will need to use a langer prefix, or the full-length Id.
func truncateID(id string) string {
	shortLen := 12
	if len(id) < shortLen {
		shortLen = len(id)
	}
	return id[:shortLen]
}

// GenUID - returns an unique id
func GenUID() string {
	for {
		id := make([]byte, 16)
		if _, err := io.ReadFull(rand.Reader, id); err != nil {
			// This shouldn't happen
			panic(err)
		}

		value := hex.EncodeToString(id)
		// if we try to parse the truncated for as an int and we don't have
		// an error then the value is all numberic and causes issues when
		// used as a hostname. ref #3869
		if _, err := strconv.ParseInt(truncateID(value), 10, 32); err == nil {
			continue
		}
		return value
	}
}

// JoinMap
func JoinMap(params map[string]interface{}, sep string) []byte {
	if params == nil {
		return nil
	}

	var v string
	for key, value := range params {
		v += fmt.Sprintf("%v=%v%v", key, value, sep)
	}

	v = strings.TrimRight(v, sep)

	return []byte(v)
}

// EncodeJSON
func EncodeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// DecodeJSON
func DecodeJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
