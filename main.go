package main

import (
	"fmt"
	"net"

	"github.com/Amul-Thantharate/KafkaConnect/database"
	"github.com/Amul-Thantharate/KafkaConnect/kafka"
	"github.com/Amul-Thantharate/KafkaConnect/server"
)

func main() {
	// Initialize Database and Kafka
	database.ConnectDatabase()
	kafka.InitKafka()

	// Start TCP Server
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Chat Server started on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Client disconnected")
			return
		}
		message := string(buf[:n])
		server.HandleCommand(conn, message)
	}
}
