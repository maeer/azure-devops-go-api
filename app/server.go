package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
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
		go serve(conn)
	}

}

func serve(conn net.Conn) {
	receiveData := make([]byte, 1024)
	length, err := conn.Read(receiveData)
	request := string(receiveData[:length])
	fmt.Println(request)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
		os.Exit(1)
	}
	requestSegments := strings.Split(request, "\r\n")
	path := strings.Split(request, " ")[1]
	headers := map[string]string{}
	body := ""
	for i := 1; i < len(requestSegments); i++ {
		if requestSegments[i] == "" {
			body = requestSegments[i+1]
			break
		}
		header := strings.Split(requestSegments[i], ": ")
		headers[header[0]] = header[1]
	}
	fmt.Sprintln(requestSegments)
	if path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if segments := strings.Split(path, "/"); segments[1] == "echo" {
		message := segments[2]
		encodeTypes, ok := headers["Accept-Encoding"]
		if !ok || !contains(strings.Split(encodeTypes, ", "), "gzip") {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
		} else {
			var buf bytes.Buffer
			gzipWrite := gzip.NewWriter(&buf)
			defer gzipWrite.Close()
			_, _ = gzipWrite.Write([]byte(message))
			gzipMessage := buf.String()
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: gzip\r\nContent-Length: %d\r\n\r\n%s", len(gzipMessage), gzipMessage)))
		}
	} else if segments[1] == "user-agent" {
		message := headers["User-Agent"]
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
	} else if segments[1] == "files" {
		args := os.Args
		dir := args[2]
		fileName := segments[2]
		if action := strings.Split(requestSegments[0], " ")[0]; action == "GET" {
			data, err := os.ReadFile(dir + fileName)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			} else {
				conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data)))
			}
		} else if action == "POST" {
			_ = os.WriteFile(dir+fileName, []byte(body), 0644)
			conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		}
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func contains(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}
