package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type responseType string

const (
	notFound       responseType = "HTTP/1.1 404 Not Found"
	okGetResponse  responseType = "HTTP/1.1 200 OK"
	okPostResponse responseType = "HTTP/1.1 201 Created"

	echoPrefix      string = "GET /echo/"
	userAgentPrefix string = "GET /user-agent"
	getFilePrefix   string = "GET /files/"
	postFilePrefix  string = "POST /files/"

	userAgentHeader       string = "User-Agent"
	contentTypeHeader     string = "Content-Type"
	contentLengthHeader   string = "Content-Length"
	acceptEncodingHeader  string = "Accept-Encoding"
	contentEncodingHeader string = "Content-Encoding"

	CRLF string = "\r\n"

	textContentType string = "text/plain"
	octetStreamType string = "application/octet-stream"
)

func getResponse(conn net.Conn, baseDirectory string) (httpResponse, error) {

	req := make([]byte, 1024)
	n, err := conn.Read(req)
	if err != nil {
		return httpResponse{}, err
	}

	s := string(req[:n])
	resp := httpResponse{resp: okGetResponse}

	if strings.HasPrefix(s, echoPrefix) {
		stringRequest := strings.Split(strings.TrimPrefix(s, echoPrefix), " ")[0]
		resp := httpResponse{
			resp:          okGetResponse,
			contentType:   textContentType,
			contentLength: len(stringRequest),
			body:          stringRequest,
		}
		if contentTypes, ok := getHeaderValue(s, acceptEncodingHeader); ok {
			compressionTypes := strings.Split(contentTypes, ",")
			for _, compressionType := range compressionTypes {
				textOnly := strings.ReplaceAll(compressionType, " ", "")
				if textOnly == "gzip" {
					resp.contentEncoding = "gzip"

					var buf bytes.Buffer 
					writer := gzip.NewWriter(&buf)
					if _, err := writer.Write([]byte(resp.body)); err != nil {
						return httpResponse{resp: notFound}, nil 
					}
					if err := writer.Close(); err != nil {
						return httpResponse{resp: notFound}, nil 
					}

					resp.body = buf.String()
					resp.contentLength = len(resp.body)
				}
			}
		}
		return resp, nil
	}
	if strings.HasPrefix(s, userAgentPrefix) {
		userAgent, ok := getHeaderValue(s, userAgentHeader)
		if !ok {
			return httpResponse{resp: notFound}, nil
		}
		return httpResponse{
			resp:          okGetResponse,
			contentType:   textContentType,
			contentLength: len(userAgent),
			body:          userAgent,
		}, nil
	}
	if strings.HasPrefix(s, getFilePrefix) {
		filename := strings.Split(strings.TrimPrefix(s, getFilePrefix), " ")[0]

		fullPath := filepath.Join(baseDirectory, filename)

		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return httpResponse{resp: notFound}, nil
		}
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return httpResponse{resp: notFound}, nil
		}

		return httpResponse{
			resp:          okGetResponse,
			contentType:   octetStreamType,
			contentLength: int(fileInfo.Size()),
			body:          string(content),
		}, nil
	}
	if strings.HasPrefix(s, postFilePrefix) {
		filename := strings.Split(strings.TrimPrefix(s, postFilePrefix), " ")[0]

		contentType, ok := getHeaderValue(s, contentTypeHeader)
		if !ok || (contentType != octetStreamType) {
			return httpResponse{resp: notFound}, nil
		}

		components := strings.Split(s, CRLF)
		content := components[len(components)-1]

		filePath := filepath.Join(baseDirectory, filename)

		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			return httpResponse{resp: notFound}, nil
		}

		return httpResponse{resp: okPostResponse}, nil
	}

	if !strings.HasPrefix(s, "GET / HTTP/1.1") {
		return httpResponse{resp: notFound}, nil
	}

	return resp, nil
}

func getHeaderValue(requestString, header string) (value string, ok bool) {
	i := strings.Index(requestString, header)
	if i == -1 {
		return "", false
	}
	return strings.Split(strings.TrimPrefix(requestString[i:], fmt.Sprintf("%v: ", header)), CRLF)[0], true
}

type httpResponse struct {
	resp            responseType
	contentType     string
	contentLength   int
	contentEncoding string
	body            string
}

func (v httpResponse) format() string {

	headerMap := map[string]string{
		contentTypeHeader:     v.contentType,
		contentLengthHeader:   strconv.Itoa(v.contentLength),
		contentEncodingHeader: v.contentEncoding,
	}

	responseString := string(v.resp)

	for k, v := range headerMap {
		if v != "" {
			responseString = responseString + CRLF + fmt.Sprintf("%v: %v", k, v)
		}
	}

	responseString = responseString + CRLF + CRLF + v.body

	return responseString
}
