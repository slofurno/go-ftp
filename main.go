package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	filestring = "-rw-r--r-- 1 owner group"
	dirstring  = "drwxr-xr-x 1 owner group"
	rootdir    = "root"
)

func init() {

}

func unixString(fi os.FileInfo) string {
	stamp := fi.ModTime().Format("Jan _2 15:04")
	size := fmt.Sprintf("%13s", strconv.Itoa(int(fi.Size())))

	if fi.IsDir() {
		return fmt.Sprintln(dirstring, size, stamp, fi.Name())
	}

	return fmt.Sprintln(filestring, size, stamp, fi.Name())
}

func main() {

	f, _ := os.Open("root/hey.txt")
	fi, _ := f.Stat()

	line := unixString(fi)

	fmt.Println(line)

	ln, _ := net.Listen("tcp", ":21")

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}

		go handleConnection(conn)
	}

}

func pipeFiles(conn net.Conn, recieved chan<- []byte) {
	var file []byte
	b := make([]byte, 4096)

	for {
		n, err := conn.Read(b)

		if err != nil {
			if err == io.EOF {
				recieved <- file
				file = nil
			} else {
				//fmt.Println(err.Error())
				//reading from closed file
			}
			return
		} else {
			file = append(file, b[:n]...)
		}
	}

}

func passiveMode(ln net.Listener, send <-chan []byte, recieved chan<- []byte, end <-chan bool) {

	//ln, _ := net.Listen("tcp", ":"+port)
	for {
		func() {

			r := make(chan []byte)

			fmt.Println("awaiting next")
			conn, _ := ln.Accept()
			defer conn.Close()
			conn.SetDeadline(time.Time{})

			go pipeFiles(conn, r)

			for {
				select {
				case file := <-r:
					recieved <- file
				case toSend := <-send:
					fmt.Println("data channel: ", toSend)
					conn.Write(toSend)
				case <-end:
					return
				}
			}
		}()
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Time{})
	ln, _ := net.Listen("tcp", ":0")

	defer ln.Close()

	i := strings.LastIndex(ln.Addr().String(), ":")
	port := ln.Addr().String()[i+1:]

	fmt.Println(ln.Addr().Network(), ln.Addr().String())

	send := make(chan []byte)
	end := make(chan bool)
	received := make(chan []byte)

	go passiveMode(ln, send, received, end)

	conn.Write([]byte("220 Service ready\r\n"))

	buffer := make([]byte, 1024)
	var lastparam string

	for {

		length, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(err)
			return
		} else {

			var command string
			var param string
			rawmsg := strings.ToLower(string(buffer[:length]))
			spc := strings.Index(rawmsg, " ")

			fmt.Println("length", length)

			if spc > 0 {
				command = rawmsg[:spc]
				param = rawmsg[spc+1 : length-2]

			} else {
				command = rawmsg[:length-2]
			}

			fmt.Println(command, param)

			switch command {
			case "user":
				conn.Write([]byte("331 User name ok\r\n"))
			case "pass":
				conn.Write([]byte("230 pass ok\r\n"))
			case "opts":
				conn.Write([]byte("200 ok\r\n"))
			case "pwd":
				conn.Write([]byte("257 \"/\" home directory \r\n"))
			case "syst":
				conn.Write([]byte("215 UNIX Type: L8\r\n"))
			case "type":
				conn.Write([]byte("200 ok\r\n"))
			case "cwd":
				conn.Write([]byte("250 ok\r\n"))
			case "epsv":
				conn.Write([]byte("229 Entering Extended Passive Mode (|||" + port + "|)\r\n"))
			case "list":
				conn.Write([]byte("150 mark\r\n"))
				if param == "" {
					f, _ := os.Open("root")
					di, _ := f.Readdir(200)
					for _, fi := range di {
						send <- []byte(unixString(fi))
					}
				}
				end <- true
				conn.Write([]byte("226 ok\r\n"))
				//send <- []byte("-rw-r--r-- 1 owner group           213 Aug 26 16:31 test1.txt\r\n")
				//send <- []byte("-rw-r--r-- 1 owner group           129 Aug 26 16:31 hey.txt\r\n")
			case "noop":
				conn.Write([]byte("200 hi\r\n"))
			case "size":
				func() {
					f, err := os.Open("root/" + param)
					defer f.Close()
					if err != nil {
						fmt.Println("file open error: ", err)
					}
					fi, err := f.Stat()
					if err != nil {
						fmt.Println("fileinfo error: ", err)
					}
					conn.Write([]byte("213 " + strconv.Itoa(int(fi.Size())) + "\r\n"))
				}()
			case "retr":
				func() {
					b, err := ioutil.ReadFile("root/" + param)

					if err != nil {
						fmt.Println("file read error: ", err)
					}
					conn.Write([]byte("150 mark\r\n"))
					send <- b
					end <- true
					conn.Write([]byte("226 ok\r\n"))
				}()
			case "stor":
				func() {
					conn.Write([]byte("150 mark\r\n"))
					file := <-received
					f, _ := os.Create("root/" + param)
					defer f.Close()
					f.Write(file)
					end <- true
					conn.Write([]byte("226 saved\r\n"))
				}()
			case "rnfr":
				lastparam = param
				conn.Write([]byte("350 exists\r\n"))
			case "rnto":
				os.Rename("root/"+lastparam, "root/"+param)
				conn.Write([]byte("250 renamed ok\r\n"))
			case "dele":
				err := os.Remove("root/" + param)
				if err != nil {
					conn.Write([]byte("550 " + err.Error() + "\r\n"))
				} else {
					conn.Write([]byte("250 removed ok\r\n"))
				}
			default:
				//conn.Write([]byte("500 tevs\r\n"))
			}

		}
	}

}
