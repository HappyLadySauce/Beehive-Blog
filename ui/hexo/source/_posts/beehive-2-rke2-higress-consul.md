---
title: RKE2 higress Consul 多集群适配
slug: rke2-higress-consul
date: 2026-04-10T15:33:09.250559+08:00
updated: 2026-04-10T15:33:09.251219+08:00
categories:
  - 默认分类
tags:
  - kubernetes
  - 网络系统管理
beehive_id: 2
---

# RKE2 多集群服务治理实践：Consul + Higress

本文档是一份详细的实践指南，旨在引导您在多个 RKE2 Kubernetes 集群之上，构建一个全面、健壮的微服务治理平台。该方案整合了多种业界领先的开源技术：

- **RKE2**: 作为安全合规的 Kubernetes 发行版，为我们的部署提供坚实的基础。
- **Submariner**: 解决跨集群网络的核心挑战，实现集群间的 Pod-to-Pod 直接通信。
- **Consul**: 提供强大的服务网格能力，实现跨集群的服务发现、健康检查和流量管理。
- **Higress**: 作为高性能的云原生网关，统一管理所有入站流量，并与 Consul 服务网格无缝集成。

我们将从底层网络打通开始，逐步完成 Consul 和 Higress 的部署与配置，并详细记录和解决在 RKE2 环境下遇到的典型问题，例如存储类缺失和 Ingress 控制器冲突。最终，您将获得一个能够实现跨集群服务安全通信、灵活路由和统一流量入口的生产级解决方案。

## 环境规划

本次实践环境包含两个 RKE2 集群（主备）和一个独立的镜像仓库。

```txt
# 主集群 (Master)
10.10.10.250 M-RKE2-Balance.huinong.internal
10.10.10.11 M-RKE2-Master01.huinong.internal
10.10.10.12 M-RKE2-Master02.huinong.internal
10.10.10.13 M-RKE2-Master03.huinong.internal
10.10.10.14 M-RKE2-Node01.huinong.internal
10.10.10.15 M-RKE2-Node02.huinong.internal

# 备集群 (Branch)
10.10.20.250 B-RKE2-Balance.huinong.internal
10.10.20.11 B-RKE2-Master01.huinong.internal
10.10.20.12 B-RKE2-Master02.huinong.internal
10.10.20.13 B-RKE2-Master03.huinong.internal
10.10.20.14 B-RKE2-Node01.huinong.internal

# 镜像仓库
10.10.10.254 registry.huinong.internal
```

**集群网络规划**：

为避免网络冲突，每个集群都规划了独立的 CIDR 地址段。

```txt
# 总部集群网络配置
cluster-cidr: "10.44.0.0/16"
service-cidr: "10.45.0.0/16"

# 分部集群网络配置
cluster-cidr: "10.46.0.0/16"
service-cidr: "10.47.0.0/16"
```

---

## 使用 Submariner 实现跨集群网络

### **为什么选择 Submariner？**

- **直接 Pod-to-Pod 通信**：通过 VPN 隧道或 UDP 封装（默认使用 Libreswan），无需依赖 NodePort/LoadBalancer。
- **自动同步 Service**：支持跨集群的 `ServiceImport`，可与服务网格无缝集成。
- **兼容性**: Submariner 作为纯粹的网络层解决方案，不干扰上层应用的部署和管理。

#### 在 Karmada 控制面或独立集群中部署 Broker

```bash
subctl deploy-broker --kubeconfig .kube/config --repository registry.huinong.internal/quay.io
```

#### 在每个 RKE2 集群中加入 Submariner

```bash
# Branch 分部
subctl join --kubeconfig .kube/branch.config broker-info.subm --clusterid branch --repository registry.huinong.internal/quay.io --globalnet-cidr 244.10.20.0/24 --nattport 4500 --force-udp-encaps --natt=true --cable-driver libreswan --air-gapped

# Master 总部
subctl join --kubeconfig .kube/master.config broker-info.subm --clusterid master --repository registry.huinong.internal/quay.io --globalnet-cidr 244.10.10.0/24 --nattport 4500 --force-udp-encaps --natt=true --cable-driver libreswan --air-gapped
```

在 Master 和 branch 中执行 `subctl show all` 验证查看是否执行成功。

---

## Consul

Consul是由HashiCorp公司开发的一款开源工具，主要用于服务发现、配置管理和分布式系统监控。其主要功能包括：

- **服务发现**：Consul提供服务注册和健康检查机制，使得微服务架构中的各个服务实例能够自动注册到Consul中，并通过DNS或HTTP API实现服务间的互相发现。
- **配置共享与管理**：Consul可以作为配置中心存储和分发配置信息给各个应用节点，支持KV存储，允许动态更新配置并在集群中快速传播。
- **健康检查**：通过健康检查机制，Consul能够持续监控服务的健康状态，确保只有健康的实例才能被调用。
- **多数据中心支持**：Consul使用基于RAFT协议的强一致性保证，可以实现跨多个数据中心的服务发现和配置同步。

### 安装 Consul

因为项目是多集群环境，我们采用主从模式，在 master 集群和 branch 集群都安装。

```shell
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo
sudo yum -y install consul
```

创建 consul 名称空间。

```shell
kubectl create ns consul
```

创建生成 Gossip 加密密钥并将其保存为 Kubernetes 密钥。

```shell
kubectl create secret generic consul-gossip-encryption-key --from-literal=key=$(consul keygen) -n consul
```

更新 helm 仓库以支持 `hashicorp` 安装。

```shell
helm repo add hashicorp https://helm.releases.hashicorp.com
```

从 helm 仓库中导出，修改镜像文件地址。

```shell
helm pull hashicorp/consul --destination .
```

---

### 在RKE2中因存储类缺失导致安装失败的问题与解决

#### 1. 问题背景

按照 Consul 官方文档在 Kubernetes 多集群部署时，遇到 `consul-server-0` pod 处于 `Pending` 状态。通过 `kubectl describe pod` 查看发现，原因是 PVC (PersistentVolumeClaim) 无法绑定，进一步排查发现是 RKE2 集群环境中没有默认的 StorageClass (存储类)。

#### 2. 解决方案

核心思路是为集群手动创建一个 `StorageClass`，并为 Consul Server 手动创建对应的 `PersistentVolume` (PV)，然后通过 Helm values 文件指定使用我们创建的存储类。

##### 步骤 1: 创建存储资源文件

首先，我们需要创建一个 YAML 文件来定义 `StorageClass` 和 `PersistentVolume`。

```bash
cat > consul-storage.yaml << 'EOF'
# 创建一个名为 consul-hostpath 的存储类，并设为默认
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: consul-hostpath
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Delete

---
# 为 consul-server-0 创建一个PV
apiVersion: v1
kind: PersistentVolume
metadata:
  name: consul-server-0-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: consul-hostpath
  local:
    # 指定数据在节点上的存储路径
    path: /opt/consul-data/server-0
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
            # **注意**: 这里需要绑定到具体的节点名称
            - m-rke2-node01.huinong.internal
EOF
```

##### 步骤 2: 在节点上创建数据目录

由于我们使用的是 `hostpath` 类型的存储，需要在上一步 `consul-storage.yaml` 文件中指定的节点上（`m-rke2-node01.huinong.internal`）手动创建数据目录并授权。

```bash
# SSH登录到指定节点
ssh m-rke2-node01.huinong.internal

# 创建目录并赋予权限
mkdir -p /opt/consul-data/server-0
chmod 777 /opt/consul-data/server-0

# 退出
exit
```

##### 步骤 3: 应用存储配置

在 master 节点上应用刚才创建的 `consul-storage.yaml` 文件。

```bash
kubectl apply -f consul-storage.yaml -n consul
```

##### 步骤 4: 清理并重新部署 Consul

如果之前有失败的部署，需要先清理干净。

```bash
# 卸载 helm release
helm uninstall consul-master-cluster -n consul

# 删除 consul 命名空间下的所有资源
kubectl delete all --all -n consul
kubectl delete pvc --all -n consul
kubectl delete secret --all -n consul
```

然后，按照您之前的流程重新部署，**关键在于修改 `cluster1-values.yaml` 文件，增加 `storageClass` 配置**。

```bash
# 1. 重新创建命名空间
kubectl create ns consul

# 2. 重新创建Gossip加密密钥
kubectl create secret generic consul-gossip-encryption-key --from-literal=key=$(consul keygen) -n consul

# 3. 准备更新后的 values 文件
cat > cluster1-values.yaml << 'EOF'
global:
  datacenter: dc1
  tls:
    enabled: true
    enableAutoEncrypt: true
  acls:
    manageSystemACLs: true
  gossipEncryption:
    secretName: consul-gossip-encryption-key
    secretKey: key
server:
  # 关键配置: 指定使用我们创建的存储类
  storageClass: consul-hostpath
  exposeService:
    enabled: true
    type: NodePort
    nodePort:
      http: 30010
      https: 30011
      serf: 30012
      rpc: 30013
      grpc: 30014
ui:
  service:
    type: NodePort
EOF

# 4. 使用本地 chart 进行部署
helm install consul-master-cluster --values values.yaml . --namespace consul
```

##### 步骤 5: 验证部署状态

```bash
# 检查 Pods 是否都 Running
kubectl get pods -n consul

# 检查 PVC 是否成功 Bound
kubectl get pvc -n consul
```

#### 3. 关键要点总结

1. **storageClass**: 核心是在 `values.yaml` 的 `server` 配置下增加 `storageClass`，指定我们手动创建的存储类。
2. **手动创建 PV**: 由于 `consul-hostpath` 存储类的 `provisioner` 是 `kubernetes.io/no-provisioner`，Kubernetes 不会自动创建 PV，需要我们手动创建。
3. **节点亲和性 (nodeAffinity)**: 手动创建 PV 时，必须通过 `nodeAffinity` 将其绑定到数据目录所在的具体节点，否则 Pod 可能被调度到没有数据目录的节点而失败。
4. **目录权限**: 节点上的数据目录 (`/opt/consul-data/server-0`) 必须有正确的权限，`chmod 777` 是一个简单有效的办法，确保 Pod 内的进程有权限读写。

#### 4. 扩展到多 Server 节点 (高可用)

如果需要部署多个 Consul Server 实例（例如3个）以实现高可用，您需要：

1. 为每个 Server 实例创建对应的 PV。
2. 在 `values.yaml` 中设置 `server.replicas=3`。 例如，为 `server-1` 添加 PV：

```yaml
---
# 为 consul-server-1 创建一个PV
apiVersion: v1
kind: PersistentVolume
metadata:
  name: consul-server-1-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: consul-hostpath
  local:
    path: /opt/consul-data/server-1
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                # 绑定到另一个节点
                - m-rke2-node02.huinong.internal
```

同时，不要忘记在 `m-rke2-node02.huinong.internal` 节点上创建 `/opt/consul-data/server-1` 目录。

---

使用 helm 进行本地安装。

```shell
helm install consul-master-cluster --values cluster1-values.yaml . --namespace consul
```

安装完成并且所有组件都运行并准备就绪后，需要提取以下信息（使用以下命令）并将其应用于第二个 Kubernetes 集群。

```
kubectl get secret consul-master-cluster-consul-ca-cert consul-master-cluster-consul-bootstrap-acl-token --output yaml > cluster1-credentials.yaml -n consul
```

---

### Branch 集群配置与验证

在主集群（Master）的 Consul Server 部署成功后，我们需要在第二个集群（Branch）上部署 Consul Client，并将其连接到主集群，最终形成一个统一的服务网格。

#### 步骤 1: 提取主集群凭据

根据 [Consul多集群部署文档](https://developer.hashicorp.com/consul/docs/deploy/server/k8s/multi-cluster)，我们需要从主集群导出 CA 证书和 ACL Bootstrap Token，以便让第二个集群信任并加入。

在**主集群**上执行：

```bash
# 注意：这里的 helm release 名称是 consul-master-cluster
kubectl get secret consul-master-cluster-consul-ca-cert consul-master-cluster-consul-bootstrap-acl-token --output yaml > cluster1-credentials.yaml -n consul
```

然后将生成的 `cluster1-credentials.yaml` 文件传输到**第二个集群**的 `/opt/consul/` 目录下。

#### 步骤 2: 准备第二个集群的 Helm Values

在**第二个集群**上，我们需要创建一个新的 values 文件 (`cluster2-values.yaml`) 来配置 Consul Client。

关键配置项说明：

- `externalServers`: 这是核心配置，用于告诉第二个集群的 Consul Client 如何找到第一个集群的 Consul Server。
- `hosts`: 第一个集群中任意一个节点的 IP 地址。这里使用 `10.10.10.11`。
- `httpsPort`: 第一个集群暴露的 Consul UI 服务的 `NodePort` 端口。通过 `kubectl get svc -n consul` 查到是 `32130`。
- `grpcPort`: 第一个集群在 values 文件中定义的 gRPC `NodePort` 端口，即 `30014`。
- `k8sAuthMethodHost`: 第二个集群可从外部访问的 Kubernetes API Server 地址。这里是 `https://10.10.20.11:6443`。

在**第二个集群**的 `/opt/consul/` 目录下创建 `cluster2-values.yaml`:

```yaml
cat > cluster2-values.yaml << 'EOF'
global:
  enabled: false
  datacenter: dc1
  acls:
    manageSystemACLs: true
    bootstrapToken:
      secretName: consul-master-cluster-consul-bootstrap-acl-token
      secretKey: token
  tls:
    enabled: true
    caCert:
      secretName: consul-master-cluster-consul-ca-cert
      secretKey: tls.crt
externalServers:
  enabled: true
  hosts: ["10.10.10.11"]
  httpsPort: 32130
  grpcPort: 30014
  tlsServerName: server.dc1.consul
  k8sAuthMethodHost: https://10.10.20.11:6443
connectInject:
  enabled: true
EOF
```

#### 步骤 3: 在第二个集群部署 Consul Client

在**第二个集群**上执行以下命令：

```bash
# 1. 创建 consul 命名空间 (如果不存在)
kubectl create ns consul

# 2. 应用从主集群拷贝过来的凭据
kubectl apply -f cluster1-credentials.yaml -n consul

# 3. 使用新的 values 文件部署 Consul Client
# 注意 helm release 名称 consul-branch-cluster 与主集群不同
helm install consul-branch-cluster --values cluster2-values.yaml . --namespace consul
```

#### 步骤 4: 跨集群服务网格验证

为了验证两个集群是否真的通过服务网格联通了，我们进行一个经典的跨集群服务调用测试。

1. **在主集群部署`static-server`**:

   - 创建一个 `static-server.yaml` 文件。
   - 注意 `image` 使用了本地镜像仓库 `registry.huinong.internal/hashicorp/http-echo:latest`。

   ```yaml
   # static-server.yaml
   ---
   apiVersion: consul.hashicorp.com/v1alpha1
   kind: ServiceIntentions
   metadata:
     name: static-server
   spec:
     destination:
       name: static-server
     sources:
       - name: static-client
         action: allow
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: static-server
   spec:
     type: ClusterIP
     selector:
       app: static-server
     ports:
       - protocol: TCP
         port: 80
         targetPort: 8080
   ---
   apiVersion: v1
   kind: ServiceAccount
   metadata:
     name: static-server
   ---
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: static-server
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: static-server
     template:
       metadata:
         name: static-server
         labels:
           app: static-server
         annotations:
           "consul.hashicorp.com/connect-inject": "true"
       spec:
         containers:
           - name: static-server
             image: registry.huinong.internal/hashicorp/http-echo:latest
             args:
               - -text="hello world from cluster1"
               - -listen=:8080
             ports:
               - containerPort: 8080
                 name: http
         serviceAccountName: static-server
   ```

   - 在**主集群**应用: `kubectl apply -f static-server.yaml -n consul`

2. **在第二个集群部署 `static-client`**:

   - 创建一个 `static-client.yaml` 文件。
   - **注意**：`image` 同样需要使用本地可访问的镜像。由于 `curlimages/curl` 无法拉取，我们换成了本地已有的 `registry.huinong.internal/hashicorp/consul:1.21.1`，该镜像内含 `curl` 工具。
   - `consul.hashicorp.com/connect-service-upstreams`: 这是实现跨集群调用的关键注解，它告诉 Consul Connect sidecar 将到 `localhost:1234` 的流量代理到 `static-server` 服务。

   ```yaml
   # static-client.yaml
   apiVersion: v1
   kind: Service
   metadata:
     name: static-client
   spec:
     selector:
       app: static-client
     ports:
       - port: 80
   ---
   apiVersion: v1
   kind: ServiceAccount
   metadata:
     name: static-client
   ---
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: static-client
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: static-client
     template:
       metadata:
         name: static-client
         labels:
           app: static-client
         annotations:
           "consul.hashicorp.com/connect-inject": "true"
           "consul.hashicorp.com/connect-service-upstreams": "static-server:1234"
       spec:
         containers:
           - name: static-client
             image: registry.huinong.internal/hashicorp/consul:1.21.1
             command: ["/bin/sh", "-c", "--"]
             args: ["while true; do sleep 30; done;"]
         serviceAccountName: static-client
   ```

   - 在**第二个集群**应用: `kubectl apply -f static-client.yaml -n consul`

3. **执行最终测试**:

   - 在**第二个集群**上，进入 `static-client` 的 Pod，执行 `curl` 命令。

```bash
# 在第二个集群的 master 节点执行
kubectl exec -n consul deploy/static-client -c static-client -- curl --silent localhost:1234
```

#### 🎉 成功验证

执行上述命令后，终端成功返回：

```
"hello world from cluster1"
```

这标志着 Consul 多集群服务网格已成功建立并正常工作。第二个集群的服务可以通过服务网格安全、透明地调用第一个集群的服务。

### 访问 Consul UI

Consul 已经通过 NodePort 暴露了 UI 服务，可以直接访问，无需修改原生配置。

#### 直接访问方式

1. **通过 NodePort 直接访问**：

```bash
# 通过 HTTPS 访问（推荐）
https://10.10.10.11:32130/ui/

# 或使用任意集群节点IP
https://<任意节点IP>:32130/ui/
```

2. **查看服务状态**：

```bash
kubectl get svc -n consul consul-master-cluster-consul-ui
```

3. **配置本地域名解析**（可选）：

   在 `/etc/hosts` 文件中添加：

```text
10.10.10.11 consul.huinong.internal
```

然后通过域名访问：

```bash
https://consul.huinong.internal:32130/ui/
```

#### 服务信息

- **服务名称**: `consul-master-cluster-consul-ui`
- **服务类型**: NodePort
- **端口映射**: `443:32130/TCP`
- **协议**: HTTPS（使用自签名证书）
- **访问路径**: `/ui/`

#### 重要说明

- Consul UI 使用 HTTPS 协议和自签名证书，浏览器可能会显示安全警告，点击"继续访问"即可

- 如需登录，请使用 bootstrap ACL token，可通过以下命令获取：

```bash
kubectl get secret consul-master-cluster-consul-bootstrap-acl-token -n consul -o jsonpath='{.data.token}' | base64 -d && echo
```

输出结果

```shell
331c00f9-bd87-2383-4394-548a0e66dea9
```

- 原生 Consul 配置保持不变，所有功能正常工作

### Consul 部署成果

- ✅ 两个集群的Consul部署成功
- ✅ 跨集群服务发现和通信验证通过
- ✅ 安全配置(TLS/ACL/Gossip加密)正常
- ✅ 服务网格功能完整工作
- ✅ Consul UI 通过 NodePort 可正常访问

## Higress Ingress on RKE2: 配置与冲突解决

本节详细记录了在 RKE2 Kubernetes 集群中，因默认的 Nginx Ingress Controller 与新部署的 Higress Ingress Controller 产生冲突，导致流量转发异常问题的整个排查与解决过程。

### 1. 问题背景与目标

- **环境**: RKE2 Kubernetes 集群，已默认安装并启用了 `rke2-ingress-nginx-controller`。
- **新组件**: 部署了 Higress 作为新的云原生网关。
- **目标**: 为 `consul.huinong.internal` 和 `higress.huinong.internal` 两个服务配置 Ingress 规则，并指定由 Higress Controller 进行处理，最终实现通过域名对服务的访问。

### 2. 初始症状

在为两个服务创建了基于 Higress Class 的 Ingress 资源后，出现了以下问题：

1. **Ingress 地址为空**: 通过 `kubectl get ingress -A` 查看时，新创建的 Ingress 资源 `ADDRESS` 字段为空。
2. **流量被 Nginx 拦截**: 使用 `curl` 访问目标域名时，收到的响应是 RKE2 默认 Nginx Ingress Controller 的 404 页面，流量并未被 Higress Gateway 处理。

这表明，尽管 Ingress 规则中指定了 `ingressClassName: higress`，但流量在到达 Higress 之前，就被 RKE2 的 Nginx Ingress 捕获并处理了。

### 3. 排查与解决步骤

#### 步骤 3.1: 定位 HostPort 端口冲突

我们首先怀疑存在端口冲突。通过检查 RKE2 Nginx Pod 的配置，定位到了问题的根源。

- **排查命令**:

```bash
kubectl describe pod rke2-ingress-nginx-controller-xxxxx -n kube-system
```

- **关键发现**: 在 Pod 的描述信息中，可以看到如下配置：

```
Host Ports:  80/TCP, 443/TCP
```

- **结论**: RKE2 的 Nginx Controller 通过 `hostPort` 的方式，直接绑定并占用了宿主节点的 80 和 443 端口。这导致了所有外部流量一进入节点就被 Nginx 进程截获，无法到达 Higress Gateway 的 Service (无论是 `NodePort` 还是 `LoadBalancer`)。

#### 步骤 3.2: 禁用 RKE2 Nginx Ingress

为了解决 `hostPort` 冲突，我们决定彻底禁用 RKE2 自带的 Nginx Ingress。

- **解决方案**: 删除 RKE2 Nginx Controller 的 `DaemonSet` 资源。这会停止所有相关的 Pod，从而释放对宿主机 80 和 443 端口的占用。

- **执行命令**:

```bash
kubectl delete daemonset rke2-ingress-nginx-controller -n kube-system
```

#### 步骤 3.3: 解决 HTTPS 证书兼容性问题

禁用 Nginx 后，HTTP 流量（80端口）可以正常被 Higress Gateway 接收。但尝试配置和访问 HTTPS 服务时，遇到了新的问题。

- **问题现象**: HTTPS 访问失败。

- 排查方式查看 Higress Gateway Pod 的日志。

```bash
kubectl logs -n higress-system $(kubectl get pods -n higress-system | grep higress-gateway | head -1 | awk '{print $1}')
```

- 关键发现日志中出现明确错误：

```shell
Failed to load certificate chain from <inline>, only P-256 ECDSA certificates are supported
```

- **结论**: 我们提供的 `*.huinong.internal` 通配符证书是 RSA 加密类型，而当前版本的 Higress/Envoy 要求使用 P-256 ECDSA 类型的证书。证书不兼容导致 TLS 握手失败。

#### 步骤 3.4: 切换为纯 HTTP 访问 (临时方案)

由于证书不兼容，我们决定暂时放弃 HTTPS，将服务全部切换为通过纯 HTTP 访问以验证路由的连通性。

- **解决方案**: 修改 `consul-ingress.yaml` 和 `higress-ingress.yaml` 文件，移除其中的 `tls` 配置块和 `higress.io/ssl-redirect` 注解。

- **修改示例 (`consul-ingress.yaml`)**:

```yaml
  apiVersion: networking.k8s.io/v1
  kind: Ingress
  metadata:
    name: consul-ui-ingress
    namespace: consul
    annotations:
      ingressclass.kubernetes.io/is-default-class: "true"
  -   higress.io/ssl-redirect: "true"
      higress.io/backend-protocol: "HTTPS"
      higress.io/proxy-ssl-verify: "off"
  spec:
    ingressClassName: higress
  - tls:
  - - hosts:
  -   - consul.huinong.internal
  -   secretName: huinong-internal-tls
    rules:
    - host: consul.huinong.internal
      # ...
```

#### 步骤 3.5: 解决 Nginx Admission Webhook 冲突

在应用纯 HTTP 的 Ingress 配置时，操作失败，遇到了 Webhook 错误。

- **问题现象**: `kubectl apply` 报错，信息指向 `validate.nginx.ingress.kubernetes.io`。

- **原因分析**: 即使 Nginx Controller 的 Pod 已经被删除，但其 Admission Controller（准入控制器）的配置仍然在集群中生效。这个 Webhook 会拦截并校验所有 Ingress 资源的创建和修改请求，由于其依赖的服务已不存在，导致请求失败。

- **解决方案**: 删除 Nginx 的 `ValidatingWebhookConfiguration`。

- **执行命令**:

```bash
# 1. 查找相关的 Webhook 配置
kubectl get validatingwebhookconfigurations | grep nginx

# 2. 删除该配置
kubectl delete validatingwebhookconfigurations rke2-ingress-nginx-admission
```

#### 步骤 3.6: 实施 HTTPS 最终解决方案

在确认了路由和基本访问正常后，我们回到证书问题上，并实施最终的HTTPS解决方案。

- **证书类型确认**: 经过确认，用户可以生成 **EC 256** 类型的证书。这完全符合 Higress/Envoy 要求的 P-256 ECDSA 规范。

- **解决方案**:

  1. **生成 EC 256 证书**: 生成一张 `*.huinong.internal` 的通配符证书，确保其加密算法为 `EC 256`

  2. 创建新的 TLS Secret，将新生成的 EC 256 证书和私钥创建为 Kubernetes Secret。建议使用新的名称以区分之前的 RSA 证书。

```bash
# 在 consul 和 higress-system 两个命名空间下都创建
kubectl create secret tls huinong-internal-tls-ec256 --cert=fullchain.pem --key=privkey.pem -n consul
 kubectl create secret tls huinong-internal-tls-ec256 --cert=fullchain.pem --key=privkey.pem -n higress-system
```

3. 更新 Ingress 资源修改

```
consul-ingress.yaml
higress-ingress.yaml
```

4. 重新启用 TLS，并指向新创建的

```
huinong-internal-tls-ec256-ec256
```

```yaml
# ...
spec:
	ingressClassName: higress
	tls:
	- hosts:
	- consul.huinong.internal # 或者 higress.huinong.internal
	secretName: huinong-internal-tls-ec256 # 指向新的EC256证书
	rules:
	# ...
```

5. 应用配置

```bash
kubectl apply -f consul-ingress.yaml
kubectl apply -f higress-ingress.yaml
```

### 4. 最终配置与验证

在清除了所有障碍并使用了兼容的证书后，我们成功地应用了 HTTPS Ingress 配置。

- **状态验证**: `kubectl get ingress -A` 显示 Ingress 状态正常，`ADDRESS` 字段也已正确填充，`PORTS` 包含 80, 443。

- **访问验证**: 使用 `curl` 命令进行最终测试，结果符合预期。

  ```bash
  # 测试 HTTP (应自动重定向到 HTTPS)
  $ curl -v http://consul.huinong.internal
  < HTTP/1.1 308 Permanent Redirect
  < location: https://consul.huinong.internal/
  < server: istio-envoy

  # 测试 HTTPS
  $ curl -k https://consul.huinong.internal/ui/
  # ... 成功返回 Consul UI 的 HTML 页面内容

  $ curl -k https://higress.huinong.internal/
  # ... 成功返回 Higress Console 的 HTML 页面内容
  ```

````

### 5. 总结

本次冲突的解决过程核心在于处理了 RKE2 环境下默认组件与新增组件之间的资源竞争和配置残留问题。关键的经验总结如下：

1. **HostPort 是首要排查点**: 在多 Ingress Controller 场景下，需要特别注意是否有 Controller 使用 `hostPort` 方式抢占了标准的 HTTP/HTTPS 端口。
2. **关注组件兼容性**: 使用 Higress 等基于 Envoy 的网关时，需注意其对 TLS 证书类型的特定要求（如 P-256 ECDSA）。
3. **彻底清理禁用组件**: 禁用一个 Kubernetes 组件时，不仅要删除其工作负载（如 Pod/DaemonSet），还需清理其关联的配置资源，特别是 `Service` 和 `ValidatingWebhookConfiguration` 等，以避免对集群其他操作产生干扰。

设置默认使用 higress 作为 ingress 网关。

```shell
kubectl annotate ingressclass higress ingressclass.kubernetes.io/is-default-class=true
````

## 总结与展望

通过本文档的详细步骤，我们成功地在一个基于 RKE2 的多集群环境中，搭建了一套完整且强大的微服务治理体系。我们通过 **Submariner** 打通了集群间的网络壁垒，利用 **Consul** 构建了统一的服务网格，实现了跨集群的服务发现与安全通信，并最终通过 **Higress** 作为统一流量入口，解决了与 RKE2 默认组件的冲突，提供了灵活高效的 Ingress 管理能力。

这套架构的优势在于：

- **标准化与解耦**: 各组件各司其职，从底层网络、服务治理到流量入口都采用了业界主流的云原生方案。
- **高可用与可扩展**: 多集群部署本身提供了高可用性基础，Consul 和 Higress 的设计也支持水平扩展。
- **安全可靠**: RKE2 的安全特性结合 Consul 的 mTLS 加密，为服务间通信提供了坚实的安全保障。

未来，您可以在此基础上进一步探索更多高级功能，例如：

- **引入联邦控制平面**: 探索使用 Karmada 或 KubeFed 等工具实现更高级的多集群应用分发和调度策略。
- **Consul 的高级流量策略**: 如 A/B 测试、金丝雀发布、流量分割等。
- **Higress 的高级插件**: 利用 Wasm 插件机制，实现自定义认证、监控、安全等逻辑。

希望本实践指南能为您在构建多集群微服务平台时提供有价值的参考。
