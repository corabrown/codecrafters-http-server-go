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
	flag.Parse()
	var baseDirectory string
	if directoryPtr != nil {
		baseDirectory = *directoryPtr
	}

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
	notFound       string = "HTTP/1.1 404 Not Found"
	okGetResponse  string = "HTTP/1.1 200 OK"
	okPostResponse string = "HTTP/1.1 201 OK"

	echoPrefix      string = "GET /echo/"
	userAgentPrefix string = "GET /user-agent"
	getFilePrefix   string = "GET /files/"
	postFilePrefix  string = "POST /files/"

	userAgentHeader     string = "User-Agent: "
	contentTypeHeader   string = "Content-Type: "
	contentLengthHeader string = "Content-Length"

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

	resp := fmt.Sprintf("%v%v%v", okGetResponse, CRLF, CRLF)

	if strings.HasPrefix(s, echoPrefix) {
		stringRequest := strings.Split(strings.TrimPrefix(s, echoPrefix), " ")[0]
		return getOkResponse(textContentType, len(stringRequest), stringRequest), nil
	}
	if strings.HasPrefix(s, userAgentPrefix) {
		userAgent, ok := getHeaderValue(s, userAgentHeader)
		if !ok {
			return getNotOkResponse(), nil
		}
		return getOkResponse(textContentType, len(userAgent), userAgent), nil
	}
	if strings.HasPrefix(s, getFilePrefix) {
		filename := strings.Split(strings.TrimPrefix(s, getFilePrefix), " ")[0]

		fullPath := filepath.Join(baseDirectory, filename)

		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return getNotOkResponse(), nil
		}
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return getNotOkResponse(), nil
		}
		return getOkResponse(octetStreamType, int(fileInfo.Size()), string(content)), nil
	}
	if strings.HasPrefix(s, postFilePrefix) {
		filename := strings.Split(strings.TrimPrefix(s, postFilePrefix), " ")[0]

		contentType, ok := getHeaderValue(s, contentTypeHeader)
		if !ok || (contentType != octetStreamType) {
			return getNotOkResponse(), nil
		}

		components := strings.Split(s, CRLF)
		content := components[len(components)-1]

		filePath := filepath.Join(baseDirectory, filename)

		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			return getNotOkResponse(), nil
		}

		return getEmptyOkResponse(), nil
	}

	if !strings.HasPrefix(s, "GET / HTTP/1.1") {
		return getNotOkResponse(), nil
	}

	return resp, nil
}

func getOkResponse(contentType string, contentLength int, content string) string {
	return fmt.Sprintf(
		"%v%v%v%v%v%v%d%v%v%v",
		okGetResponse,
		CRLF,
		contentTypeHeader,
		contentType,
		CRLF,
		contentLengthHeader,
		contentLength,
		CRLF,
		CRLF,
		content,
	)
}

func getOkPostResponse() string {
	return okPostResponse + CRLF + CRLF
}

func getNotOkResponse() string {
	return notFound + CRLF + CRLF
}

func getHeaderValue(requestString, header string) (value string, ok bool) {
	i := strings.Index(requestString, header)
	if i == -1 {
		return "", false
	}
	return strings.Split(strings.TrimPrefix(requestString[i:], header), CRLF)[0], true
}
