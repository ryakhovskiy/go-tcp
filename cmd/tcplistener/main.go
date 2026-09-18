package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"ryakhovskiy/httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if nil != err {
		fmt.Printf("Error while creating tcp listener on port %s: %v\n", port, err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if nil != err {
			fmt.Printf("Error while listening port %s: %v\n", port, err)
			return
		}
		defer conn.Close()
		fmt.Println("connection accepted")
		//linesCh := getLinesChannel(conn)
		req, err := request.RequestFromReader(conn)
		if nil != err {
			fmt.Print(err)
			return
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

		fmt.Println("connection closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	buff := make([]byte, 8)
	currentLine := ""
	totalBytesRead := 0
	go func() {
		defer close(lines)
		for {
			i, err := io.Reader.Read(f, buff)
			totalBytesRead += i
			if errors.Is(err, io.EOF) {
				lines <- currentLine
				break
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				currentLine += string(buff[:i])
				lines <- currentLine
				break
			}
			if nil != err {
				fmt.Printf("Error while reading file: %v\nTotal bytes read: %d", err, totalBytesRead)
				return
			}
			newLineIndex := bytes.IndexByte(buff, '\n')
			if newLineIndex == -1 {
				currentLine += string(buff[:i])
			} else {
				currentLine += string(buff[:newLineIndex])
				lines <- currentLine
				if newLineIndex != len(buff)-1 {
					currentLine = string(buff[newLineIndex+1:])
				} else {
					currentLine = ""
				}
			}
		}
	}()
	return lines
}
