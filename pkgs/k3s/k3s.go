package k3s

import (
	"encoding/json"
	"fmt"
	"platform-vpn/pkgs/utils"
	"strings"

	"golang.org/x/crypto/ssh"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

type RancherK3sConfig struct {
	WriteKubeconfigMode  string   `yaml:"write-kubeconfig-mode"`
	Disable              []string `yaml:"disable"`
	ClusterCIDR          string   `yaml:"cluster-cidr"`
	ServiceCIDR          string   `yaml:"service-cidr"`
	ClusterDNS           string   `yaml:"cluster-dns"`
	ServiceNodePortRange string   `yaml:"service-node-port-range"`
	ClusterInit          bool     `yaml:"cluster-init"`
}

type Client struct {
	SSHClient *ssh.Client
}

var _client Client

func NewClient(host string, port int, user string, password string) (*Client, error) {
	if _client.SSHClient != nil {
		return &_client, nil
	}
	sshClient, err := utils.NewSSHClient(host, port, user, password)
	if err != nil {
		fmt.Println("创建ssh连接失败")
		return &_client, err
	}
	_client.SSHClient = sshClient

	return &_client, nil
}

func (c *Client) remoteExec(cmd string) ([]byte, error) {
	session, err := c.SSHClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	output, err := session.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("执行命令失败: %v", err)
	}
	return output, nil
}

func (c *Client) GetCidrs() ([]string, error) {
	cidrs := make([]string, 0)
	var newConfig RancherK3sConfig
	output, err := c.remoteExec("cat /etc/rancher/k3s/config.yaml")
	clusterCidr := ""
	serviceCidr := ""
	if err != nil {
		var serviceCFG *utils.ServiceConfig
		output, err = c.remoteExec("cat /etc/systemd/system/k3s.service")
		if err != nil {
			return cidrs, err
		}
		serviceCFG, err = utils.ParseServiceFile(string(output))
		execStartCMD := serviceCFG.Service["ExecStart"]
		cmds := strings.Split(execStartCMD, " ")
		pureCmds := make([]string, 0)
		for _, cmd := range cmds {
			cmd = strings.TrimSpace(cmd)
			cmd = strings.Trim(cmd, "'")
			cmd = strings.Trim(cmd, "\"")
			cmd = strings.TrimSpace(cmd)
			if cmd != "" {
				pureCmds = append(pureCmds, cmd)
			}
		}
		cmds = pureCmds
		for i, cmd := range cmds {
			if cmd == "--cluster-cidr" && i < len(cmds)-1 {
				clusterCidr = cmds[i+1]
			}
			if cmd == "--service-cidr" && i < len(cmds)-1 {
				serviceCidr = cmds[i+1]
			}
		}
		if clusterCidr != "" {
			cidrs = append(cidrs, clusterCidr)
		}
		if serviceCidr != "" {
			cidrs = append(cidrs, serviceCidr)
		}
		return cidrs, nil
	}
	// 从 YAML 反序列化
	err = yaml.Unmarshal(output, &newConfig)
	if err != nil {
		fmt.Printf("反序列化失败: %v\n", err)
	}
	cidrs = append(cidrs, newConfig.ClusterCIDR)
	cidrs = append(cidrs, newConfig.ServiceCIDR)
	return cidrs, nil
}

func (c *Client) GetNsServices() (map[string][]corev1.Service, error) {
	nsSvcs := make(map[string][]corev1.Service)
	var serviceList corev1.ServiceList
	output, err := c.remoteExec("k3s kubectl get svc -A -o json")
	if err != nil {
		return nsSvcs, err
	}

	// 解析 JSON 数据
	if err := json.Unmarshal(output, &serviceList); err != nil {
		return nsSvcs, fmt.Errorf("解析服务列表失败: %v", err)
	}

	for _, svc := range serviceList.Items {
		ns := svc.GetNamespace()
		svcs := nsSvcs[ns]
		if svcs == nil {
			svcs = make([]corev1.Service, 0)
		}
		svcs = append(svcs, svc)
		nsSvcs[ns] = svcs
	}

	return nsSvcs, nil
}

func (c *Client) GetServiceHosts() (string, error) {
	nsSvcs, err := c.GetNsServices()
	if err != nil {
		return "", err
	}

	k8sHosts := utils.HOST_START

	nsHosts := make(map[string]map[string]string)
	for ns, svcs := range nsSvcs {
		hlSvc := make(map[string]corev1.Service)
		items := make([]corev1.Service, 0)

		hosts := make(map[string]string)

		for _, svc := range svcs {
			if svc.Spec.ClusterIP == "None" {
				hlSvc[svc.Name] = svc
			} else {
				items = append(items, svc)
			}
		}

		for _, svc := range items {
			if s, ok := hlSvc[svc.Name+"-hl"]; ok {
				hosts[s.Name] = svc.Spec.ClusterIP
			}
			hosts[svc.Name] = svc.Spec.ClusterIP
		}
		nsHosts[ns] = hosts
	}

	for ns, hosts := range nsHosts {
		for name, host := range hosts {
			k8sHosts += fmt.Sprintf("%s\t%s\n", host, name+"."+ns)
			k8sHosts += fmt.Sprintf("%s\t%s\n", host, name+"."+ns+"."+"svc.cluster.local")
		}
	}

	for name, host := range nsHosts["default"] {
		k8sHosts += fmt.Sprintf("%s\t%s\n", host, name)
	}

	k8sHosts = k8sHosts + utils.HOST_END
	return k8sHosts, nil
}

func GetServices() (map[string][]corev1.Service, error) {
	return _client.GetNsServices()
}

func GetNamespaces() ([]string, error) {
	var namespaceList corev1.NamespaceList
	output, err := _client.remoteExec("k3s kubectl get ns -o json")
	if err != nil {
		return []string{}, err
	}

	// 解析 JSON 数据
	if err := json.Unmarshal(output, &namespaceList); err != nil {
		return []string{}, fmt.Errorf("解析命名空间列表失败: %v", err)
	}
	namespaces := make([]string, len(namespaceList.Items))
	for i, ns := range namespaceList.Items {
		namespaces[i] = ns.Name
	}
	return namespaces, nil
}
