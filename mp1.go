/*
When you type `send (pid) (message)` to send a message to a remote process,
the length of the message is computed and sent as a signed 32-bit int in Big Endian
format. This is followed by the calling process's PID, also as a signed 32 bit
Big Endian int, and finally by the message which is a plain ASCII string.
There is no termination character because the message length was sent first,
so the receiver knows how many ASCII characters to expect after it receives the PID.
The message can contain spaces and can be any length.
*/
package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// Immutable program info specified in config file
type LocalProcess struct {
	min_delay int
	max_delay int

	// Simulated process ID of this process
	pid int

	// Maps remote process IDs to IPs and ports
	remote_processes map[int]Server
}

type Server struct {
	ip   string
	port uint16
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func readConfig() []string {
	file, err := os.Open("./config.txt")
	check(err)

	out := make([]string, 0)

	// Release file handle when this function returns
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	err = scanner.Err()
	check(err)

	return out
}

func parseConfig(local_pid int, lines []string) LocalProcess {
	var min_delay int
	var max_delay int
	remote_processes := make(map[int]Server)

	_, err := fmt.Sscan(lines[0], &min_delay, &max_delay)
	check(err)

	for _, line := range lines[1:] {
		var pid int
		var ip string
		var port uint16

		_, err := fmt.Sscan(line, &pid, &ip, &port)
		check(err)

		server := Server{ip, port}
		remote_processes[pid] = server
	}

	return LocalProcess{min_delay, max_delay, local_pid, remote_processes}
}

func receive_incoming(ln net.Listener) error {
	conn, err := ln.Accept()
	if err != nil {
		return err
	}

	defer conn.Close()

	// Read the length of the message first
	var size int32
	err = binary.Read(conn, binary.BigEndian, &size)
	if err != nil {
		return err
	}

	// Read the sending process's PID
	var remote_pid int32
	err = binary.Read(conn, binary.BigEndian, &remote_pid)
	if err != nil {
		return err
	}

	// Read the actual message
	raw_data := make([]byte, size)
	_, err = conn.Read(raw_data)
	if err != nil {
		return err
	}

	message := string(raw_data)
	if err != nil {
		return err
	}

	fmt.Printf("Received %s from process %d\n", message, remote_pid)
	return nil
}

func listen_for_incoming(port uint16) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println(err)
		return
	}

	for {
		err = receive_incoming(ln)
		if err != nil {
			log.Println(err)
		}
	}
}

func send_message(sender LocalProcess, dest_pid int, message string) error {
	server, ok := sender.remote_processes[dest_pid]

	if !ok {
		return errors.New(fmt.Sprintf("Tried to send message to process %d which does not have entry in config file", dest_pid))
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", server.ip, server.port))
	if err != nil {
		return err
	}

	defer conn.Close()

	// Send the message length first, then the PID, then the message
	err = binary.Write(conn, binary.BigEndian, int32(len(message)))
	if err != nil {
		return err
	}

	err = binary.Write(conn, binary.BigEndian, int32(sender.pid))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(conn, message)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	args := os.Args

	if len(args) <= 1 {
		log.Fatal("Provide PID as first argument")
	}

	pid, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatal(err)
	}

	lines := readConfig()
	local_process := parseConfig(pid, lines)
	local_entry, ok := local_process.remote_processes[pid];
	if !ok {
		log.Fatal(fmt.Sprintf("PID %d was supplied but does not exist in config\n", pid))
	}

	go listen_for_incoming(local_entry.port)

	fmt.Printf("Process %d\n", pid)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")

		if tokens[0] == "q" {
			return
		}

		if tokens[0] != "send" {
			log.Println("Type `send (pid) (message)` to send a message")
			continue
		}

		dest_pid, err := strconv.Atoi(tokens[1])
		if err != nil {
			log.Println(err)
			continue
		}

		message := strings.Join(tokens[2:], " ")

		err = send_message(local_process, dest_pid, message)
		if err != nil {
			log.Println(err)
		}
	}
}
