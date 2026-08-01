package utils

import (
	"errors"
	"fmt"
	"strings"
)

func ParseDnsKeyValue(s string) (map[string]string, error) {
	result := make(map[string]string)

	i := 0
	n := len(s)

	for i < n {
		// Skip whitespace
		for i < n && s[i] == ' ' {
			i++
		}
		if i >= n {
			break
		}

		// Read key
		start := i
		for i < n && s[i] != '=' {
			i++
		}
		if i >= n {
			return nil, errors.New("missing '='")
		}

		key := s[start:i]
		i++ // skip '='

		// Read value
		var value string

		if i < n && s[i] == '"' {
			i++
			var b strings.Builder

			for i < n {
				switch s[i] {
				case '\\':
					i++
					if i < n {
						b.WriteByte(s[i])
						i++
					}
				case '"':
					i++
					value = b.String()
					goto done
				default:
					b.WriteByte(s[i])
					i++
				}
			}

			return nil, fmt.Errorf("unterminated quote")
		}

		start = i
		for i < n && s[i] != ' ' {
			i++
		}
		value = s[start:i]

	done:
		result[key] = value
	}

	return result, nil
}
