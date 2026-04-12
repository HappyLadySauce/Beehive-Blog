---
title: 技能大赛国赛 Ubuntu + RKE2 + Karmada + Karmada-Dashboard(二开) + Hingress 方案
slug: ubuntu-rke2-karmada-karmada-dashboard-hingress
date: 2026-04-12T12:11:32+08:00
updated: 2026-04-12T20:12:22.506182+08:00
categories:
  - 默认分类
tags:
  - docker
  - kubernetes
  - ubuntu
  - 网络系统管理
  - 计算机网络
beehive_id: 6
---

通过我的不断摸索，终于确定了 `Kubernetes` 多集群管理搭建的架构模型。

## 前期环境准备

### VMware 虚拟网络

首先配置 VMware 虚拟网卡，需要配置为项目需求的配置，服务器集群使用两个网段一个是总部城市银行网络（10.10.10.0/24）另一个是分部乡镇银行（10.10.20.0/24）。

### 模板机

#### 安装过程中

然后得准备一台模板虚拟机用于克隆初始化环境的系统。使用 `ubunut 22.04` 创建一台虚拟机 `Ubuntu-Template` 。模板机配置说明：CPU：4核，内存：6G，硬盘：60G。

初始化网络，配置为：

- 网段：10.10.10.0/24
- IP：10.10.10.254
- 网关：10.10.10.1
- DNS Server：223.5.5.5

初始化磁盘，配置为：

- `/` 目录 50G
- `/home` 5G
- 其他内存保留备用

#### 安装完成后

安装完成后，首先创建 root 用户

```shell
sudo passwd root
```

为了防止 `Ubuntu` 网卡开机还原，我们还需要修改网卡参数文件。

```shell
touch /etc/cloud/cloud.cfg.d/99-disable-network-config.cfg
echo "network: {config: disabled}" > /etc/cloud/cloud.cfg.d/99-disable-network-config.cfg
```

最后关机，打快照，名称为 `初始化`。

---

## RKE2 集群部署

### 前期准备

#### 修改主机名

使用模板机快照**完整复制** 15 台虚拟机，其中 6 台作为 `K8S` 控制节点，4 台作为 `K8S` 工作节点，2台作为集群 `nginx` 七层网关，1 台作为 `Kuboard` 集群管理节点并且承担 `Harbor` 镜像仓库的职能，最后 2 台作为 `Karmada` 宿主集群的承载机，一个为控制节点，一个为工作节点。名称命名分别为：

- K8S-M-Master01
- K8S-M-Master02
- K8S-M-Master03
- K8S-M-Node01
- K8S-M-Node02
- K8S-M-Nginx
- K8S-B-Master01
- K8S-B-Master02
- K8S-B-Master03
- K8S-B-Node01
- K8S-B-Node02
- K8S-B-Nginx
- Karmada-Master
- Karmada-Node
- Harbor

创建好虚拟机后，首先修改每一台机器的网络IP地址：
- K8S-M-Master01（10.10.10.11/24）
- K8S-M-Master02（10.10.10.12/24）
- K8S-M-Master03（10.10.10.13/24）
- K8S-M-Node01（10.10.10.14/24）
- K8S-M-Node02（10.10.10.15/24）
- K8S-M-Nginx（10.10.10.250/24）
- K8S-B-Master01（10.10.20.11/24）
- K8S-B-Master02（10.10.20.12/24）
- K8S-B-Master03（10.10.20.13/24）
- K8S-B-Node01（10.10.20.14/24）
- K8S-B-Node02（10.10.20.15/24）
- K8S-M-Nginx（10.10.20.250/24）
- Karmada-Master（10.10.10.2/24）
- Karmada-Node（10.10.10.3/24）
- Harbor（10.10.10.254/24）

网络IP地址修改完成后，我们就可以通过 `MobaX` 超级终端软件连接到每一个机器节点，方便我们操作。连接上去之后第一件事是修改每一台主机的主机名称，然后再保证每台机器的时间同步，这是 `K8S` 强制要求的。

修改主机名例子：

```shell
hostnamectl set-hostname K8S-M-Master01.huinong.internal
```

所有的机器主机名修改完成之后，我们需要在所有机器上添加主机名 IP 映射关系（本地 DNS 解析）。

```shell
vim /etc/hosts
```

在文件内容底部添加内容：

```shell
10.10.10.254 Harbor
# 总部集群
10.10.10.11 K8S-M-Master01.huinong.internal
10.10.10.12 K8S-M-Master02.huinong.internal
10.10.10.13 K8S-M-Master03.huinong.internal
10.10.10.14 K8S-M-Node01.huinong.internal
10.10.10.15 K8S-M-Node02.huinong.internal
10.10.10.250 K8S-M-Nginx.huinong.internal
# 分部集群
10.10.20.11 K8S-B-Master01.huinong.internal
10.10.20.12 K8S-B-Master02.huinong.internal
10.10.20.13 K8S-B-Master03.huinong.internal
10.10.20.14 K8S-B-Node01.huinong.internal
10.10.20.15 K8S-B-Node02.huinong.internal
10.10.20.250 K8S-B-Nginx.huinong.internal
# Karmada
10.10.10.2 Karmada-Master.huinong.internal
10.10.10.3 Karmada-Node.huinong.internal
```

### 系统前置环境配置

```shell
##关闭防火墙
systemctl stop firewalld
##禁止防火墙开机启动
systemctl disable firewalld
##永久关闭selinux 注：重启机器后，selinux配置才能永久生效
sed -i 's/enforcing/disabled/' /etc/selinux/config
##临时关闭selinux   执行getenforce   显示Disabled说明selinux已经关闭
setenforce 0
sudo sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config

##临时关闭交换分区swap
swapoff -a
##永久关闭交换分区swap 注：重启机器后，才能永久生效
sed -ri 's/.*swap.*/#&/' /etc/**fstab**
```

```shell
echo "* hard nofile 65536" >> /etc/security/limits.conf
echo "* soft nofile 65536" >> /etc/security/limits.conf
```

### 时间同步服务

##### Karmada 时间服务器

然后我们使用 `Kuboard` 节点作为时间同步服务器，其他节点则是作为它的客户端。
[chronyd 时间同步服务 | HappyLadySauce](https://www.happyladysauce.cn/chronyd%E6%97%B6%E9%97%B4%E5%90%8C%E6%AD%A5%E6%9C%8D%E5%8A%A1_Chronyd/)
`Ubuntu` 的 `chrony` 服务不是自带的，需要我们进行手动安装。

```shell
apt install chrony ntpdate -y
```

设置系统时区为**亚洲上海**。

```shell
timedatectl set-timezone "Asia/Shanghai"
```

安装完成后，编辑 `chrony` 的配置文件。

```shell
vim /etc/chrony/chrony.conf
```

在文件最底下添加新的内容，按 `GG` 可以跳转到文件内容底部，然后保存并退出。

```shell
allow all
local
local stratum 10
```

重启 `chrony` 服务

```shell
systemctl restart chronyd
```

##### 客户端配置

同样的下载 `chrony` 服务，因为每一台都执行同样的操作，使用 `MobaX` 的多重执行，可以在多台机器上同时执行相同的命令。

```shell
apt install chrony ntpdate -y
```

设置系统时区为**亚洲上海**。

```shell
timedatectl set-timezone "Asia/Shanghai"
```

编辑 `chony` 配置文件。

```shell
vim /etc/chrony/chrony.conf
```

注释掉原本的 `ntp` 时间同步服务器，增加 `Harbor` 作为时间同步服务器。

```shell
 #pool ntp.ubuntu.com        iburst maxsources 4
 #pool 0.ubuntu.pool.ntp.org iburst maxsources 1
 #pool 1.ubuntu.pool.ntp.org iburst maxsources 1
 #pool 2.ubuntu.pool.ntp.org iburst maxsources 2
 pool 10.10.10.2 iburst
```

重启 `chrony` 服务，并开启 `ntp` 同步。

```shell
systemctl restart chronyd
timedatectl set-ntp yes
```

设置系统计划任务，配置 `crontab` 每 5 分钟同步一次时间。

```shell
vim /etc/crontab
```

增加定时任务

```shell
*/5 *  *  *  * ntpdate 10.10.10.254
```

#### 时间同步验证

在 `Kuboard` 时间同步服务器上执行命令，出现以下内容则配置成功。

```
root@kuboard:~# chronyc clients
Hostname                      NTP   Drop Int IntL Last     Cmd   Drop Int  Last
===============================================================================
K8S-M-Master02                 28      0   6   -    29       0      0   -     -
K8S-M-Node01                   28      0   6   -    31       0      0   -     -
K8S-M-Node01                   28      0   6   -    32       0      0   -     -
K8S-M-Master01                 28      0   6   -    28       0      0   -     -
K8S-M-Master03                 28      0   6   -    31       0      0   -     -
```

所有主机关机打快照，保存为 `主机名-时间同步` 。

```shell
shutdown -h 0
```

关机打完快照后，启动所有主机，进行下一步操作。

### 内核优化配置

```shell
touch /etc/modules-load.d/br_netfilter.conf
echo "br_netfilter" > /etc/modules-load.d/br_netfilter.conf
```

```shell
sysctl -w net.bridge.bridge-nf-call-ip6tables=1
sysctl -w net.bridge.bridge-nf-call-iptables=1
sysctl -w net.ipv4.ip_forward=1
modprobe br_netfilter
sysctl -p
```

### Nginx-Balance 分流

```shell
stream {
    upstream apiserver-6443 {
        server 10.10.10.11:6443 weight=5 max_fails=3 fail_timeout=30s;
        server 10.10.10.12:6443 weight=5 max_fails=3 fail_timeout=30s;
    }
    server {
        listen 6443;
        proxy_pass apiserver-6443;  # 直接使用上游服务器组的名字
    }

    upstream apiserver-9345 {
        server 10.10.10.11:9345 weight=5 max_fails=3 fail_timeout=30s;
        server 10.10.10.12:9345 weight=5 max_fails=3 fail_timeout=30s;
    }
    server {
        listen 9345;
        proxy_pass apiserver-9345;  # 直接使用上游服务器组的名字
    }
}
```

### RKE2 前置环境部署

#### 在线环境安装

#### 安装 Helm

```shell
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
sudo apt-get install apt-transport-https --yes
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm
```

添加包含安装 `Rancher` 的 `Chart` 的 `Helm Chart` 仓库。

```shell
helm repo add rancher-stable https://releases.rancher.com/server-charts/stable
```

在可以连接互联网的系统中，将 `cert-manager` 仓库添加到 `Helm`：

```shell
helm repo add jetstack https://charts.jetstack.io
helm repo update
```

获取 `cert-manager`

```shell
helm fetch jetstack/cert-manager --version=v1.17.1
```

使用包管理工具安装 `kubectl`

```shell
sudo apt-get update -y
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
sudo chmod 644 /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo chmod 644 /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update -y
sudo apt-get install -y kubectl
```

#### RKE2 基础配置和仓库配置

修改 RKE2 的配置文件和镜像仓库。

```shell
mkdir -p /etc/rancher/rke2/
```

REK2 配置文件。

```shell
cat > /etc/rancher/rke2/config.yaml << EOF
token: example
tls-san:
  - huinong.internal
EOF
```

RKE2 镜像仓库配置，上传证书。

```shell
mkdir /opt/ssl/
```

添加系统信任证书。

```shell
cp /opt/ssl/ca.crt /usr/local/share/ca-certificates && update-ca-certificates
```

```shell
cat > /etc/rancher/rke2/registries.yaml << EOF
---
mirrors:
  customreg:
    endpoint:
      - "https//registry.huinong.internal"
configs:
  customreg:
    auth:
      username: admin # 镜像仓库的用户名
      password: Harbor12345 # 镜像仓库的密码
    tls:
      cert_file: /opt/ssl/registry.huinong.internal.crt
      key_file: /opt/ssl/registry.huinong.internal.key
      ca_file: /opt/ssl/ca.crt
EOF
```

### RKE2 Install.sh 脚本安装

#### Server 节点安装

```shell
mkdir /root/rke2-artifacts && cd /root/rke2-artifacts/
curl -OLs https://github.com/rancher/rke2/releases/download/v1.31.7%2Brke2r1/rke2-images.linux-amd64.tar.zst
curl -OLs https://github.com/rancher/rke2/releases/download/v1.31.7%2Brke2r1/rke2.linux-amd64.tar.gz
curl -OLs https://github.com/rancher/rke2/releases/download/v1.31.7%2Brke2r1/sha256sum-amd64.txt
curl -sfL https://get.rke2.io --output install.sh
```

使用该目录运行 install.sh，如下例所示：

```shell
INSTALL_RKE2_ARTIFACT_PATH=/root/rke2-artifacts sh install.sh
```

```shell
systemctl start rke2-server && systemctl enable rke2-server
```

```shell
mkdir ~/.kube && cp /etc/rancher/rke2/rke2.yaml ~/.kube/config
```

#### Worker 节点安装

```shell
systemctl start rke2-agent && systemctl enable rke2-agent
```

### RKE2 Ingress 调整

RKE2 部署完成之后，nginx-ingress 用的是DaemonSet的方式进行部署，并且没有映射到主机IP，所以导致直接通过本机的 80，443 端口访问不了 Igress 。
解决办法：需要修改DaemonSet的配置，添加一个参数：`hostNetwork: true`

```shell
kubectl edit daemonset -n kube-system rke2-ingress-nginx-controller
```

---

## Karmada 部署

### 安装 Karmada CLI 工具

`Karmada` CLI 工具是允许你控制 `Karmada` 控制面的 `Karmada` 命令行工具，`Karmada` CLI 工具分为两种，一种是 `kubectl-karmada` 该工具表现为一个 `kubectl` 插件，另一个工具是 `karmadactl` 这是一个完全专用于 `Karmada` 的 CLI 工具。

#### 安装 Karmadactl

若要安装最新版本，请运行：

```
curl -s https://raw.githubusercontent.com/karmada-io/karmada/master/hack/install-cli.sh | sudo bash
```

你也可以导出 `INSTALL_CLI_VERSION` 环境变量选择要安装的版本。

例如，使用以下命令安装 `1.3.0 karmadactl`：

```
curl -s https://raw.githubusercontent.com/karmada-io/karmada/master/hack/install-cli.sh | sudo INSTALL_CLI_VER
```

#### 安装 kubectl-karmada

要安装最新版本，请运行：

```
curl -s https://raw.githubusercontent.com/karmada-io/karmada/master/hack/install-cli.sh | sudo bash -s kubectl-karmada
```

### 安装 Karmada

首先下载相应 `Karmada` 版本的 `crds.tar.gz` 文件，并下载相应的容器镜像。

```text
# Karmada 离线下载容器镜像列表
registry.k8s.io/kube-apiserver:v1.31.3
registry.k8s.io/etcd:3.5.16-0
docker.io/alpine:3.21.0
docker.io/karmada/karmada-aggregated-apiserver:v1.13.2
registry.k8s.io/kube-controller-manager:v1.31.3
docker.io/karmada/karmada-scheduler:v1.13.2
docker.io/karmada/karmada-webhook:v1.13.2
docker.io/karmada/karmada-controller-manager:v1.13.2
```

运行安装命令

```shell
karmadactl init \
--crds crds.tar.gz \
--kube-image-registry='registry.huinong.internal/karmada' \
--etcd-init-image='registry.huinong.internal/docker.io/alpine@sha256:fa7042902b0e812e73bbee26a6918a6138ccf6d7ecf1746e1488c0bd76cf1f34' \
--karmada-aggregated-apiserver-image='registry.huinong.internal/karmada/karmada/karmada-aggregated-apiserver@sha256:4ff038adc9581a94378b531be3c37684bf682d94136c8aea48f949a5f1f3d3b2' \
--karmada-kube-controller-manager-image='registry.huinong.internal/karmada/kube-controller-manager@sha256:fc71da92458606629f03d62f01762010367347c3d98de8c45576ef60cf85ee4d' \
--karmada-scheduler-image='registry.huinong.internal/karmada/karmada/karmada-scheduler@sha256:0e5da2005cc9e6c54a95ce7ab9e09ea11cfdf5df3b635fb99a0d1ff953cbdd5e' \
--karmada-webhook-image='registry.huinong.internal/karmada/karmada/karmada-webhook@sha256:6a968e72efdd32c733f26b4a6cbfc0e50895d732d59ad8fa36b68fee08932f44' \
--karmada-controller-manager-image='registry.huinong.internal/karmada/karmada/karmada-controller-manager@sha256:04f1924fa45599e7a2b71881ac80cda939d147487d1df9b8a293f99924c0dab3'
```

```
karmadactl init \
--crds crds.tar.gz \
--kube-image-registry='registry.huinong.internal/karmada' \
--karmada-aggregated-apiserver-image='registry.huinong.internal/karmada/karmada/karmada-aggregated-apiserver@sha256:4ff038adc9581a94378b531be3c37684bf682d94136c8aea48f949a5f1f3d3b2' \
--karmada-kube-controller-manager-image='registry.huinong.internal/karmada/kube-controller-manager@sha256:fc71da92458606629f03d62f01762010367347c3d98de8c45576ef60cf85ee4d' \
--karmada-scheduler-image='registry.huinong.internal/karmada/karmada/karmada-scheduler@sha256:0e5da2005cc9e6c54a95ce7ab9e09ea11cfdf5df3b635fb99a0d1ff953cbdd5e' \
--karmada-webhook-image='registry.huinong.internal/karmada/karmada/karmada-webhook@sha256:6a968e72efdd32c733f26b4a6cbfc0e50895d732d59ad8fa36b68fee08932f44' \
--karmada-controller-manager-image='registry.huinong.internal/karmada/karmada/karmada-controller-manager@sha256:04f1924fa45599e7a2b71881ac80cda939d147487d1df9b8a293f99924c0dab3'
```

### 清理 Karmada 资源

```shell
karmadactl delete cluster --all
kubectl delete ns karmada-system
```

### 加入 Karmada 集群

```shell
karmadactl join <registered-cluster-name> \
--kubeconfig /etc/karmada/karmada-apiserver.config \
--karmada-context='karmada-apiserver' \
--cluster-kubeconfig=<kubeconfig> \
--cluster-context='registered-cluster-context'
```

#### Master

```shell
karmadactl join master \
--kubeconfig /etc/karmada/karmada-apiserver.config \
--karmada-context='karmada-apiserver' \
--cluster-kubeconfig=/root/.kube/master.config
```

#### Branch

```
karmadactl join branch \
--kubeconfig /etc/karmada/karmada-apiserver.config \
--karmada-context='karmada-apiserver' \
--cluster-kubeconfig=/root/.kube/branch.config
```

查看集群状态

```shell
kubectl --kubeconfig /etc/karmada/karmada-apiserver.config get cluster
```

为集群打上标签

```bash
kubectl label clusters.cluster.karmada.io master cluster=master
```

## Rancher 部署

下载特定的 Rancher 版本，你可以用 Helm `--version` 参数指定版本，如下：

```
helm fetch rancher-stable/rancher --version=v2.10.3
```

安装 `cert-manager`

```shell
kubectl create namespace cert-manager
```

```shell
kubectl apply -f cert-manager.crds.yaml
```

```shell
helm uninstall cert-manager -n cert-manager
```

```shell
helm install cert-manager ./cert-manager-v1.17.2.tgz \
    --namespace cert-manager \
    --set image.repository=registry.huinong.internal/quay.io/jetstack/cert-manager-controller \
    --set webhook.image.repository=registry.huinong.internal/quay.io/jetstack/cert-manager-webhook \
    --set cainjector.image.repository=registry.huinong.internal/quay.io/jetstack/cert-manager-cainjector \
    --set startupapicheck.image.repository=registry.huinong.internal/quay.io/jetstack/cert-manager-startupapicheck
```

安装 `rancher`

```shell
kubectl create namespace cattle-system
```

```shell
helm install rancher ./rancher-2.11.1.tgz \
--namespace cattle-system \
--set hostname=rancher.huinong.internal \
--set rancherImage=registry.huinong.internal/rancher/rancher \
--set systemDefaultRegistry=registry.huinong.internal \
--set useBundledSystemChart=true
```

如果你使用的是私有 CA 签名的证书，请在 `--set ingress.tls.source=secret` 后加上 `--set privateCA=true`：

```shell
helm install rancher ./rancher-2.10.3.tgz \
--namespace cattle-system \
--set hostname= rancher.huinong.internal\
--set certmanager.version=v1.17.1 \
--set rancherImage=registry.huinong.internal:443/rancher/rancher \
--set systemDefaultRegistry=registry.huinong.internal:443 \ # 设置在 Rancher 中使用的默认私有镜像仓库
--set useBundledSystemChart=true # 使用打包的 Rancher System Chart
```

```shell
NAME: rancher
LAST DEPLOYED: Wed Apr 30 05:29:15 2025
NAMESPACE: cattle-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
Rancher Server has been installed.

NOTE: Rancher may take several minutes to fully initialize. Please standby while Certificates are being issued, Containers are started and the Ingress rule comes up.

Check out our docs at https://rancher.com/docs/

If you provided your own bootstrap password during installation, browse to https://rancher.huinong.internal to get started.

If this is the first time you installed Rancher, get started by running this command and clicking the URL it generates:

echo https://rancher.huinong.internal/dashboard/?setup=$(kubectl get secret --namespace cattle-system bootstrap-secret -o go-template='{{.data.bootstrapPassword|base64decode}}')

To get just the bootstrap password on its own, run:

kubectl get secret --namespace cattle-system bootstrap-secret -o go-template='{{.data.bootstrapPassword|base64decode}}{{ "\n" }}'

```

## Higress 离线部署

在线下载 helm 软件包

```shell
helm repo add higress.io https://higress.cn/helm-charts
helm fetch higress.io/higress --untar
```

上传并解压软件包，进入目录执行安装命令

```shell
helm install higress -n higress-system . --create-namespace --render-subchart-notes
```

删除命令

```shell
helm uninstall higress -n higress-system
```

修改 higress 监听本机端口。

```shell
helm upgrade higress -n higress-system . --reuse-values --set higress-core.gateway.hostNetwork=true
```

安装 Istio CRD

```bash
# 创建 istio-system 命名空间（如果不存在）
kubectl create namespace istio-system

# 使用 Helm 安装 Istio Base（CRDs）
helm install istio-base base-1.26.1.tgz -n istio-system
```

使用 Istio CRD

```bash
helm upgrade higress -n higress-system . --set global.enableIstioAPI=true --reuse-values
```
