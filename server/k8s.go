package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var config *rest.Config

func init_k8s() error {
	// 设置 kubeconfig 文件路径
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	// 检查 kubeconfig 文件是否存在
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return err
	} else {
		// kubeconfig 文件存在，使用它创建 config
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		return err
	}
}

func getK8sHosts() (string, error) {
	header := "## Platform START\n"
	end := "## Platform END\n"
	k8sHosts := header
	// 创建 clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return k8sHosts, err
	}
	// 获取所有命名空间
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return k8sHosts, err
	}

	// 遍历所有命名空间，获取服务
	for _, ns := range namespaces.Items {
		services, err := clientset.CoreV1().Services(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing services in namespace %s: %v", ns.Name, err)
			continue
		}
		hlSvc := make(map[string]corev1.Service)
		items := make(map[string]corev1.Service)

		for _, svc := range services.Items {
			if svc.Spec.ClusterIP == "None" {
				hlSvc[svc.Name] = svc
			} else {
				items[svc.Name] = svc
			}
		}

		for _, svc := range items {
			if s, ok := hlSvc[svc.Name+"-hl"]; ok {
				k8sHosts += fmt.Sprintf("%s\t%s\n", svc.Spec.ClusterIP, s.Name+"."+ns.Name+"."+"svc.cluster.local")
			}
			k8sHosts += fmt.Sprintf("%s\t%s\n", svc.Spec.ClusterIP, svc.Name+"."+ns.Name+"."+"svc.cluster.local")
		}
	}
	k8sHosts += end
	return k8sHosts, nil
}

func getK8sNet() (string, error) {
	return "10.96.0.0/12", nil
}
