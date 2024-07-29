package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
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

func parseRequest(buffer []byte) (Request, error) {
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

func (response Response) buildHeaders() string {
	var headers strings.Builder
	if response.Headers.Encoding != "" {
		headers.WriteString("Content-Encoding: ")
		headers.WriteString(response.Headers.Encoding)
		headers.WriteString("\r\n")
	}
	if response.Headers.ContentType != "" {
		headers.WriteString("Content-Type: ")
		headers.WriteString(response.Headers.ContentType)
		headers.WriteString("\r\n")
	}
	if response.Headers.ContentLength != "" {
		headers.WriteString("Content-Length: ")
		headers.WriteString(response.Headers.ContentLength)
		headers.WriteString("\r\n")
	}
	return headers.String()
}

func (response Response) buildResponse() string {
	var responseBuilder strings.Builder
	responseBuilder.WriteString(response.StatusLine)
	responseBuilder.WriteString("\r\n")
	responseBuilder.WriteString(response.buildHeaders())
	responseBuilder.WriteString("\r\n")
	responseBuilder.WriteString(response.Body)
	responseBuilder.WriteString("\r\n")
	return responseBuilder.String()
}

func BaseEndpoint(request Request) (Response, error) {
	statusLine := "HTTP/1.1 200 OK"

	response := Response{
		StatusLine: statusLine,
	}

	return response, nil
}

func EchoEndpoint(request Request) (Response, error) {
	responseHeaders := ResponseHeaders{}
	statusLine := "HTTP/1.1 200 OK"
	body := strings.TrimPrefix(request.Target, "/echo/")

	switch {
	case strings.Contains(request.Encoding, "gzip"):
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

	response := Response{
		StatusLine: statusLine,
		Headers:    responseHeaders,
		Body:       body,
	}

	return response, nil
}

func UserAgentEndpoint(request Request) (Response, error) {
	responseHeaders := ResponseHeaders{}
	statusLine := "HTTP/1.1 200 OK"

	responseHeaders.ContentType = "text/plain"
	responseHeaders.ContentLength = fmt.Sprintf("%d", len(request.UserAgent))
	body := request.UserAgent

	response := Response{
		StatusLine: statusLine,
		Headers:    responseHeaders,
		Body:       body,
	}

	return response, nil
}

func FilesEndpoint(request Request) (Response, error) {
	var statusLine string
	var body string
	var err error
	responseHeaders := ResponseHeaders{}
	fileName := strings.TrimPrefix(request.Target, "/files/")

	switch request.Method {
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
		if strings.Contains(request.ContentType, "application/octet-stream") {
			byteBody := []byte(request.Body)
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
	response := Response{
		StatusLine: statusLine,
		Headers:    responseHeaders,
		Body:       body,
	}

	return response, err
}

func handleRequest(conn net.Conn, request Request) (Response, error) {
	var response Response
	var err error

	switch {
	case request.Target == "/":
		response, err = BaseEndpoint(request)
	case strings.HasPrefix(request.Target, "/echo/"):
		response, err = EchoEndpoint(request)
	case strings.HasPrefix(request.Target, "/user-agent"):
		response, err = UserAgentEndpoint(request)
	case strings.HasPrefix(request.Target, "/files/"):
		response, err = FilesEndpoint(request)

	default:
		response.StatusLine = "HTTP/1.1 404 Not Found"
		response.Body = "Not Found"
	}

	if err != nil {
		err = fmt.Errorf("error handling request: %w", err)
	}

	fmt.Println("\nTarget:", request.Target, "\nResponse:", response)

	if _, writeErr := conn.Write([]byte(response.buildResponse())); err != nil {
		return response, fmt.Errorf("error writing response: %w", writeErr)
	}
	return response, err
}

func main() {
	listener, err := net.Listen("tcp", ":4221")
	if err != nil {
		fmt.Println("Error creating listener:", err)
		return
	}

	defer listener.Close()

	fmt.Println("Waiting for connection...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("Accepted connection from:", conn.RemoteAddr())

		go func(conn net.Conn) {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(5 * time.Second))

			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Error reading from connection:", err)
				return
			}

			request, err := parseRequest(buffer[:n])
			if err != nil {
				fmt.Println("Error parsing request:", err)
				return
			}

			_, err = handleRequest(conn, request)
			if err != nil {
				fmt.Println("error handling request:", err)
				return
			}

			//fmt.Println("Received request:")
			//fmt.Println(request)
		}(conn)
	}

}
