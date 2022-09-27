# MP1_TCP-IP-Network-Simulation
 Network simulation of TCP/IP server with delays.
 
 ## How to run
 
 Just make sure the config file has a process id, port and ip.
 
 You need two processes, one that has `pid1` and `pid2`
 
 From there just type `go run mp1.go [pid1]` where pid is the process id.
 
 Alternatively, you can build the executable.
 
 Get the other process running with `go run mp1.go [pid2]`
 
 From there you should be able to send and recieve commands from other processes.
