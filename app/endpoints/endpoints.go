package endpoints

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"strings"

	"github.com/chiefmatias/simple-http-server/app/request"
	"github.com/chiefmatias/simple-http-server/app/response"
)

func BaseEndpoint(req request.Request) (response.Response, error) {
	statusLine := "HTTP/1.1 200 OK"

	res := response.Response{
		StatusLine: statusLine,
	}

	return res, nil
}

func EchoEndpoint(req request.Request) (response.Response, error) {
	responseHeaders := response.ResponseHeaders{}
	statusLine := "HTTP/1.1 200 OK"
	body := strings.TrimPrefix(req.Target, "/echo/")

	switch {
	case strings.Contains(req.Encoding, "gzip"):
		var buffer bytes.Buffer
		writer := gzip.NewWriter(&buffer)
		writer.Write([]byte(body))
		writer.Close()
		body = buffer.String()

		responseHeaders.Encoding = "gzip"
		responseHeaders.ContentType = "text/plain"
		responseHeaders.ContentLength = fmt.Sprintf("%d", len(body))

	default:
		responseHeaders.ContentType = "text/plain"
		responseHeaders.ContentLength = fmt.Sprintf("%d", len(body))
	}

	res := response.Response{
		StatusLine: statusLine,
		Headers:    responseHeaders,
		Body:       body,
	}

	return res, nil
}

func UserAgentEndpoint(req request.Request) (response.Response, error) {
	responseHeaders := response.ResponseHeaders{}
	statusLine := "HTTP/1.1 200 OK"

	responseHeaders.ContentType = "text/plain"
	responseHeaders.ContentLength = fmt.Sprintf("%d", len(req.UserAgent))
	body := req.UserAgent

	res := response.Response{
		StatusLine: statusLine,
		Headers:    responseHeaders,
		Body:       body,
	}

	return res, nil
}

func FilesEndpoint(req request.Request) (response.Response, error) {
	var statusLine string
	var body string
	var err error
	responseHeaders := response.ResponseHeaders{}
	fileName := strings.TrimPrefix(req.Target, "/files/")

	switch req.Method {
	case "GET":
		var file []byte
		file, err = os.ReadFile(os.Args[2] + fileName)
		if err != nil {
			err = fmt.Errorf("error reading file: %w", err)
			statusLine = "HTTP/1.1 404 Not Found"
			body = "File not found"
		} else {
			statusLine = "HTTP/1.1 200 OK"
			responseHeaders.ContentType = "application/octet-stream"
			responseHeaders.ContentLength = fmt.Sprintf("%d", len(file))
			body = string(file)
		}
	case "POST":
		if strings.Contains(req.ContentType, "application/octet-stream") {
			byteBody := []byte(req.Body)
			err = os.WriteFile(os.Args[2]+fileName, byteBody, 0666)
			if err != nil {
				err = fmt.Errorf("error writing file: %w", err)
				statusLine = "HTTP/1.1 500 Internal Server Error"
				body = "Error Writing file"
			} else {
				statusLine = "HTTP/1.1 201 Created"
				body = string(byteBody)
			}
		} else {
			statusLine = "HTTP/1.1 400 Bad Request"
			body = "Invalid content type"
		}
	}
	res := response.Response{
		StatusLine: statusLine,
		Headers:    responseHeaders,
		Body:       body,
	}

	return res, err
}
