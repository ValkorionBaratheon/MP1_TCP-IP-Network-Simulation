package main

import "net"

func unicast_receive(source string, message []byte) {

}

func main() {

	// Server should read its port from the config file.
	ln, _:= net.Listen("tcp", 3)
	defer ln.Close()

	for {

	}

}
