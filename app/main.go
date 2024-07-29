package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/chiefmatias/simple-http-server/app/endpoints"
	"github.com/chiefmatias/simple-http-server/app/request"
	"github.com/chiefmatias/simple-http-server/app/response"
)

func handleRequest(conn net.Conn, req request.Request) (response.Response, error) {
	var res response.Response
	var err error

	switch {
	case req.Target == "/":
		res, err = endpoints.BaseEndpoint(req)
	case strings.HasPrefix(req.Target, "/echo/"):
		res, err = endpoints.EchoEndpoint(req)
	case strings.HasPrefix(req.Target, "/user-agent"):
		res, err = endpoints.UserAgentEndpoint(req)
	case strings.HasPrefix(req.Target, "/files/"):
		res, err = endpoints.FilesEndpoint(req)

	default:
		res.StatusLine = "HTTP/1.1 404 Not Found"
		res.Body = "Not Found"
	}

	if err != nil {
		err = fmt.Errorf("error handling request: %w", err)
	}

	fmt.Println("\nTarget:", req.Target, "\nResponse:", res)

	if _, writeErr := conn.Write([]byte(res.BuildResponse())); writeErr != nil {
		return res, fmt.Errorf("error writing response: %w", writeErr)
	}
	return res, err
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

			req, err := request.ParseRequest(buffer[:n])
			if err != nil {
				fmt.Println("Error parsing request:", err)
				return
			}

			_, err = handleRequest(conn, req)
			if err != nil {
				fmt.Println("error handling request:", err)
				return
			}
		}(conn)
	}
}
