package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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

func parseConfig(local_pid int, lines []string) LocalProcess {
	var min_delay int
	var max_delay int
	remote_processes := make(map[int]Server)

	_, err := fmt.Sscan(lines[0], &min_delay, &max_delay)

	if err != nil {
		log.Fatal(err)
	}

	for _, line := range lines[1:] {
		var pid int
		var ip string
		var port uint16

		_, err := fmt.Sscan(line, &pid, &ip, &port)

		if err != nil {
			log.Fatal(err)
		}

		server := Server{ip, port}
		remote_processes[pid] = server
	}

	return LocalProcess{min_delay, max_delay, local_pid, remote_processes}
}

func main() {
	lines := readConfig()
	local_process := parseConfig(1, lines)

	fmt.Printf("%+v\n", local_process)
}

/*
func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide port number")
		return
	}

	PORT := ":" + arguments[1]
	l, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	c, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		if strings.TrimSpace(string(netData)) == "STOP" {
			fmt.Println("Exiting TCP server!")
			return
		}

		fmt.Print("-> ", string(netData))
		t := time.Now()
		myTime := t.Format(time.RFC3339) + "\n"
		c.Write([]byte(myTime))
	}
}
*/
