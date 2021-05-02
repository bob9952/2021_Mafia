package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

const PORT = 8888

func main() {

	s := newServer()
	go s.run()

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(PORT))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to start server: "+err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Server started on port " + strconv.Itoa(PORT))

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to accept client connection: "+err.Error())
			continue
		}
		go s.newClient(conn)
	}
}
