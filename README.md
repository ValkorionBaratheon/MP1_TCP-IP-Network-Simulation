# Machine Problem 1: TCP/IP Network Simulation

This is Machine Problem 1, which simulates delays in a TCP/IP server-client relationship.

This program documentation consists of the following sections:

1. Program Design
2. Code Flow of mp1.go
3. How to Run

## Program Design

### File Structure

This program consists of two files:

- `config.txt`
- `mp1.go` (later built into an executable called `mp1.exe` or `mp1`)

### Overview of config.txt

This file contains the minimum and maximum delays (in milliseconds), as well as two processes (at the time of writing), each consisting of a process ID (PID), IP address, and port number. The simulation is designed to work with any number of processes, even just one, so you can add lines to this config to simulate more processes. The IP and port in each entry are real and tell the program where to connect when trying to talk to another process.

### Overview of mp1.go

This file is responsible for reading the contents of `config.txt`. With command-line input, it starts processes that constantly wait for messages to be received from another process. These messages are also generated by user input.

#### Process Type and Reading config.txt

A key component of this file is the creation of the program state. The program has one object of type `Process` that contains all state information. Its fields consist of IP address, port number, minimum delay, maximum delay, and PID. It also maps remote process IDs to IPs and ports. After the program reads and parses `config.txt`, the values are stored according to the Process type.

```go
type Process struct {
	min_delay int
	max_delay int
	pid int
	remote_processes map[int]Server
	message_queue chan Message
	r *rand.Rand
}
```

#### Sending Messages

Utilizing goroutines, the program handles sending messages by implementing unicast send functionality. A specific process will take a command in the terminal. This command, `send` will take `destination` and `message` parameters, and will send a message to the destination process, with a system message outputting the time stamp. The message is not sent immediately. To simulate TCP delay and possible out-of-order message delivery, a random delay† is computed and assigned to the message, and the message is queued in a circular buffer of size 256 (implemented by a go channel). A separate goroutine blocks until it reads a queued message from this go channel. If the message is ready to be sent (i.e., the message's delay has passed), the message is sent to the remote process. If the message is not ready to be sent, it is passed to the other end of the go channel. This opens the possibility for messages to be sent out of order. If the buffer has two messages, and Message 1 with a delay of 500 ms is in front of Message B with a delay of 200 ms, Message 1 will be inserted behind Message 2, and Message 2 will be sent before Message 1.

_† A random delay is not computed if the message's sender matches the destination. If a process tries to send a message to itself, there should not be any simulated delay because the message does not go through the network._

#### Receiving Messages

Utilizing goroutines, the program handles receiving messages by implementing unicast receive functionality. Running the program creates a process with a PID (passed in through args). This process opens a TCP socket and listens for incoming connections from other processes. The code that does this blocks until another process initiates a connection, so it runs in a goroutine. When another process connects, it immediately sends its message and disconnects. The message and sender are printed to the console with a timestamp.

#### Message Protocol

Processes communicate with each other using a simple protocol over TCP. A message consists of an 8 byte header followed by a variable-length ASCII text field. The header consists of two signed 32 bit integers, sent in Big-Endian order. The first integer is the size of the text field in bytes, and the second integer is the process ID of the sender. The ASCII text is sent last, and this field has the size given by the first byte of the header.

## Code Flow

### Process Type

Please see the earlier section under `Overview of mp1.go`. You will see `type Process struct` in the code.

### Reading config.txt

`read_config()` and `parse_config()` deal with reading `config.txt` and extracting the relevant information, mapping the remote processes.

### Sending Messages

Three functions, `process_message_queue()`, `queue_message()`, and `send_message()` work together to send messages to another process. `process_message_queue()` runs as a goroutine, listening for incoming messages on the message queue channel. When it receives a message that is ready to be sent, it calls `send_message()` to send it. `queue_message()` is used to add a message to the queue. This function uses the `min_delay` and `max_delay` from `config.txt` to compute a random delay for the message before putting it in the queue. `queue_message()` is called as a goroutine in case the message queue is full and it blocks. We don't want to block the main thread listening for user commands, so `queue_message()` should be run as a goroutine.

### Receiving Messages

Two functions, `listen_for_incoming()` and `receive_incoming()`, work together to receive messages from another process. The first function is constantly listening on its given port. It calls upon `receive_incoming()` to receive a connection and process the data sent by the client process. We expect the client process to close the connection when it has sent a message, but in case it does not or in case something else connects and sends garbage, the server will close the connection before `receive_incoming()` returns.

### Main() - Putting it All Together

`config.txt` is read, parsed, and processes mapped accordingly. Then, goroutines are used to send and receive messages until exited by user-input.

## How To Run

1. Open a terminal.
2. In the terminal, navigate to the project folder.
3. Input the following command to build the executable for your machine: `go build mp1.go`.
4. Input the following command: `./mp1 <PID>` (`mp1.exe <PID>` on Windows). PID should be a single integer corresponding to the desired process you wish to start (1 or 2).
5. Open a second terminal and navigate to the project folder.
6. Repeat step 4 with a different PID.
7. In each terminal, input the following command: `send <PID> <message>`. The message can be any string.
8. Type `q` to quit the process. Must do this for each terminal in which a process is running.

_Note: The processes will display send and receive messages with time stamps in their respective terminals._
