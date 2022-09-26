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
	remote_processes map[string]Process
}

func (process *Process) get_delay() (int, int) {
	return process.min_delay, process.max_delay
}

func get_config_file() {
	// TODO?

}

func (process *Process) read_min_max_delay() {
	// TODO?

}

func (process *Process) read_remote_processes() {
	// TODO ?

}

func (process *Process) read_config() {
	// This can be get_config_file
	file, err := os.Open("./config.txt")
	if err != nil {
		fmt.Println("config file not found")
		os.Exit(1)
	}
	// This can be get min_max_delay
	fmt.Fscanln(file, &process.min_delay, &process.max_delay)

	// This can be read_remote_processes
	// Fills up the process map, this can be it's own function.
	process.remote_processes = make(map[string]Process)
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
			port := strconv.Itoa(port)
			process.remote_processes[ip+":"+port] = remote_process
		}
	}
}

func (process *Process) unicast_send(destination string, message []byte) {
	// TODO:
}

func (process *Process) unicast_recv(source net.Conn, msg []byte) {
	source.Read(msg)
	address := source.RemoteAddr().String()
	pid := process.remote_processes[address].pid
	fmt.Printf("Received %s from %d, system time is %v\n", msg, pid, time.Now())
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

	if err != nil {
		panic(err)
	}

	// From there the program loops indefinitely
	// Sending and receiving messages (to/from) to other processes.
	for {
		// TODO: Unicast send should go here

		// Unicast recv
		source, _ := ln.Accept()
		msg := make([]byte, 2048)
		go process.unicast_recv(source, msg)
	}
}
