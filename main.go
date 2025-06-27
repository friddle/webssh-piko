package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"webssh-piko/config"
	"webssh-piko/controller"

	pikoconfig "github.com/andydunstall/piko/agent/config"
	"github.com/andydunstall/piko/agent/reverseproxy"
	"github.com/andydunstall/piko/client"
	"github.com/andydunstall/piko/pkg/log"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
)

//go:embed web/dist/*
var f embed.FS

// 全局变量

// FindAvailablePort 查找可用端口
func FindAvailablePort() int {
	startPort := 8080
	for port := startPort; port < startPort+100; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return startPort // 如果都不可用，返回默认端口
}

// isPortAvailable 检查端口是否可用
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func main() {
	rootCmd := MakeMainCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func MakeMainCmd() *cobra.Command {
	var (
		name     string
		remote   string
		savePass bool
		username string
		password string
		timeout  int
		debug    bool
	)

	cmd := &cobra.Command{
		Use:   "webssh-piko",
		Short: "webssh-piko 客户端 - 基于终端的远程协助工具",
		Long: `webssh-piko 是一个基于终端的高效远程协助工具，集成了 webssh 和 piko 服务。
专为复杂网络环境下的远程协助而设计，避免传统远程桌面对高带宽的依赖，也无需复杂的网络配置和外网地址。

支持的操作系统:
  - Linux (默认使用 bash)
  - macOS (默认使用 zsh)  
  - Windows (默认使用 powershell)

使用示例:
  webssh-piko --name=my-server --remote=192.168.1.100:8088                    # 连接到远程 piko 服务器
  webssh-piko --name=client1 --remote=piko.example.com:8022  --username=admin --password=123456
  `,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 显示版本信息
			// 创建配置
			cfg := &config.Config{
				Name:     name,
				Remote:   remote,
				SavePass: savePass,
				Username: username,
				Password: password,
				Timeout:  timeout,
			}

			// 验证配置
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("配置验证失败: %v", err)
			}

			// 创建服务管理器并启动
			sm := NewServiceManager(cfg)
			return sm.Start()
		},
	}

	// 添加命令行参数
	cmd.Flags().StringVar(&name, "name", "", "piko 客户端标识名称")
	cmd.Flags().StringVar(&remote, "remote", "", "远程 piko 服务器地址 (格式: host:port)")
	cmd.Flags().BoolVar(&savePass, "save-pass", false, "是否保存密码")
	cmd.Flags().StringVar(&username, "username", "", "用户名")
	cmd.Flags().StringVar(&password, "password", "", "密码")
	cmd.Flags().IntVar(&timeout, "timeout", 30, "超时时间（秒）")
	cmd.Flags().BoolVar(&debug, "debug", false, "启用调试模式")

	// 添加别名
	cmd.Flags().StringVarP(&name, "n", "n", "", "piko 客户端标识名称 (简写)")
	cmd.Flags().StringVarP(&remote, "r", "r", "", "远程 piko 服务器地址 (简写)")
	cmd.Flags().StringVarP(&username, "u", "u", "", "用户名 (简写)")
	cmd.Flags().StringVarP(&password, "p", "p", "", "密码 (简写)")

	// 设置必需参数
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("remote")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("webssh-piko 版本: 1.0.0\n")
		},
	}
	cmd.AddCommand(versionCmd)
	return cmd
}

// ServiceManager 服务管理器
type ServiceManager struct {
	config *config.Config
	ctx    context.Context
	cancel context.CancelFunc
}

// NewServiceManager 创建新的服务管理器
func NewServiceManager(cfg *config.Config) *ServiceManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServiceManager{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动所有服务
func (sm *ServiceManager) Start() error {
	fmt.Printf("🚀 启动 webssh-piko 客户端\n")
	fmt.Printf("客户端名称: %s\n", sm.config.Name)
	fmt.Printf("远程服务器: %s\n", sm.config.Remote)

	// 自动分配可用端口（如果未指定）
	if sm.config.ServerPort == 0 {
		sm.config.ServerPort = FindAvailablePort()
	}
	fmt.Printf("本地监听端口: %d\n", sm.config.ServerPort)

	// 使用 oklog/run 启动服务
	return sm.startServices()
}

// startServices 使用 oklog/run 启动所有服务
func (sm *ServiceManager) startServices() error {
	var g run.Group

	// 启动 piko 服务
	g.Add(func() error {
		err := sm.startPiko()
		if err != nil {
			fmt.Printf("启动piko失败:%v\n", err)
		}
		return err
	}, func(error) {
		// piko 服务会在 context 取消时自动停止
	})

	// 启动 webssh 服务
	g.Add(func() error {
		return sm.startWebServer()
	}, func(error) {
		// webssh 服务会在 context 取消时自动停止
	})

	// 信号处理
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		// 根据操作系统选择不同的信号
		if runtime.GOOS == "windows" {
			// Windows 支持 Ctrl+C (SIGINT) 和 Ctrl+Break
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		} else {
			// Unix-like 系统支持更多信号
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		}
		select {
		case sig := <-c:
			fmt.Printf("\n🛑 收到停止信号 %v，正在关闭服务...\n", sig)
			return nil
		case <-sm.ctx.Done():
			return sm.ctx.Err()
		}
	}, func(error) {
		sm.cancel()
	})

	// 24小时超时
	g.Add(func() error {
		ctx, cancel := context.WithTimeout(sm.ctx, 24*time.Hour)
		defer cancel()
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("\n⏰ 服务运行时间达到24小时，正在停止...\n")
		}
		return ctx.Err()
	}, func(error) {
		sm.cancel()
	})

	fmt.Printf("✅ 服务启动成功！\n")
	fmt.Printf("🌐 访问地址: http://localhost:%d\n", sm.config.ServerPort)
	// 运行所有服务
	return g.Run()
}

// Stop 停止所有服务
func (sm *ServiceManager) Stop() {
	sm.cancel()
	fmt.Printf("✅ 服务已停止\n")
}

// startPiko 启动piko服务
func (sm *ServiceManager) startPiko() error {
	// 创建 piko 配置
	fmt.Printf("启动piko中\n")
	remote := sm.config.Remote
	if strings.HasPrefix(remote, "http") {
		remote = sm.config.Remote
	} else {
		remote = fmt.Sprintf("http://%s", sm.config.Remote)
	}
	conf := &pikoconfig.Config{
		Connect: pikoconfig.ConnectConfig{
			URL:     remote,
			Timeout: 30 * time.Second,
		},
		Listeners: []pikoconfig.ListenerConfig{
			{
				EndpointID: sm.config.Name,
				Protocol:   pikoconfig.ListenerProtocolHTTP,
				Addr:       fmt.Sprintf("127.0.0.1:%d", sm.config.ServerPort),
				AccessLog:  false,
				Timeout:    30 * time.Second,
				TLS:        pikoconfig.TLSConfig{},
			},
		},
		Log: log.Config{
			Level:      "info",
			Subsystems: []string{},
		},
		GracePeriod: 30 * time.Second,
	}

	// 创建日志记录器
	logger, err := log.NewLogger("info", []string{})
	if err != nil {
		return fmt.Errorf("创建日志记录器失败: %v", err)
	}

	// 验证配置
	if err := conf.Validate(); err != nil {
		return fmt.Errorf("piko 配置验证失败: %v", err)
	}

	// 解析连接 URL
	connectURL, err := url.Parse(conf.Connect.URL)
	if err != nil {
		return fmt.Errorf("解析连接 URL 失败: %v", err)
	}

	// 创建上游客户端
	upstream := &client.Upstream{
		URL:       connectURL,
		TLSConfig: nil, // 不使用 TLS
		Logger:    logger.WithSubsystem("client"),
	}

	// 为每个监听器创建连接
	for _, listenerConfig := range conf.Listeners {
		fmt.Printf("正在连接到端点: %s\n", listenerConfig.EndpointID)

		ln, err := upstream.Listen(sm.ctx, listenerConfig.EndpointID)
		if err != nil {
			return fmt.Errorf("监听端点失败 %s: %v", listenerConfig.EndpointID, err)
		}

		fmt.Printf("成功连接到端点: %s\n", listenerConfig.EndpointID)

		// 创建 HTTP 代理服务器，传入正确的配置而不是 nil
		metrics := reverseproxy.NewMetrics("proxy")
		server := reverseproxy.NewServer(listenerConfig, metrics, logger)
		if server == nil {
			return fmt.Errorf("创建 HTTP 代理服务器失败")
		}
		// 启动代理服务器
		go func() {
			if err := server.Serve(ln); err != nil {
				fmt.Printf("代理服务器运行错误: %v\n", err)
			}
		}()
	}
	return nil
}

// startWebServer 启动web服务
func (sm *ServiceManager) startWebServer() error {
	server := gin.Default()
	server.SetTrustedProxies(nil)
	server.Use(gzip.Gzip(gzip.DefaultCompression))
	ctxRouter := server.Group("/" + sm.config.Name)

	sm.staticRouter(ctxRouter)
	ctxRouter.GET("/term", func(c *gin.Context) {
		controller.TermWs(c, time.Duration(sm.config.Timeout)*time.Second)
	})
	ctxRouter.GET("/check", func(c *gin.Context) {
		responseBody := controller.CheckSSH(c)
		responseBody.Data = map[string]interface{}{
			"savePass": sm.config.SavePass,
		}
		c.JSON(200, responseBody)
	})
	file := ctxRouter.Group("/file")
	{
		file.GET("/list", func(c *gin.Context) {
			c.JSON(200, controller.FileList(c))
		})
		file.GET("/download", func(c *gin.Context) {
			controller.DownloadFile(c)
		})
		file.POST("/upload", func(c *gin.Context) {
			c.JSON(200, controller.UploadFile(c))
		})
		file.GET("/progress", func(c *gin.Context) {
			controller.UploadProgressWs(c)
		})
	}
	return server.Run(fmt.Sprintf(":%d", sm.config.ServerPort))
}

// staticRouter 设置静态文件路由
func (sm *ServiceManager) staticRouter(router *gin.RouterGroup) {
	if sm.config.Password != "" {
		accountList := map[string]string{
			sm.config.Username: sm.config.Password,
		}
		authorized := router.Group("/", gin.BasicAuth(accountList))
		authorized.GET("", func(c *gin.Context) {
			indexHTML, _ := f.ReadFile("web/dist/" + "index.html")
			// 注入子目录路径到HTML中
			htmlContent := string(indexHTML)
			htmlContent = strings.Replace(htmlContent,
				"<head>",
				fmt.Sprintf("<head><script>window.SUB_PATH = '%s';</script>", sm.config.Name),
				1)
			c.Writer.Write([]byte(htmlContent))
		})
	} else {
		router.GET("/", func(c *gin.Context) {
			indexHTML, _ := f.ReadFile("web/dist/" + "index.html")
			// 注入子目录路径到HTML中
			htmlContent := string(indexHTML)
			htmlContent = strings.Replace(htmlContent,
				"<head>",
				fmt.Sprintf("<head><script>window.SUB_PATH = '%s';</script>", sm.config.Name),
				1)
			c.Writer.Write([]byte(htmlContent))
		})
	}
	staticFs, _ := fs.Sub(f, "web/dist/static")
	router.StaticFS("/static", http.FS(staticFs))
}
