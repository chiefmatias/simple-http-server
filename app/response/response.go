package response

import (
	"strings"
)

type Response struct {
	StatusLine string
	Headers    ResponseHeaders
	Body       string
}

type ResponseHeaders struct {
	Encoding      string
	ContentType   string
	ContentLength string
}

func (r Response) BuildHeaders() string {
	var headers strings.Builder
	if r.Headers.Encoding != "" {
		headers.WriteString("Content-Encoding: ")
		headers.WriteString(r.Headers.Encoding)
		headers.WriteString("\r\n")
	}
	if r.Headers.ContentType != "" {
		headers.WriteString("Content-Type: ")
		headers.WriteString(r.Headers.ContentType)
		headers.WriteString("\r\n")
	}
	if r.Headers.ContentLength != "" {
		headers.WriteString("Content-Length: ")
		headers.WriteString(r.Headers.ContentLength)
		headers.WriteString("\r\n")
	}
	return headers.String()
}

func (r Response) BuildResponse() string {
	var responseBuilder strings.Builder
	responseBuilder.WriteString(r.StatusLine)
	responseBuilder.WriteString("\r\n")
	responseBuilder.WriteString(r.BuildHeaders())
	responseBuilder.WriteString("\r\n")
	responseBuilder.WriteString(r.Body)
	responseBuilder.WriteString("\r\n")
	return responseBuilder.String()
}
