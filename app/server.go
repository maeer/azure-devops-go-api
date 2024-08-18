package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
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
	for i := 1; i < len(requestSegments); i++ {
		if requestSegments[i] == "" {
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
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
	} else if segments[1] == "user-agent" {
		message := headers["User-Agent"]
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
