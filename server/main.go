package main

import (
	"fmt"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 读取第一个字节来判断是HTTP还是SOCKS5
	firstByte := make([]byte, 1)
	_, err := conn.Read(firstByte)
	if err != nil {
		log.Printf("Failed to read first byte: %v", err)
		return
	}

	// 将读取的字节放回连接
	fullConn := &fullConn{
		Conn:    conn,
		peeked:  firstByte,
		peekedN: 1,
	}

	if firstByte[0] == 0x05 {
		// SOCKS5
		handleSocks5(fullConn)
	} else {
		// 假设是HTTP
		handleHTTP(fullConn)
	}
}

func main2() {
	listener, err := net.Listen("tcp", ":1080")
	if err != nil {
		log.Fatalf("Failed to listen on port 1080: %v", err)
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 1080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func main() {
	init_docker()
	n, err := getDockerNet()
	if err != nil {
		log.Printf("get docker net error: %v", err)
	}
	log.Printf("docker net: %s", n)
}
