package utils

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHConfig SSH连接配置
type SSHConfig struct {
	Host        string        // SSH 服务器地址
	Port        int          // SSH 服务器端口
	User        string        // 用户名
	Password    string        // 密码
	Timeout     time.Duration // 连接超时时间
	RetryTimes  int          // 重试次数
	RetryDelay  time.Duration // 重试间隔
}

// NewSSHClient 创建新的SSH客户端
//
// 参数:
//   - host: SSH 服务器地址
//   - port: SSH 服务器端口
//   - user: 用户名
//   - password: 密码
//
// 返回值:
//   - *ssh.Client: SSH客户端实例
//   - error: 连接过程中的错误，如果成功则返回 nil
func NewSSHClient(host string, port int, user string, password string) (*ssh.Client, error) {
	config := &SSHConfig{
		Host:        host,
		Port:        port,
		User:        user,
		Password:    password,
		Timeout:     30 * time.Second,
		RetryTimes:  3,
		RetryDelay:  5 * time.Second,
	}
	return config.Connect()
}

// Connect 根据配置建立SSH连接
func (c *SSHConfig) Connect() (*ssh.Client, error) {
	if c.Host == "" || c.Port <= 0 || c.Port > 65535 {
		return nil, fmt.Errorf("无效的主机地址或端口: %s:%d", c.Host, c.Port)
	}

	if c.User == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}

	sshConfig := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:        c.Timeout,
	}

	var client *ssh.Client
	var err error
	
	for i := 0; i <= c.RetryTimes; i++ {
		if i > 0 {
			time.Sleep(c.RetryDelay)
		}
		
		client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), sshConfig)
		if err == nil {
			return client, nil
		}
	}

	return nil, fmt.Errorf("SSH 连接失败 (已重试 %d 次): %v", c.RetryTimes, err)
}
