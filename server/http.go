package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func handleHTTP(conn net.Conn) {
	reader := bufio.NewReader(conn)
	_, err := http.ReadRequest(reader)
	if err != nil {
		log.Printf("Error reading HTTP request: %v", err)
		return
	}

	var resp *http.Response
	var content string
	h, err := getK8sHosts()
	if err != nil {
		resp = &http.Response{
			StatusCode: 500,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
		}
		content = err.Error()
	} else {
		resp = &http.Response{
			StatusCode: 200,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
		}
		content = h
	}

	resp.Header.Set("Content-Type", "text/plain")

	resp.Body = io.NopCloser(strings.NewReader(content))

	if err := resp.Write(conn); err != nil {
		log.Printf("Error writing HTTP response: %v", err)
	}
}
