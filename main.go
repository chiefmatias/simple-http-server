package main

import (
	"fmt"
	"net"
	"strings"
)

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
			return
		}

		fmt.Println("Accepted connection from:", conn.RemoteAddr())

		if _, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
			fmt.Println(err)

			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Error reading from connection:", err)
			}

			request := string(buffer[:n])

			//fmt.Println("Received request:")

			//fmt.Println(request)

			parsedRequest := strings.Split(request, "\r\n")
			requestTarget := strings.Split(parsedRequest[0], " ")

			if requestTarget[1] != "/" {
				fmt.Println("Resource not found:", requestTarget[1])
				if _, err := conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n")); err != nil {
					fmt.Println(err)
				}

			} else {
				fmt.Println("Resource found: /")
				if _, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
