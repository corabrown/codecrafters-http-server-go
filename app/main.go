package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	directoryPtr := flag.String("directory", "/tmp", "the directory for files")
	flag.Parse()
	var baseDirectory string
	if directoryPtr != nil {
		baseDirectory = *directoryPtr
	}

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

		go func() {
			resp, err := getResponse(conn, baseDirectory)
			if err != nil {
				panic("unable to get a response")
			}

			conn.Write([]byte(resp.format()))
			conn.Close()
		}()

	}
}
