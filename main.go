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

	buffer := make([]byte, 1024)

	for {

		length, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(string(buffer[:length]))
		}
	}

}
