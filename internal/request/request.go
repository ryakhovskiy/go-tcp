package request

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

const (
	stateInitialized = iota
	stateDone
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type requestParser struct {
	request *Request
	state   int
}

var allowedHttpMethods = []string{"GET", "HEAD", "OPTIONS", "TRACE", "PUT", "DELETE", "POST", "PATCH", "CONNECT"}

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	p := &requestParser{
		request: &Request{},
		state:   stateInitialized,
	}

	for p.state != stateDone {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}
		n, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			p.state = stateDone
			break
		}
		if err != nil {
			return nil, err
		}
		readToIndex += n
		parsed, err := p.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[parsed:readToIndex])
		readToIndex -= parsed
	}

	return p.request, nil
}

func (p *requestParser) parse(data []byte) (int, error) {
	switch p.state {
	case stateInitialized:
		n, rl, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		p.request.RequestLine = rl
		p.state = stateDone
		return n, nil
	case stateDone:
		return 0, errors.New("error: trying to read data in a done state")
	default:
		return 0, errors.New("error: unknown state")
	}
}

func parseRequestLine(data []byte) (int, RequestLine, error) {
	idx := -1
	for i := 0; i < len(data)-1; i++ {
		if data[i] == '\r' && data[i+1] == '\n' {
			idx = i
			break
		}
	}
	if idx == -1 {
		return 0, RequestLine{}, nil
	}

	line := string(data[:idx])
	consumed := idx + 2

	if len(line) == 0 {
		return 0, RequestLine{}, errors.New("request line is empty")
	}

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return 0, RequestLine{}, fmt.Errorf("request line does not have exactly 3 parts: %s", line)
	}

	method := parts[0]
	if !slices.Contains(allowedHttpMethods, method) {
		return 0, RequestLine{}, fmt.Errorf("request line has invalid HTTP method: %s", method)
	}
	requestTarget := parts[1]
	if !strings.HasPrefix(requestTarget, "/") {
		return 0, RequestLine{}, fmt.Errorf("request line has invalid request target: %s", requestTarget)
	}
	httpVersion := parts[2]
	if !strings.HasPrefix(httpVersion, "HTTP/") {
		return 0, RequestLine{}, fmt.Errorf("request line has malformed http version: %s", httpVersion)
	}
	httpVersion = strings.TrimPrefix(httpVersion, "HTTP/")
	if httpVersion != "1.1" {
		return 0, RequestLine{}, fmt.Errorf("HTTP/1.1 is the only supported version, got: %s", httpVersion)
	}

	return consumed, RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: requestTarget,
		Method:        method,
	}, nil
}
