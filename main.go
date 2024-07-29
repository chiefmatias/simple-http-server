package main

import (
	"fmt"

	"net"
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

		}

	}

}
