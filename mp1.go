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
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Program state
type LocalProcess struct {
	min_delay int
	max_delay int

	// Simulated process ID of this process
	pid int

	// Maps remote process IDs to IPs and ports
	remote_processes map[int]Server

	// Messages to be sent
	message_queue chan Message

	// Random number generator
	r *rand.Rand
}

// Remote server info
type Server struct {
	ip string
	port uint16
}

// Message to be sent with simulated delay
type Message struct {
	start_time time.Time
	delay int
	sender_pid int
	dest Server
	content string
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Random delay is selected from a uniform distribution between min_delay and max_delay
func rand_delay(local_process *LocalProcess) int {
	return local_process.r.Intn(local_process.max_delay - local_process.min_delay) + local_process.min_delay 
}

// Get the lines of the config file
func read_config() []string {
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

// Parse lines of the config and construct initial program state
func parse_config(local_pid int, lines []string) LocalProcess {
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

	// Circular buffer with size 256
	mq := make(chan Message, 256)
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	return LocalProcess{min_delay, max_delay, local_pid, remote_processes, mq, r}
}

func receive_incoming(ln net.Listener) error {
	conn, err := ln.Accept()
	if err != nil {
		return err
	}

	defer conn.Close()

	// Read the length of the message and sender's PID
	header := make([]int32, 2)
	err = binary.Read(conn, binary.BigEndian, header)
	if err != nil {
		return err
	}

	size, remote_pid := header[0], header[1]

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

	fmt.Printf("Received '%s' from process %d at %s\n", message, remote_pid, time.Now())
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

func send_message(message *Message) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", message.dest.ip, message.dest.port))
	if err != nil {
		return err
	}

	defer conn.Close()

	header := []int32{int32(len(message.content)), int32(message.sender_pid)}

	// Send the message length first, then the PID, then the message
	err = binary.Write(conn, binary.BigEndian, header)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(conn, message.content)
	if err != nil {
		return err
	}

	return nil
}

func queue_message(sender *LocalProcess, dest_pid int, message string) {
	server, ok := sender.remote_processes[dest_pid]

	if !ok {
		log.Printf("Tried to queue message to process %d which does not have entry in config file\n", dest_pid)
		return
	}

	now := time.Now()
	msg_struct := Message{now, rand_delay(sender), sender.pid, server, message}
	sender.message_queue <- msg_struct

	log.Printf("Queuing '%s' to be sent to process %d. Current time is %s\n", message, dest_pid, now)
}

func process_message_queue(local_process *LocalProcess) {
	for {
		// Read a message from the channel
		msg, ok := <- local_process.message_queue

		if !ok {
			return
		}

		now := time.Now().UnixMilli()

		// Check if the message is ready to be sent
		if (now - msg.start_time.UnixMilli() > int64(msg.delay)) {
			// Send it
			err := send_message(&msg)

			if err != nil {
				log.Println(err)
			}
		} else {
			// If it isn't ready to be sent, queue it again
			local_process.message_queue <- msg
		}
	}
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

	lines := read_config()
	local_process := parse_config(pid, lines)
	defer close(local_process.message_queue)
	local_entry, ok := local_process.remote_processes[pid];
	if !ok {
		log.Fatal(fmt.Sprintf("PID %d was supplied but does not exist in config\n", pid))
	}

	go listen_for_incoming(local_entry.port)
	go process_message_queue(&local_process)

	log.Printf("Process %d\n", pid)

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

		go queue_message(&local_process, dest_pid, message)
		if err != nil {
			log.Println(err)
		}
	}
}
