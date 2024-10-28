package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

var getHosts func() (string, error)
var getNets func() (string, error)

func init() {
	err := init_k8s()
	if err != nil {
		log.Printf("connect k8s error: %v", err)
		log.Printf("read hosts from docker")
		init_docker()
		getHosts = getDockerHosts
		getNets = getDockerNet
	} else {
		getHosts = getK8sHosts
		getNets = getK8sNet
	}
}

func handleHTTP(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Printf("Error reading HTTP request: %v", err)
		return
	}

	var resp *http.Response
	switch req.URL.Path {
	case "/hosts":
		h, err := getHosts()
		if err != nil {
			resp = createResponse(500, err.Error())
		} else {
			resp = createResponse(200, h)
		}
	case "/nets":
		n, err := getNets()
		if err != nil {
			resp = createResponse(500, err.Error())
		} else {
			resp = createResponse(200, n)
		}
	default:
		resp = createResponse(404, "Not Found")
	}

	if err := resp.Write(conn); err != nil {
		log.Printf("Error writing HTTP response: %v", err)
	}
}

func createResponse(statusCode int, content string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{"Content-Type": {"text/plain"}},
		Body:       io.NopCloser(strings.NewReader(content)),
	}
}
