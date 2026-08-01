package utils

import (
	"errors"
	"strings"
)

type HttpVersion uint8

const (
	Http11 HttpVersion = 1 << iota
	Http2
	Http3

	HttpAll = Http11 | Http2 | Http3
)

func (v HttpVersion) String() string {
	var s []string
	if v&Http3 == Http3 {
		s = append(s, "h3")
	}
	if v&Http2 == Http2 {
		s = append(s, "h2")
	}
	if v&Http11 == Http11 {
		s = append(s, "http1.1")
	}
	return strings.Join(s, ",")
}

func (v HttpVersion) IsValid() bool {
	return v&HttpAll == v
}

func ParseHttpVersion(s string) (HttpVersion, error) {
	items := strings.Split(s, ",")
	var final HttpVersion
	for _, item := range items {
		switch item {
		case "h3":
			final |= Http3
		case "h2":
			final |= Http2
		case "http1.1":
			final |= Http11
		default:
			return HttpAll, errors.New("invalid http version")
		}
	}
	return final, nil
}
