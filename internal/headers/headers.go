package headers

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIdx := -1
	for i := 0; i < len(data)-1; i++ {
		if data[i] == '\r' && data[i+1] == '\n' {
			crlfIdx = i
			break
		}
	}
	if crlfIdx == -1 {
		return 0, false, nil
	}
	if crlfIdx == 0 {
		return 2, true, nil
	}

	line := string(data[:crlfIdx])
	consumed := crlfIdx + 2

	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return 0, false, errors.New("malformed header: missing colon")
	}

	key := strings.ToLower(line[:colonIdx])
	pattern := `^[A-Za-z0-9!#\$%&'*+\-.^_\` + "`" + `|~]+$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(key) {
		return 0, false, errors.New("malformed header: invalid characters")
	}

	rawValue := line[colonIdx+1:]

	if strings.ContainsAny(key, " \t") {
		return 0, false, errors.New("malformed header: whitespace in key")
	}
	if len(rawValue) == 0 || rawValue[0] != ' ' {
		return 0, false, errors.New("malformed header: missing space after colon")
	}

	val, ok := h[key]
	if !ok {
		h[key] = strings.TrimSpace(rawValue)
	} else {
		h[key] = fmt.Sprintf("%s, %s", val, strings.TrimSpace(rawValue))
	}

	return consumed, false, nil
}
