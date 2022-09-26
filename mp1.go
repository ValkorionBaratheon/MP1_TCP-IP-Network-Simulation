package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
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

func (process *Process) read_config() {
	// Gets the file, this can be it's own function.
	file, err := os.Open("./config.txt")
	if err != nil {
		fmt.Println("config file not found")
		os.Exit(1)
	}

	// Gets the min and max delay, this can be it's own function.
	fmt.Fscanln(file, &process.min_delay, &process.max_delay)
	// fmt.Printf("%d: %d, %d\n", n, min_delay, max_delay)
	process.remote_processes = make(map[string]Process)

	// Fills up the process map, this can be it's own function.
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
		// No need to set the remote process.
		if pid == process.pid {
			process.ip = ip
			process.port = port
		} else {
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
	// duration := time.Duration(rand.Intn(max_delay) + min_delay)
	// time.Sleep(duration)
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2, // needs to be local process port.
		},
	}
	conn, _ := dialer.Dial("tcp", destination)
	conn.Write(message)
	conn.Close()
}

func (process *Process) unicast_recv(source net.Conn, msg []byte) {
	source.Read(msg)
	address := source.RemoteAddr().String()
	pid := process.remote_processes[address].pid
	fmt.Printf("Received %s from %d, system time is %v\n", msg, pid, time.Now())
}

func readConfig() []string {
	file, err := os.Open("./config.txt")

	if err != nil {
		log.Fatal(err)
	}

	out := make([]string, 0)

	// Release file handle when this function returns
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	err = scanner.Err()

	if err != nil {
		log.Fatal(err)
	}

	return out
}

func main() {
	pid, _ := strconv.Atoi(os.Args[1])
	process := Process{pid: pid}
	process.read_config()
	port := strconv.Itoa(process.port)

	ln, err := net.Listen("tcp", process.ip+":"+port)
	if err != nil {
		panic(err)
	}

	for {
		// fmt.Print("please input a command \n>>")

		// TODO: Unicast send should go here

		// Unicast recv
		source, _ := ln.Accept()
		msg := make([]byte, 2048)
		go process.unicast_recv(source, msg)
	}
}
