package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	directoryPtr := flag.String("directory", "/tmp", "the directory for files")
	var baseDirectory string
	if directoryPtr != nil {
		baseDirectory = *directoryPtr
	}
	fmt.Println(baseDirectory)

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

		go func() {
			resp, err := getResponse(conn, baseDirectory)
			if err != nil {
				panic("FDSJKFH")
			}

			conn.Write([]byte(resp))
			conn.Close()
		}()

	}
}

const (
	notFound   string = "HTTP/1.1 404 Not Found"
	okResponse string = "HTTP/1.1 200 OK"

	echoPrefix      string = "GET /echo/"
	userAgentPrefix string = "GET /user-agent"
	filePrefix      string = "GET /files/"
	userAgentHeader string = "User-Agent: "

	CRLF string = "\r\n"

	textContentType string = "text/plain"
	octetStreamType string = "application/octet-stream"
)

func getResponse(conn net.Conn, baseDirectory string) (string, error) {

	req := make([]byte, 1024)
	n, err := conn.Read(req)
	if err != nil {
		return "", err
	}

	s := string(req[:n])

	resp := fmt.Sprintf("%v%v%v", okResponse, CRLF, CRLF)

	if strings.HasPrefix(s, echoPrefix) {
		stringRequest := strings.Split(strings.TrimPrefix(s, echoPrefix), " ")[0]
		return getOkResponse(textContentType, len(stringRequest), stringRequest), nil
	}
	if strings.HasPrefix(s, userAgentPrefix) {
		i := strings.Index(s, userAgentHeader)
		if i == -1 {
			return getOkResponse(textContentType, 0, ""), nil
		}
		userAgent := strings.Split(strings.TrimPrefix(s[i:], userAgentHeader), CRLF)[0]
		return getOkResponse(textContentType, len(userAgent), userAgent), nil
	}
	if strings.HasPrefix(s, filePrefix) {
		filename := strings.Split(strings.TrimPrefix(s, filePrefix), " ")[0]

		fullPath := filepath.Join(baseDirectory, filename)

		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return notFound + CRLF + CRLF, nil
		}
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return notFound + CRLF + CRLF, nil
		}
		return getOkResponse(octetStreamType, int(fileInfo.Size()), string(content)), nil
	}

	if !strings.HasPrefix(s, "GET / HTTP/1.1") {
		return notFound + CRLF + CRLF, nil
	}

	return resp, nil
}

func getOkResponse(contentType string, contentLength int, content string) string {
	return fmt.Sprintf(
		"%v%vContent-Type: %v%vContent-Length: %d%v%v%v",
		okResponse,
		CRLF,
		contentType,
		CRLF,
		contentLength,
		CRLF,
		CRLF,
		content,
	)
}
