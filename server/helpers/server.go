package helpers

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	HOST = "localhost"
	PORT = "8888"
	TYPE = "tcp"
	BUFFER_SIZE = 1024
)

type Server struct {
	clients map[Client]bool
	mu sync.Mutex
	signalSwitch chan struct{}
}

func NewServer() *Server {
	return &Server{
		clients: make(map[Client]bool),
		signalSwitch: make(chan struct{}),
	}
}

func (s *Server) AddClientConnection(c Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c] = true
}

func (s *Server) RemoveClientConnection(c Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, c)
	c.conn.Close()
}

func HandleAuth(conn net.Conn) (bool, string, error) {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, "", err
	}
	// fmt.Println(line)
	if !strings.HasPrefix(line, "AUTH ") {
		return false, "", errors.New("no basic auth provided")
	}

	auth := strings.TrimPrefix(line, "AUTH ")
	decodedAuth := strings.TrimSuffix(auth, "\n")

	authParts := strings.Split(string(decodedAuth), ":")
	if len(authParts) != 2 {
		return false, "", errors.New("invalid auth format")
	}

	username, password := authParts[0], authParts[1]
	filename := "passwords/passwords.txt"

	file, err := os.Open(filename)
	if err != nil {
		return false, "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			storedUsername := parts[0]
			storedPassword := parts[1]
			if username == storedUsername && password == storedPassword {
				return true, username, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, "", err
	}

	return false, "", errors.New("invalid credentials")
}

func (s *Server) Run() {
	listener, err := net.Listen(TYPE, ":"+PORT)
	if err != nil {
		fmt.Println("Error starting Server: ", err)
		return
	}

	defer listener.Close()

	fmt.Println("Server started on port", PORT)
	fmt.Println("Available services:\n1. Upload a file\n2. Download a file\n3. List all your files\n4. Delete a file\n5. Preview a file")


	// listen to client connections 
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}
		isAuth, username, err := HandleAuth(conn)
		// fmt.Println(isAuth, username, err)
		if !isAuth {
			conn.Write([]byte("Unauthorized"+"\n"))
			continue
		}
		conn.Write([]byte("Authorized"+"\n"))
		go s.HandleClient(conn, username)
	}
}

func (s *Server) HandleClient(conn net.Conn, username string) {
	fmt.Println("Connection accepted: ", username)
	client := NewClient(username, conn)
	s.AddClientConnection(*client)
	// fmt.Println(s.clients)
	defer s.RemoveClientConnection(*client)

	for {
		reader := bufio.NewReader(conn)
		line, err := reader.ReadString('\n')
		if err != nil{
			fmt.Println("Some error occured")
			break
		} 
		fmt.Println(line)
		line = strings.TrimSuffix(line, "\n")
		fields := strings.Split(line, ",")
		
		if len(fields) == 0 {
			conn.Write([]byte("Error: Incomplete Command\n"))
			continue
		}

		command := fields[0]
		// fmt.Println(command)
		switch command {
		case "LIST":
			client.ListFiles()
		case "UPLOAD":
			if len(fields) < 3{
				conn.Write([]byte("Error: Missing filename\n"))
				continue
			}
			client.UploadFile(fields[1], fields[2])
		case "DOWNLOAD":
			if len(fields) < 2{
				conn.Write([]byte("Error: Missing filename\n"))
				continue
			}
			client.DownloadFile(fields[1])
		case "DELETE":
			if len(fields) < 2{
				conn.Write([]byte("Error: Missing filename\n"))
				continue
			}
			client.DeleteFile(fields[1])
		case "PREVIEW":
			if len(fields) < 2{
				conn.Write([]byte("Error: Missing filename\n"))
				continue
			}
			client.PreviewFile(fields[1])
		default:
			conn.Write([]byte("Error: Invalid Command\n"))
		}
	}
}