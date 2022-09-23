package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func unicast_send(destination string, message []byte) {
	conn, _ := net.Dial("tcp", destination)
	conn.Write(message)
	conn.Close()
}

func delay_and_send(min_delay, max_delay int, dest string, msg []byte) {
	duration := time.Duration(rand.Intn(max_delay) + min_delay)
	time.Sleep(duration)
	unicast_send(dest, msg)
}

func read_line(reader io.Reader, lineNumber int) (text string, err error) {
	scanner := bufio.NewScanner(reader)
	line := 0
	for scanner.Scan() {
		line++
		if line == lineNumber {
			return scanner.Text(), scanner.Err()
		}
	}
	return "", io.EOF
}

func main() {
	// Gets a command from user of the form send [Integer] [String]
	// Where Integer is the process ID and String is the message to send.
	fmt.Print("please input a command \n>")

	// Some error checking might be useful here.
	command, _ := bufio.NewReader(os.Stdin).ReadString('\n')

	// Trims new line from command.
	command = strings.TrimSuffix(command, "\n")

	// Splits the command into elements of an array.
	arr := strings.Split(string(command), " ")
	message := arr[3]
	destination := arr[2]

	// Should error check here to see if file exists.
	reader, _ := os.Open("./config.txt")

	// Reads a line from the config file, getting ip:port
	// of receiving process.
	line, _ := strconv.Atoi(destination)

	// Should error check here to see if user typed in a valid line.
	destination, _ = read_line(reader, line)
	delay, _ := read_line(reader, 1)

	arr = strings.Split(destination, " ")

	arr2 := strings.Split(delay, " ")

	ip, port := arr[1], arr[2]

	min_delay, _ := strconv.Atoi(arr2[0])
	max_delay, _ := strconv.Atoi(arr2[1])

	// Sends the message to the intended process.
	go delay_and_send(min_delay, max_delay, ip+":"+port, []byte(message))

	// Message successfully sent output.
	fmt.Printf("Sent %s to %s, system time is %v\n", message, destination, time.Now())
}
