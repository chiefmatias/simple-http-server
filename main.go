package main

import (
	"fmt"
	"net"
	"strings"
)

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

	switch {
	case requestTarget[1] == "/":
		responseBody := "HTTP/1.1 200 OK\r\n\r\n"
		fmt.Println(responseBody)
		if _, err := conn.Write([]byte(responseBody)); err != nil {
			fmt.Println(err)
		}

	case strings.HasPrefix(requestTarget[1], "/echo/"):
		fmt.Println("Resource found:", requestTarget[1])
		responseText := strings.TrimPrefix(requestTarget[1], "/echo/")
		responseBody = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprint(len(responseText)) + "\r\n\r\n" + responseText
		fmt.Println(responseBody)
		if _, err := conn.Write([]byte(responseBody)); err != nil {
			fmt.Println(err)
		}

	default:
		responseBody := "HTTP/1.1 404 Not Found\r\n\r\n"
		fmt.Println("Resource not found:", requestTarget[1])
		if _, err := conn.Write([]byte(responseBody)); err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println(responseBody)
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
