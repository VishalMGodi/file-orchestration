package helpers

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Client struct {
	username string
	conn net.Conn
}

func NewClient(username string, conn net.Conn) *Client {
	return &Client{
		username: username,
		conn: conn,
	}
}

func (c *Client) UploadFile(filesize string, filename string) {
	c.conn.Write([]byte("Ready to receive file\n"))

	current_dir, err := os.Getwd()
	if err != nil {
		log.Fatal("could not find current dir")
	}
	storage_path := filepath.Join(current_dir, "server_storage")
	storage_path = filepath.Join(storage_path, c.username)
	if _, err := os.Stat(storage_path); os.IsNotExist(err) {
		err = os.MkdirAll(storage_path, 0755)
		if err != nil {
			c.conn.Write([]byte("ERROR: Unable to create directory for client\n"))
			return
		}
	}

	client_file_path := filepath.Join(storage_path, filename)
	file, err := os.Create(client_file_path)
	if err != nil {
		c.conn.Write([]byte("ERROR: Unable to create file\n"))
		return
	}

	defer file.Close()

	buffer := make([]byte, BUFFER_SIZE)
	fileSize, err := strconv.Atoi(filesize)
	// fmt.Println(filesize, filename)
	// fmt.Println(fileSize)
	if err != nil {
		c.conn.Write([]byte("ERROR: Filesize not in correct format\n"))
		return
	}
	for fileSize > 0 {
		n, err := c.conn.Read(buffer)
		if err != nil || n == 0 {
			break
		}
		file.Write(buffer[:n])
		fileSize -= n
	}

	fmt.Println("Succesful upload")
	c.conn.Write([]byte("SUCCESSFUL UPLOAD\n"))
}

func (c *Client) DownloadFile(filename string) {
	current_dir, err := os.Getwd()
	if err != nil {
		log.Fatal("could not find current dir")
	}
	storage_path := filepath.Join(current_dir, "server_storage")
	storage_path = filepath.Join(storage_path, c.username)
	if _, err := os.Stat(storage_path); os.IsNotExist(err) {
		err = os.MkdirAll(storage_path, 0755)
		if err != nil {
			c.conn.Write([]byte("ERROR: Unable to create directory for client\n"))
			return
		}
	}

	client_file_path := filepath.Join(storage_path, filename)
	file, err := os.Open(client_file_path)
	if err != nil {
		c.conn.Write([]byte("File not found"))
		return
	}

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	c.conn.Write([]byte(strconv.Itoa(int(fileSize))+"\n"))

	buffer := make([]byte, BUFFER_SIZE)
	for fileSize > 0 {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("ERROR: Reading file", err)
			return
		}
		if n == 0 {
			break
		}
		c.conn.Write(buffer[:n])
	}
	fmt.Println("File sent successfully")
}

func (c *Client) PreviewFile(filename string) {
	current_dir, err := os.Getwd()
	if err != nil {
		log.Fatal("could not find current dir")
	}
	storage_path := filepath.Join(current_dir, "server_storage")
	storage_path = filepath.Join(storage_path, c.username)
	if _, err := os.Stat(storage_path); os.IsNotExist(err) {
		err = os.MkdirAll(storage_path, 0755)
		if err != nil {
			c.conn.Write([]byte("ERROR: Unable to create directory for client\n"))
			return
		}
	}

	client_file_path := filepath.Join(storage_path, filename)
	file, err := os.Open(client_file_path)
	if err != nil {
		c.conn.Write([]byte("File not found"))
		return
	}

	buffer := make([]byte, BUFFER_SIZE)
	_, err = file.Read(buffer)

	c.conn.Write([]byte(string(buffer)+"\n"))
}

func (c *Client) DeleteFile(filename string) {
	current_dir, err := os.Getwd()
	if err != nil {
		log.Fatal("could not find current dir")
	}
	storage_path := filepath.Join(current_dir, "server_storage")
	storage_path = filepath.Join(storage_path, c.username)
	if _, err := os.Stat(storage_path); os.IsNotExist(err) {
		err = os.MkdirAll(storage_path, 0755)
		if err != nil {
			c.conn.Write([]byte("ERROR: Unable to create directory for client\n"))
			return
		}
	}

	client_file_path := filepath.Join(storage_path, filename)
	err = os.Remove(client_file_path)
	if err != nil {
		c.conn.Write([]byte("File not found\n"))
		return
	}

	c.conn.Write([]byte("File deleted successfully\n"))
}

func (c *Client) ListFiles() {
	current_dir, err := os.Getwd()
	if err != nil {
		log.Fatal("could not find current dir")
	}
	storage_path := filepath.Join(current_dir, "server_storage")
	storage_path = filepath.Join(storage_path, c.username)

	var allFiles []string
	files, err := os.ReadDir(storage_path)
	if err != nil {
		c.conn.Write([]byte("No username\n"))
	}

	for _, file := range files{
		allFiles = append(allFiles, file.Name())
	}

	listFiles := strings.Join(allFiles, "$") + "\n"
	c.conn.Write([]byte(listFiles))
}