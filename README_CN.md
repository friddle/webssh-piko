# webssh-piko
webssh: https://github.com/Jrohy/webssh
piko: https://github.com/andydunstall/piko

一个基于终端的高效远程协助工具，集成了 webssh 和 piko 服务。专为复杂网络环境下的远程协助而设计，避免传统远程桌面对高带宽的依赖，也无需复杂的网络配置和外网地址。

## 项目特点

- 🚀 **轻量级**: 基于终端的远程协助，资源占用低
- 🌐 **网络友好**: 支持内网穿透，无需公网IP
- 🔧 **简单部署**: Docker 一键部署，配置简单
- 🔒 **安全可靠**: 基于 SSH 协议，支持用户认证
- 📱 **跨平台**: 支持 Linux、macOS、Windows

## 架构说明

```
客户端 (webssh-piko client) 
    ↓ 本地Shell
webssh服务
    ↓ HTTP访问
浏览器终端
```

## 快速开始

### 服务端部署

1. **使用 Docker Compose 部署**

```yaml
# docker-compose.yaml
version: "3.8"
services:
  piko:
    image: ghcr.io/friddle/webssh-piko-server:latest
    container_name: webssh-piko-server
    environment:
      - PIKO_UPSTREAM_PORT=8022
      - LISTEN_PORT=8088
    ports:
      - "8022:8022"
      - "8088:8088"
    restart: unless-stopped
```

或直接使用 Docker：

```bash
docker run -ti --network=host --rm --name=piko-server ghcr.io/friddle/webssh-piko-server
```

2. **启动服务**

```bash
docker-compose up -d
```

### 客户端使用

#### Linux 客户端

```bash
# 下载客户端
wget https://github.com/friddle/webssh-piko/releases/download/v1.0.1/webssh-piko-linux-amd64 -O ./websshp
chmod +x ./websshp

./websshp --name=local --remote=192.168.1.100:8088
```

#### Windows客户端

```cmd
# 下载客户端 (PowerShell)
Invoke-WebRequest -Uri "https://github.com/friddle/webssh-piko/releases/download/v1.0.1/webssh-piko-windows-amd64.exe" -OutFile "websshp.exe"

# 带认证的访问
websshp.exe --name=local --remote=192.168.1.100:8088 --username=admin --password=123456
```

#### macOS 客户端

```bash
# 下载客户端
curl -L -o websshp https://github.com/friddle/webssh-piko/releases/download/v1.0.1/webssh-piko-darwin-amd64
chmod +x ./websshp

./websshp --name=local --remote=192.168.1.100:8088
```

![客户端启动截图](screenshot/start_cli.png)
![Web界面截图](screenshot/webui.png)

## 访问方式

当客户端启动后，通过以下地址访问对应的终端：
```
http://主机服务器IP:端口/客户端名称
```

例如：
- 服务端监听的地址: `192.168.1.100:8088` (服务端IP和NGINX)
- 客户端名称: `local`
- 访问地址: `http://192.168.1.100:8088/local`

## 配置说明

### 客户端参数

| 参数 | 简写 | 说明 | 默认值 | 必填 |
|------|------|------|--------|------|
| `--name` | `-n` | piko 客户端标识名称 | - | ✅ |
| `--remote` | `-r` | 远程 piko 服务器地址 (格式: host:port) | - | ✅ |
| `--username` | `-u` | 用户名 | - | ❌ |
| `--password` | `-p` | 密码 | - | ❌ |
| `--timeout` | - | 超时时间（秒） | 30 | ❌ |
| `--debug` | - | 启用调试模式 | false | ❌ |

### 服务端环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PIKO_UPSTREAM_PORT` | Piko 上游端口 | 8022 |
| `LISTEN_PORT` | HTTP 监听端口 | 8088 |
