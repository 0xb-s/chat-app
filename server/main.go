package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

const (
	PORT = "6000"
)

type Client struct {
	conn     net.Conn
	username string
}

type Server struct {
	clients    map[net.Conn]Client
	broadcast  chan string
	register   chan Client
	unregister chan net.Conn
	mu         sync.Mutex
}

func NewServer() *Server {
	return &Server{
		clients:    make(map[net.Conn]Client),
		broadcast:  make(chan string),
		register:   make(chan Client),
		unregister: make(chan net.Conn),
	}
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.conn] = client
			s.mu.Unlock()
			message := fmt.Sprintf("%s has joined the chat.", client.username)
			fmt.Println(message) // Log to server terminal
			s.broadcast <- message
		case conn := <-s.unregister:
			s.mu.Lock()
			client, ok := s.clients[conn]
			if ok {
				delete(s.clients, conn)
				message := fmt.Sprintf("%s has left the chat.", client.username)
				fmt.Println(message) // Log to server terminal
				s.broadcast <- message
			}
			s.mu.Unlock()
			conn.Close()
		case message := <-s.broadcast:
			s.mu.Lock()
			for _, client := range s.clients {
				_, err := fmt.Fprintln(client.conn, message)
				if err != nil {
					fmt.Printf("Error sending message to %s: %v\n", client.username, err)
				//todo better handling here
				}
			}
			s.mu.Unlock()
		}
	}
}

func handleConnection(server *Server, conn net.Conn) {
	defer func() {
		server.unregister <- conn
	}()

	reader := bufio.NewReader(conn)

	// Read username from client
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading username: %v\n", err)
		return
	}
	username = strings.TrimSpace(username)
	if username == "" {
		username = "Anonymous"
	}

	client := Client{
		conn:     conn,
		username: username,
	}
	server.register <- client

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading message from %s: %v\n", client.username, err)
			break
		}
		message = strings.TrimSpace(message)
		if message != "" {
			formattedMsg := fmt.Sprintf("%s: %s", client.username, message)
			fmt.Println(formattedMsg) // Log to server terminal
			server.broadcast <- formattedMsg
		}
	}
}

func main() {
	server := NewServer()
	go server.Run()

	listener, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Chat server started on port", PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Printf("New connection from %s\n", conn.RemoteAddr().String())
		go handleConnection(server, conn)
	}
}
