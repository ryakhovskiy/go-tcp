package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const nl byte = '\n'

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if nil != err {
		fmt.Printf("Error while creating UDP Address: %v\n", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if nil != err {
		fmt.Printf("Error while creating creating remote connection: %v\n", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString(nl)
		if nil != err {
			fmt.Printf("<ERROR> Error reading input: %v\n        Try again\n", err)
			continue
		}
		_, err = conn.Write([]byte(line))
		if nil != err {
			fmt.Printf("<ERROR> Error sending data over to remote host: %v\n        Try again\n")
			continue
		}
	}
}
