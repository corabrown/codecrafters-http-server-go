package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type responseType string
type endpoint string

const (
	notFound       responseType = "HTTP/1.1 404 Not Found"
	okGetResponse  responseType = "HTTP/1.1 200 OK"
	okPostResponse responseType = "HTTP/1.1 201 Created"

	echoEndpoint      endpoint = "GET /echo/"
	userAgentEndpoint endpoint = "GET /user-agent"
	getFileEndpoint   endpoint = "GET /files/"
	postFileEndpoint  endpoint = "POST /files/"
	baseEndpoint      endpoint = "GET / "

	contentTypeHeader     string = "Content-Type"
	contentLengthHeader   string = "Content-Length"
	contentEncodingHeader string = "Content-Encoding"
	connectionHeader      string = "Connection"

	userAgentHeader      string = "User-Agent"
	acceptEncodingHeader string = "Accept-Encoding"

	CRLF string = "\r\n"

	textContentType string = "text/plain"
	octetStreamType string = "application/octet-stream"
)

var supportedEndpoints = []endpoint{echoEndpoint, userAgentEndpoint, getFileEndpoint, postFileEndpoint, baseEndpoint}

type request struct {
	endpoint      endpoint
	requestTarget string
	userAgent     string
	encodingTypes []string
	contentType   string
	connection    string
	body          string
}

func parseRequest(conn net.Conn) (request, error) {
	reqBytes := make([]byte, 1024)
	n, err := conn.Read(reqBytes)
	if err != nil {
		return request{}, err
	}
	s := string(reqBytes[:n])

	req := request{encodingTypes: make([]string, 0)}

	requestComponents := strings.Split(s, CRLF+CRLF)
	if len(requestComponents) == 2 {
		req.body = requestComponents[1]
	}

	mainRequest := strings.Split(requestComponents[0], CRLF)
	requestLine := mainRequest[0]
	for _, supportedEndpoint := range supportedEndpoints {
		if strings.HasPrefix(requestLine, string(supportedEndpoint)) {
			req.endpoint = supportedEndpoint
		}
	}
	req.requestTarget = strings.Split(strings.TrimPrefix(s, string(req.endpoint)), " ")[0]

	for _, component := range mainRequest[1:] {
		headerPair := strings.Split(component, ": ")
		if len(headerPair) == 2 {
			header, value := headerPair[0], headerPair[1]
			switch header {
			case contentTypeHeader:
				req.contentType = value
			case userAgentHeader:
				req.userAgent = value
			case acceptEncodingHeader:
				compressionTypes := strings.Split(value, ",")
				for _, compressionType := range compressionTypes {
					if strings.ReplaceAll(compressionType, " ", "") == "gzip" {
						req.encodingTypes = append(req.encodingTypes, "gzip")
					}
				}
			case connectionHeader:
				req.connection = value
			}
		}
	}

	fmt.Println("PRINTING ENDPOINT ", req.endpoint)

	return req, nil
}

type httpResponse struct {
	resp            responseType
	contentType     string
	contentLength   int
	contentEncoding string
	connection      string
	body            string
}

func (v httpResponse) format() string {

	headerMap := map[string]string{
		contentTypeHeader:     v.contentType,
		contentLengthHeader:   strconv.Itoa(v.contentLength),
		contentEncodingHeader: v.contentEncoding,
		connectionHeader:      v.connection,
	}

	if v.resp == "" {
		v.resp = notFound
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

func getResponse(conn net.Conn, baseDirectory string) (result httpResponse, err error) {

	req, err := parseRequest(conn)
	if err != nil {
		return httpResponse{resp: notFound}, nil
	}

	result.connection = req.connection

	switch req.endpoint {
	case echoEndpoint:
		result.resp = okGetResponse
		result.contentType = textContentType
		result.contentLength = len(req.requestTarget)
		result.body = req.requestTarget
		for _, compressionType := range req.encodingTypes {
			if compressionType == "gzip" {
				var buf bytes.Buffer
				writer := gzip.NewWriter(&buf)
				if _, err := writer.Write([]byte(result.body)); err != nil {
					result.resp = notFound
				}
				if err := writer.Close(); err != nil {
					result.resp = notFound
				}

				result.body = buf.String()
				result.contentLength = len(result.body)
				result.contentEncoding = "gzip"
			}
		}
		return result, nil
	case userAgentEndpoint:
		result.resp = okGetResponse
		result.contentType = textContentType
		result.contentLength = len(req.userAgent)
		result.body = req.userAgent
		return result, nil
	case getFileEndpoint:
		filename := req.requestTarget
		fullPath := filepath.Join(baseDirectory, filename)

		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			result.resp = notFound
			return result, nil
		}
		content, err := os.ReadFile(fullPath)
		if err != nil {
			result.resp = notFound
			return result, nil
		}
		result.resp = okGetResponse
		result.contentType = octetStreamType
		result.contentLength = int(fileInfo.Size())
		result.body = string(content)
		return result, nil

	case postFileEndpoint:
		filename := req.requestTarget

		if req.contentType != octetStreamType {
			result.resp = notFound
			return result, nil
		}
		filePath := filepath.Join(baseDirectory, filename)
		err := os.WriteFile(filePath, []byte(req.body), 0644)
		if err != nil {
			result.resp = notFound
			return result, nil
		}
		result.resp = okPostResponse
		return result, nil

	case baseEndpoint:
		result.resp = okGetResponse
		return result, nil
	}

	return result, nil
}
