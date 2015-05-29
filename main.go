package main

import (
	"fmt"
	"net"
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

			command := buffer[:4]

			if string(command) == "USER" {
				conn.Write([]byte("331 User name ok\r\n"))
			}

			fmt.Println(string(buffer[:length]))
		}
	}

}
