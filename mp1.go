package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// This would be the structure representing a process in our program.
type Process struct {
	ip        string
	port      int
	min_delay int
	max_delay int
	// Simulated process ID of this process
	pid int32
	// Maps remote process IDs to IPs and ports
	remote_processes map[int32]Process
	exit             chan (bool)
}

func (process *Process) get_delay() (int, int) {
	return process.min_delay, process.max_delay
}

func get_config_file() (file *os.File) {
	file, err := os.Open("./config.txt")
	if err != nil {
		fmt.Println("config.txt file not found")
		os.Exit(1)
	}
	return file
}

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

func (process *Process) read_config() {
	file := get_config_file()
	fmt.Fscanln(file, &process.min_delay, &process.max_delay)
	process.read_remote_processes(file)
}

func (process *Process) unicast_send(destination string, message []byte) {
	id, _ := strconv.ParseInt(destination, 10, 32)
	pid := int32(id)
	ip := process.remote_processes[pid].ip
	port := strconv.Itoa(process.remote_processes[pid].port)
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		panic(err)
	}
	// Stimulating delay.
	min_delay, max_delay := process.get_delay()
	duration := time.Duration(rand.Intn(max_delay) + min_delay)
	time.Sleep(duration)

	binary.Write(conn, binary.BigEndian, process.pid)
	conn.Write(message)
	conn.Close()
	fmt.Printf("Sent \"%s\" to process %d, system time is %v\n", message, pid, time.Now())
}

func (process *Process) unicast_recv(source net.Conn, msg []byte) {
	var pid int32
	binary.Read(source, binary.BigEndian, &pid)
	source.Read(msg)
	fmt.Printf("Received \"%s\" from %d, system time is %v\n>> ", msg, pid, time.Now())
}

func (process *Process) get_command() (string, string) {
	command, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	command = strings.TrimSuffix(command, "\n")
	if command == "q" {
		fmt.Println("closing TCP server...")
		os.Exit(0)
	}
	commandArray := strings.Split(string(command), " ")
	fmt.Println(commandArray)
	message := strings.Join(commandArray[2:], " ")
	destination := commandArray[1]
	return message, destination
}

func (process *Process) recv_commands() {
	for {
		fmt.Printf("Please input a command or 'q' to quit \n>> ")
		message, destination := process.get_command()
		process.unicast_send(destination, []byte(message))
	}
}

func (process *Process) recv_messages() {
	port := strconv.Itoa(process.port)
	ln, _ := net.Listen("tcp", process.ip+":"+port)
	for {
		//fmt.Printf("please input a command: send [Integer] [String] \n>> ")
		// message, destination := process_command()
		// go process.unicast_send(destination, []byte(message))
		source, _ := ln.Accept()
		msg := make([]byte, 2048)
		go process.unicast_recv(source, msg)
	}
}

func main() {
	id, _ := strconv.ParseInt(os.Args[1], 10, 32)
	pid := int32(id)
	// Entry point, a new process is created and it reads
	// the config file to learn about other processes.
	process := Process{pid: pid}
	process.read_config()
	go process.recv_commands()
	go process.recv_messages()
	for {
	}
}
