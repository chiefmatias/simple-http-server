package main

import (
	"fmt"
	"net"
	"strings"
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

func handleRequest(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
	}

	request := string(buffer[:n])
	parsedRequest := strings.Split(request, "\r\n")
	requestTarget := strings.Split(parsedRequest[0], " ")
	var responseBody string

	var statusLine, headers, body string

	switch {
	case requestTarget[1] == "/":
		statusLine = "HTTP/1.1 200 OK"
		headers = ""
		body = ""

	case strings.HasPrefix(requestTarget[1], "/echo/"):
		responseText := strings.TrimPrefix(requestTarget[1], "/echo/")

		statusLine = "HTTP/1.1 200 OK"
		headers = fmt.Sprintf("Content-Type: text/plain\r\nContent-Length: %d", len(responseText))
		body = responseText

	case strings.HasPrefix(requestTarget[1], "/user-agent"):
		responseText := strings.TrimPrefix(parsedRequest[2], "User-Agent: ")

		statusLine = "HTTP/1.1 200 OK"
		headers = fmt.Sprintf("Content-Type: text/plain\r\nContent-Length: %d", len(responseText))
		body = responseText

	default:
		statusLine = "HTTP/1.1 404 Not Found"
		headers = ""
		body = ""
		if _, err := conn.Write([]byte(responseBody)); err != nil {
			fmt.Println(err)
		}
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

		handleRequest(conn)
		//fmt.Println("Received request:")
		//fmt.Println(request)

	}
}
