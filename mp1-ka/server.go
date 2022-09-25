package main

import "net"

func unicast_receive(source string, message []byte) {

}

func main() {

	// Server should read its port from the config file.
	ln, _ := net.Listen("tcp", "host:port")
	defer ln.Close()
	for {
		// Can receive messages from cients here.
		conn, _ := ln.Accept()
		_ = conn
	}

}
