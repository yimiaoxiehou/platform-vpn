package main

import (
	"context"
	"fmt"
	"os/exec"
	"platform-vpn/pkgs/k3s"
	"platform-vpn/pkgs/log"
	"platform-vpn/pkgs/vpn"
	"runtime"
	"time"
)

// App struct
type App struct {
	ctx context.Context
}

// AppService struct
type AppService struct {
	Name  string
	IP    string
	Ports []int32
}

type AppNsService struct {
	Namespace string
	Services  []AppService
}

// AppResponse struct
type AppResponse[T any] struct {
	OK      bool
	Message string
	Data    T
}

// VPNConfig struct
type VPNConfig struct {
	Server          string
	User            string
	Port            int
	Password        string
	RefreshInterval int
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) StartVPN(config *VPNConfig) (bool, string) {

	err := vpn.StartVPN(config.User, config.Password, config.Server, config.Port, time.Duration(config.RefreshInterval)*time.Minute)
	if err != nil {
		return false, fmt.Sprintf("连接 VPN 服务器[%s]失败: %v", config.Server, err)
	}

	return true, fmt.Sprintf("连接 VPN 服务器[%s]成功。", config.Server)
}

func (a *App) StopVPN() error {
	err := vpn.StopVPN()
	if err != nil {
		log.Error(fmt.Sprintf("停止VPN出错: %v", err))
	} else {
		log.Info("停止VPN成功")
	}
	return err
}

func (a *App) RefreshHosts() error {
	return vpn.UpdateHosts()
}

func (a *App) OpenHosts() error {

	var hostsPath string
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		hostsPath = `C:\Windows\System32\drivers\etc\hosts`
		cmd = exec.Command("notepad.exe", hostsPath)
	case "darwin":
		hostsPath = "/etc/hosts"
		cmd = exec.Command("open", "-e", hostsPath)
	case "linux":
		hostsPath = "/etc/hosts"
		// 根据不同的桌面环境，选择合适的文本编辑器
		cmd = exec.Command("gedit", hostsPath) // 或者使用 "nano", "vim" 等
	default:
		return fmt.Errorf("不支持的操作系统")
	}

	return cmd.Start()
}

func (a *App) GetServices() ([]AppNsService, error) {

	servicess := make([]AppNsService, 0)

	svcss, err := k3s.GetServices()
	if err != nil {
		return servicess, err
	}
	for ns, svcs := range svcss {
		services := make([]AppService, 0)
		for _, svc := range svcs {
			ports := make([]int32, 0)
			for _, port := range svc.Spec.Ports {
				ports = append(ports, port.Port)
			}
			services = append(services, AppService{
				Name:  svc.Name,
				IP:    svc.Spec.ClusterIP,
				Ports: ports,
			})
		}
		servicess = append(servicess, AppNsService{
			Namespace: ns,
			Services:  services,
		})
	}

	return servicess, nil
}

func (a *App) GetLogs() ([]*log.LogItem, error) {
	return log.GetLogs(), nil
}
