

## platform 开发平台 vpn 工具 服务端

### 使用方式

```
# ps: /root/.kube 为 k8s 配置文件目录，其中需要有 k8s 认证文件 config
docker run -d --network host --name vpn-server -v /root/.kube:/root/.kube docker.utpf.cn/docker.io/yimiaoxiehou/platform-vpn-server

```