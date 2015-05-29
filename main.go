package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {

	ln, err := net.Listen("tcp", ":21")

	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {

	conn.Write([]byte("220 Service ready\r\n"))

	buffer := make([]byte, 1024)

	for {

		length, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(err)
			return
		} else {

			command := strings.ToLower(string(buffer[:4]))

			switch command {
			case "user":
				conn.Write([]byte("331 User name ok\r\n"))
			case "pass":
				conn.Write([]byte("230 pass ok\r\n"))
			case "opts":
				conn.Write([]byte("200 ok\r\n"))
			case "pwd":
				conn.Write([]byte("257 ftp/\r\n"))
			}

			fmt.Println(string(buffer[:length]))
		}
	}

}
