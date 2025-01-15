package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"platform-vpn/pkgs/vpn"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	var err error
	// 接受标准输入以获取用户名、密码和服务器
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("请输入服务器地址: ")
	server, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("读取输入时出错:", err)
		return
	}
	server = strings.TrimSpace(server)

	fmt.Print("请输入服务器端口[22]: ")
	port, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("读取输入时出错:", err)
		return
	}
	port = strings.TrimSpace(port)
	var portInt int
	if port == "" {
		portInt = 22
	} else {
		portInt, err = strconv.Atoi(port)
		if err != nil {
			fmt.Println("端口号无效:", err)
			return
		}
	}

	fmt.Print("请输入用户名[root]: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("读取输入时出错:", err)
		return
	}
	username = strings.TrimSpace(username)
	if username == "" {
		username = "root"
	}

	fmt.Print("请输入密码: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("读取输入时出错:", err)
		return
	}
	password = strings.TrimSpace(password)

	// 使用输入的用户名、密码和服务器进行 VPN 连接
	err = vpn.StartVPN(username, password, server, portInt, 1*time.Minute)
	if err != nil {
		fmt.Printf("连接 VPN 服务器[%s]失败: %v\n", server, err)
		return
	}

	fmt.Printf("连接 VPN 服务器[%s]成功。\n", server)

	// 添加信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("接收到退出信号，正在清理...")
	// 在这里添加清理代码，例如停止 VPN
	os.Exit(0)
}
