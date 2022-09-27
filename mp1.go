package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Process struct
// This represents a process that can unicast_send
// and unicast_receive to and from other processes.
type Process struct {
	ip        string
	port      int
	min_delay int
	max_delay int
	// process ID of this process
	pid int32
	// Maps remote process IDs to IPs and ports
	remote_processes map[int32]Process
}

// Returns the delay for the current process.
func (process *Process) get_delay() (int, int) {
	return process.min_delay, process.max_delay
}

// Gets the file handle for the config file.
func get_config_file() (file *os.File) {
	file, err := os.Open("./config.txt")
	if err != nil {
		fmt.Println("config.txt file not found")
		os.Exit(1)
	}
	return file
}

// Populates a map of processes that the current
// process can unicast_send and unicast_receive from.
func (process *Process) read_remote_processes(file *os.File) {
	process.remote_processes = make(map[int32]Process)
	for {
		var (
			pid  int32
			ip   string
			port int
		)
		_, err := fmt.Fscanln(file, &pid, &ip, &port)
		if err == io.EOF {
			return
		}
		// If the PID is the id of the current process
		// No need to put an entry into the map.
		if pid == process.pid {
			process.ip = ip
			process.port = port
		} else {
			// Otherwise creates a new process struct and puts it in the map.
			remote_process := Process{
				pid:  pid,
				ip:   ip,
				port: port,
			}
			process.remote_processes[pid] = remote_process
		}
	}
}

// Reads the configuration file provided.
func (process *Process) read_config() {
	file := get_config_file()
	fmt.Fscanln(file, &process.min_delay, &process.max_delay)
	process.read_remote_processes(file)
}

func (process *Process) unicast_send(destination string, message []byte) {
	id, err := strconv.ParseInt(destination, 10, 32)
	if err != nil {
		fmt.Println("error: Process ID must be integer.")
		return
	}
	pid := int32(id)
	// Gets the process ID and port of
	// the receving process.
	ip := process.remote_processes[pid].ip
	port := strconv.Itoa(process.remote_processes[pid].port)
	conn, err := net.Dial("tcp", ip+":"+port)
	// Checks for connection issues with receiving process.
	if err != nil {
		fmt.Println("error: Process ID might not exist in the config.txt file.")
		return
	}

	// Stimulates the delay of sending a message.
	min_delay, max_delay := process.get_delay()
	duration := time.Duration(rand.Intn(max_delay) + min_delay)
	time.Sleep(duration)

	// Writes the process ID into the TCP channel.
	binary.Write(conn, binary.BigEndian, process.pid)
	// Writes the message into the TCP channel.
	conn.Write(message)
	conn.Close()
	fmt.Printf("Sent " + string(message) + " to process %d, system time is %v\n", pid, time.Now())
}

func (process *Process) unicast_recv(source net.Conn, msg []byte) {
	var pid int32
	binary.Read(source, binary.BigEndian, &pid)
	source.Read(msg)
	fmt.Printf("Received " + from_c_str(msg) + " from process %d, system time is %v\n>> ", pid, time.Now())
}

func (process *Process) get_command() (string, string, error) {
	// Reads the command from Stdin
	command, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	// Trims the new line obtained when the user hits Enter.
	command = strings.TrimSuffix(command, "\r\n")
	// If the command is 'q' exits the program.
	if command == "q" {
		fmt.Println("Closing TCP server...")
		os.Exit(0)
	}
	// Checks that the command is valid command.
	commandArray := strings.Split(string(command), " ")
	if len(commandArray) < 3 {
		return "", "", errors.New(`error: Invalid command, must be of the form send [Integer] [String]`)
	}
	message := strings.Join(commandArray[2:], " ")
	destination := commandArray[1]
	return message, destination, nil
}

func (process *Process) recv_commands() {
	// Receives commands from the user.
	// User can press 'q' to quit the program.
	for {
		fmt.Printf("Please input a command or 'q' to quit \n>> ")
		message, destination, err := process.get_command()
		if err != nil {
			fmt.Println(err)
			continue
		}
		// Once the command is received, sends the appropriate unicast to
		// the receiving process.
		process.unicast_send(destination, []byte(message))
	}
}

func from_c_str(bytes []byte) string {
	null := 0
	for i, b := range bytes {
		if b == 0 {
			null = i
			break
		}
	}

	return string(bytes[0:null])
}

func (process *Process) recv_messages() {
	// Obtains the process port.
	port := strconv.Itoa(process.port)
	ln, _ := net.Listen("tcp", process.ip+":"+port)
	for {
		source, _ := ln.Accept()
		msg := make([]byte, 4096)
		// Receives an incoming message.
		go process.unicast_recv(source, msg)
	}
}

func main() {
	// Parses the process id received from the user in the command line.
	id, err := strconv.ParseInt(os.Args[1], 10, 32)
	if err != nil {
		fmt.Println("error: Must provide a valid integer for process ID.")
		os.Exit(1)
	}
	pid := int32(id)

	// A process struct, that represents the current process.
	// Contains information about the ip and port as well
	// min and max delay, as a map to other processes.
	process := Process{pid: pid}

	// Reads the config.txt file to create a map
	// from process id's to their respective ip and port.
	process.read_config()

	// A go routine to receive user commands
	// and perform a unicast send to the appropriate process.
	go process.recv_commands()

	// A go routine to listen for message and continously
	// deliver the message to the application.
	go process.recv_messages()

	// An infinite loop to prevent the program from terminating
	// except instructing the process to terminate.
	for {
	}
}
