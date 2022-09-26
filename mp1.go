package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

// This would be the structure representing a process in our program.
type Process struct {
	ip        string
	port      int
	min_delay int
	max_delay int
	// Simulated process ID of this process
	pid int
	// Maps remote process IDs to IPs and ports
	remote_processes map[int]Process
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
	process.remote_processes = make(map[int]Process)
	for {
		var (
			pid  int
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

func (process *Process) get_dailer() (dialer *net.Dialer) {
	fmt.Println("dailer", process.ip, process.port)
	dialer = &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP:   net.ParseIP(process.ip),
			Port: process.port,
		},
	}
	return dialer
}

func (process *Process) unicast_send(destination string, message []byte) {
	conn, err := net.Dial("tcp", destination)
	if err != nil {
		panic(err)
	}
	conn.Write([]byte(string(rune(process.pid))))
	conn.Write(message)
	conn.Close()
}

func (process *Process) unicast_recv(source net.Conn, msg []byte) {
	source.Read(msg)
	pid := msg[:1]
	message := msg[1:]
	fmt.Printf("Received %s from %d, system time is %v\n", message, pid, time.Now())
}

func main() {
	pid, _ := strconv.Atoi(os.Args[1])

	// Entry point, a new process is created and it reads
	// the config file to learn about other processes.
	process := Process{pid: pid}
	process.read_config()

	// After creating a map of the other process it sets up the
	// tcp client.
	port := strconv.Itoa(process.port)

	ln, err := net.Listen("tcp", process.ip+":"+port)
	_ = ln

	if err != nil {
		panic(err)
	}

	// From there the program loops indefinitely
	// Sending and receiving messages (to/from) to other processes.
	// process.unicast_send("127.0.0.1:9024", []byte("hello"))

	// process.unicast_send("127.0.0.1:8001", []byte("hello"))
	/*for {
		// TODO: Unicast send should go here

		// Unicast recv
		//source, _ := ln.Accept()
		// msg := make([]byte, 2048)
		// go process.unicast_recv(source, msg)
	} */
}
