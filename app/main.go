package main

import (
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

		resp, err := getResponse(conn)
		if err != nil {
			panic("FDSJKFH")
		}

		conn.Write([]byte(resp))
		conn.Close()

	}
}

var (
	notFound   string = "HTTP/1.1 404 Not Found\r\n\r\n"
	okResponse string = "HTTP/1.1 200 OK"
)

var echoPrefix string = "GET /echo/"

func getResponse(conn net.Conn) (string, error) {

	req := make([]byte, 1024)
	_, err := conn.Read(req)
	if err != nil {
		return "", err
	}

	s := string(req)

	resp := fmt.Sprintf("%v\r\n\r\n", okResponse)
	if !strings.HasPrefix(s, "GET / HTTP/1.1") && !strings.HasPrefix(s, "GET /echo/") {
		return notFound, nil
	}

	if strings.HasPrefix(s, echoPrefix) {
		stringRequest := strings.Split(strings.TrimPrefix(s, echoPrefix), " ")[0]
		resp = fmt.Sprintf(
			"%v\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v",
			okResponse,
			len(stringRequest),
			stringRequest,
		)

	}

	return resp, nil
}
