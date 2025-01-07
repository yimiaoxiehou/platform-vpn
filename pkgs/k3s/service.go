package k3s

import "k8s.io/apimachinery/pkg/util/intstr"

type ServiceList struct {
	ApiVersion string    `json:"apiVersion"`
	Items      []Service `json:"items"`
	Kind       string    `json:"kind"`
}

type Service struct {
	ApiVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
	Status     Status   `json:"status"`
}

type Metadata struct {
	Annotations       map[string]string `json:"annotations"`
	CreationTimestamp string            `json:"creationTimestamp"`
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	ResourceVersion   string            `json:"resourceVersion"`
	Uid               string            `json:"uid"`
}

type Spec struct {
	ClusterIP             string            `json:"clusterIP"`
	ClusterIPs            []string          `json:"clusterIPs"`
	InternalTrafficPolicy string            `json:"internalTrafficPolicy"`
	IpFamilies            []string          `json:"ipFamilies"`
	IpFamilyPolicy        string            `json:"ipFamilyPolicy"`
	Ports                 []Port            `json:"ports"`
	Selector              map[string]string `json:"selector"`
	SessionAffinity       string            `json:"sessionAffinity"`
	Type                  string            `json:"type"`
}

type Port struct {
	Name       string             `json:"name"`
	Port       int                `json:"port"`
	Protocol   string             `json:"protocol"`
	TargetPort intstr.IntOrString `json:"targetPort"`
}

type Status struct {
	LoadBalancer map[string]interface{} `json:"loadBalancer"`
}
