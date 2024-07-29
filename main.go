package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func buildResponse(statusLine, headers, body string) string {
	var response strings.Builder
	response.WriteString(statusLine)
	response.WriteString("\r\n")
	response.WriteString(headers)
	response.WriteString("\r\n")
	response.WriteString("\r\n")
	response.WriteString(body)
	response.WriteString("\r\n")
	return response.String()
}

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

func parseRequest(request string) (Request, error) {

	parsedRequest := strings.Split(request, "\r\n")
	statusLine := strings.Split(parsedRequest[0], " ")

	if len(parsedRequest) < 2 || len(statusLine) < 2 {
		fmt.Println("Invalid request:", request)
		err := fmt.Errorf("Invalid request:", request)
		return Request{}, err
	}

	r := Request{
		Method: statusLine[0],
		Target: statusLine[1],
	}

	for _, header := range parsedRequest[1:] {

		switch {
		case strings.HasPrefix(header, "User-Agent: "):
			r.UserAgent = strings.TrimPrefix(header, "User-Agent: ")

		case strings.HasPrefix(header, "Accept: "):
			r.Accept = strings.TrimPrefix(header, "Accept: ")

		case strings.HasPrefix(header, "Accept-Encoding: "):
			r.Encoding = strings.TrimPrefix(header, "Accept-Encoding: ")

		case strings.HasPrefix(header, "Content-Type: "):
			r.ContentType = strings.TrimPrefix(header, "Content-Type: ")

		case strings.HasPrefix(header, "Content-Length: "):
			r.ContentLength = strings.TrimPrefix(header, "Content-Length: ")
		}

		if r.ContentLength != "0" {
			r.Body = parsedRequest[len(parsedRequest)-1]
		}
	}
	return r, nil
}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	request := string(buffer[:n])
	parsedRequest := strings.Split(request, "\r\n")

	requestTarget := strings.Split(parsedRequest[0], " ")
	requestMethod := (requestTarget[0])
	requestBody := parsedRequest[len(parsedRequest)-1]
	requestContent := parsedRequest[3]
	requestEncoding := parsedRequest[5]

	var statusLine, headers, body string

	switch {
	case requestTarget[1] == "/":
		statusLine = "HTTP/1.1 200 OK"
		headers = ""
		body = ""

	case strings.HasPrefix(requestTarget[1], "/echo/"):
		responseText := strings.TrimPrefix(requestTarget[1], "/echo/")
		statusLine = "HTTP/1.1 200 OK"

		switch {
		case strings.Contains(requestEncoding, "gzip"):
			headers = fmt.Sprintf("Content-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d", len(responseText))
			body = responseText
		default:
			headers = fmt.Sprintf("Content-Type: text/plain\r\nContent-Length: %d", len(responseText))
			body = responseText

		}

	case strings.HasPrefix(requestTarget[1], "/user-agent"):
		responseText := strings.TrimPrefix(parsedRequest[2], "User-Agent: ")

		statusLine = "HTTP/1.1 200 OK"
		headers = fmt.Sprintf("Content-Type: text/plain\r\nContent-Length: %d", len(responseText))
		body = responseText

	case strings.HasPrefix(requestTarget[1], "/files/"):

		fileName := strings.TrimPrefix(requestTarget[1], "/files/")
		file, err := os.ReadFile(os.Args[2] + fileName)
		switch {
		case requestMethod == "GET":
			if err != nil {
				fmt.Println("Error reading file:", err)
				statusLine = "HTTP/1.1 404 Not Found"
				headers = ""
				body = "File not found"
			} else {
				statusLine = "HTTP/1.1 200 OK"
				headers = fmt.Sprintf("Content-Type: application/octet-stream\r\nContent-Length: %d", len(file))
				body = string(file)
			}
		case requestMethod == "POST" && strings.Contains(requestContent, "application/octet-stream"):
			err := os.WriteFile(os.Args[2]+fileName, []byte(requestBody), 0666)
			if err != nil {
				fmt.Println("Error writing file:", err)
				statusLine = "HTTP/1.1 500 Internal Server Error"
			} else {
				statusLine = "HTTP/1.1 201 Created"
			}
		}

	default:
		statusLine = "HTTP/1.1 404 Not Found"
		headers = ""
		body = ""
	}
	response := buildResponse(statusLine, headers, body)
	fmt.Println("\nTarget:", requestTarget[1], "\nResponse:", response)

	if _, err := conn.Write([]byte(response)); err != nil {
		fmt.Println("Error writing response:", err)
	}
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

		go handleRequest(conn)
		//fmt.Println("Received request:")
		//fmt.Println(request)

	}
}
