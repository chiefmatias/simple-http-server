package request

import (
	"fmt"
	"strings"
)

type Request struct {
	Method        string
	Target        string
	UserAgent     string
	Accept        string
	Encoding      string
	ContentType   string
	ContentLength string
	Body          string
}

func ParseRequest(buffer []byte) (Request, error) {
	requestBuffer := string(buffer)
	parsedRequest := strings.Split(requestBuffer, "\r\n")
	statusLine := strings.Split(parsedRequest[0], " ")

	if len(parsedRequest) < 2 || len(statusLine) < 2 {
		return Request{}, fmt.Errorf("invalid request: %s", requestBuffer)
	}

	request := Request{
		Method: statusLine[0],
		Target: statusLine[1],
	}

	for _, header := range parsedRequest[1:] {
		switch {
		case strings.HasPrefix(header, "User-Agent: "):
			request.UserAgent = strings.TrimPrefix(header, "User-Agent: ")

		case strings.HasPrefix(header, "Accept: "):
			request.Accept = strings.TrimPrefix(header, "Accept: ")

		case strings.HasPrefix(header, "Accept-Encoding: "):
			request.Encoding = strings.TrimPrefix(header, "Accept-Encoding: ")

		case strings.HasPrefix(header, "Content-Type: "):
			request.ContentType = strings.TrimPrefix(header, "Content-Type: ")

		case strings.HasPrefix(header, "Content-Length: "):
			request.ContentLength = strings.TrimPrefix(header, "Content-Length: ")
		}
	}
	if request.ContentLength != "0" {
		request.Body = parsedRequest[len(parsedRequest)-1]
	}

	return request, nil
}
