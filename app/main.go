package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		reader := bufio.NewReader(conn)

		s, err := reader.ReadString('\n')
		if err != nil {
			panic("ahhhh")
		}

		requestComponents := strings.Split(s, " ")

		var resp response
		if len(requestComponents) < 2 {
			resp = notFound
		} else {
			if requestComponents[1] == "/" {
				resp = okResponse
			} else {
				resp = notFound
			}
		}

		conn.Write(resp)
		conn.Close()

	}
}

type response []byte

var (
	notFound   response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
	okResponse response = []byte("HTTP/1.1 200 OK\r\n\r\n")
)
