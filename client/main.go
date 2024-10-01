package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	HOST = "localhost"
	PORT = "8888"
	TYPE = "tcp"
	BUFFER_SIZE = 1024
)

func HandleAuth(conn net.Conn, username string, password string) bool {
	conn.Write([]byte("AUTH "+username+":"+password+"\n"))

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil{
		fmt.Println("Some error occured")
		return false
	} 
	line = strings.TrimSuffix(line, "\n")

	if line == "Unauthorized" {
		fmt.Println("Unauthorized: Wrong credentials")
		return false
	} else if line == "Authorized" {
		fmt.Println("You have logged in")
	}
	return true
}

func UploadFile(conn net.Conn, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("ERROR: Unable to open file", filename)
		return
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	// fmt.Println(fileSize)
	var commandFields []string
	commandFields = append(commandFields, "UPLOAD")
	commandFields = append(commandFields, strconv.Itoa(int(fileSize)))
	commandFields = append(commandFields, filename)
	
	commandBuffer := strings.Join(commandFields, ",") + "\n"
	// fmt.Println(commandBuffer)
	conn.Write([]byte(commandBuffer))

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil{
		fmt.Println("Some error occured")
		return
	} 
	line = strings.TrimSuffix(line, "\n")
	if line == "Ready to receive file"{
		buffer := make([]byte, BUFFER_SIZE)
		for {
			n, err := file.Read(buffer)
			if err != nil && err != io.EOF {
				fmt.Println("ERROR: Reading file", err)
				return
			}
			if n == 0 {
				break
			}
			conn.Write(buffer[:n])
		}
	}

	success, err := reader.ReadString('\n')
	if err != nil{
		fmt.Println("Some error occured")
		return
	} 
	fmt.Println(success)
}

func DownloadFile(conn net.Conn, filename string) {
	conn.Write([]byte("DOWNLOAD,"+filename+"\n"))
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("could not create file")
	}
	defer file.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil{
		fmt.Println("Some error occured")
		return
	} 
	line = strings.TrimSuffix(line, "\n")
	if line == "File not found"{
		fmt.Println(line)
		return
	} else{
		fileSize, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println("Filesize not in right format")
		}

		buffer := make([]byte, BUFFER_SIZE)
		for fileSize > 0 {
			n, err := conn.Read(buffer)
			if err != nil || n == 0 {
				break
			}
			file.Write(buffer[:n])
			fileSize -= n
		}
	}
	fmt.Println("File downloaded successfully")
}

func DeleteFile(conn net.Conn, filename string) {
	conn.Write([]byte("DELETE,"+filename+"\n"))

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil{
		fmt.Println("Some error occured")
		return
	} 
	line = strings.TrimSuffix(line, "\n")
	fmt.Println(line)
}

func PreviewFile(conn net.Conn, filename string) {
	conn.Write([]byte("PREVIEW,"+filename+"\n"))

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil{
		fmt.Println("Some error occured")
		return
	} 
	line = strings.TrimSuffix(line, "\n")
	fmt.Println(line)
}

func ListFiles(conn net.Conn) {
	conn.Write([]byte("LIST\n"))

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil{
		fmt.Println("Some error occured")
		return
	} 
	line = strings.TrimSuffix(line, "\n")
	files := strings.Split(line, "$")
	fmt.Println("Your files:")
	for _, file := range files{
		fmt.Println(file)
	}
}

func main() {
	conn, err := net.Dial(TYPE, ":"+PORT)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	username, password := os.Args[1], os.Args[2]
	fmt.Println(username, password)
	isAuth := HandleAuth(conn, username, password)
	if !isAuth {
		conn.Close()
		return
	}
	for {
		var command string
		fmt.Println("Available services:\n1. Upload a file\n2. Download a file\n3. List all your files\n4. Delete a file\n5. Preview a file")
		fmt.Println("Syntax:\nUPLOAD,<filename.txt>\nDOWNLOAD,<filename.txt>\nLIST\nDELETE,<filename.txt>\nPREVIEW,<filename.txt>")
		fmt.Println("\nEnter command: ")
		fmt.Scanln(&command)
		fields := strings.Split(command, ",")

		if fields[0] == "UPLOAD" {
			UploadFile(conn, fields[1])
		} else if fields[0] == "LIST" {
			ListFiles(conn)
		} else if fields[0] == "DOWNLOAD"{
			DownloadFile(conn, fields[1])
		} else if fields[0] == "DELETE" {
			DeleteFile(conn, fields[1])
		} else if fields[0] == "PREVIEW" {
			PreviewFile(conn, fields[1])
		} else {
			// Read server response
			buffer := make([]byte, BUFFER_SIZE)
			for {
				n, err := conn.Read(buffer)
				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Println("Error reading:", err)
					return
				}
				fmt.Print(string(buffer[:n]))
				if n < BUFFER_SIZE {
					break
				}
			}
		}
	}

}