package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/bamzest/xiaomi-router-shell-enabler/pkg/client"
	"github.com/bamzest/xiaomi-router-shell-enabler/pkg/logger"
	"github.com/bamzest/xiaomi-router-shell-enabler/pkg/utils"
	"github.com/bamzest/xiaomi-router-shell-enabler/pkg/version"
)

func main() {
	// 定义命令行参数
	host := flag.String("host", "", "路由器IP地址")
	password := flag.String("password", "", "路由器管理密码") // 修改为密码参数
	model := flag.String("model", "", "路由器型号")
	listModels := flag.Bool("list", false, "列出所有支持的路由器型号")
	verbose := flag.Bool("verbose", false, "显示详细日志")
	showVersion := flag.Bool("version", false, "显示版本信息")
	serialNumber := flag.String("sn", "", "路由器序列号，用于计算SSH密码")
	calcPasswordOnly := flag.Bool("calc-password", false, "仅计算并显示SSH密码")
	execCommand := flag.String("exec", "", "执行自定义命令")
	enableShell := flag.Bool("enable_shell", false, "启用SSH和Telnet")
	disableShell := flag.Bool("disable_shell", false, "关闭SSH和Telnet")
	shellStatus := flag.Bool("shell_status", false, "检查SSH和Telnet的开启状态")
	
	// 兼容旧版本的 token 参数
	token := flag.String("token", "", "[已弃用] 路由器登录Token (请使用 -password 参数)")
	
	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Xiaomi Router Shell Enabler v%s\n\n", version.Version)
		fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s -model redmi_ax5400pro -host 192.168.31.1 -password YOUR_PASSWORD -enable_shell -verbose\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -model redmi_ax5400pro -host 192.168.31.1 -password YOUR_PASSWORD -disable_shell -verbose\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -model redmi_ax5400pro -host 192.168.31.1 -password YOUR_PASSWORD -shell_status -verbose\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -sn 39668/A1ZZ38217 -calc-password\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -model redmi_ax5400pro -host 192.168.31.1 -password YOUR_PASSWORD -exec \"cat /etc/passwd\" -verbose\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -version\n", os.Args[0])
	}
	
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Xiaomi Router Shell Enabler v%s\n", version.Version)
		fmt.Printf("构建时间: %s\n", version.BuildTime)
		fmt.Printf("提交哈希: %s\n", version.GitCommit)
		return
	}

	// 配置日志级别
	if *verbose {
		logger.SetLevel(logger.LevelDebug)
		logger.Debug("调试模式已启用")
	} else {
		logger.SetLevel(logger.LevelInfo)
	}

	// 显示支持的型号列表
	if *listModels {
		fmt.Println("支持的路由器型号:")
		for _, m := range client.GetSupportedModels() {
			fmt.Printf("- %s\n", m)
		}
		return
	}

	// 如果提供了序列号，计算SSH密码
	if *serialNumber != "" {
		sshPassword := utils.CalculateSSHPassword(*serialNumber)
		if sshPassword != "" {
			fmt.Printf("序列号: %s\n", *serialNumber)
			fmt.Printf("计算得到的SSH密码: %s\n", sshPassword)
			
			// 如果只是计算密码，则退出
			if *calcPasswordOnly {
				return
			}
		} else {
			fmt.Println("无法计算SSH密码，请检查序列号是否正确。")
			if *calcPasswordOnly {
				os.Exit(1)
			}
		}
	}

	// 检查必需参数
	if (*host == "" || (*password == "" && *token == "") || *model == "") && !*calcPasswordOnly {
		if *serialNumber == "" || !*calcPasswordOnly {
			fmt.Println("错误: 必须提供路由器IP地址、管理密码和型号")
			fmt.Println("用法示例: xiaomi-router-shell-enabler -model redmi_ax5400pro -host 192.168.31.1 -password YOUR_PASSWORD -enable_shell")
			fmt.Println("或使用 -h 查看帮助信息")
			os.Exit(1)
		}
	}

	// 如果只是计算密码并且已经完成，则退出
	if *calcPasswordOnly && *serialNumber != "" {
		return
	}

	// 验证主机地址格式
	if !strings.HasPrefix(*host, "http://") && !strings.HasPrefix(*host, "https://") {
		// 如果用户没有提供协议前缀，默认添加http://
		*host = strings.TrimPrefix(*host, "http://")
		*host = strings.TrimPrefix(*host, "https://")
		// 移除可能的尾部斜杠
		*host = strings.TrimSuffix(*host, "/")
		logger.Debug("使用主机地址: %s", *host)
	}

	// 如果用户提供了 token 而不是 password，发出警告
	routerPassword := *password
	if routerPassword == "" && *token != "" {
		logger.Warn("-token 参数已弃用，请使用 -password 参数")
		routerPassword = *token
	}

	logger.Debug("连接信息: 主机=%s, 型号=%s", *host, *model)

	// 创建路由器客户端
	routerClient, err := client.NewRouterClient(*host, routerPassword, *model)
	if err != nil {
		logger.Error("%v", err)
		fmt.Println("支持的型号: ", client.GetSupportedModels())
		os.Exit(1)
	}

	// 处理不同的操作模式
	if *shellStatus {
		// 检查SSH和Telnet状态
		logger.Info("检查 %s 路由器的SSH和Telnet状态...", *model)
		status, details, err := routerClient.CheckShellStatus()
		if err != nil {
			logger.Error("检查状态失败: %v", err)
			os.Exit(1)
		}
		
		// 显示状态摘要
		if status {
			logger.Info("SSH和Telnet服务状态: 已启用并可访问")
		} else {
			logger.Warn("SSH和Telnet服务状态: 未完全启用或不可访问")
		}
		
		// 显示详细状态信息
		fmt.Println("\n详细状态信息:")
		fmt.Println(details)
		
	} else if *execCommand != "" {
		// 执行自定义命令模式
		logger.Info("执行自定义命令: %s", *execCommand)
		err := routerClient.ExecuteCustomCommand(*execCommand)
		if err != nil {
			logger.Error("执行命令失败: %v", err)
			os.Exit(1)
		}
		logger.Info("命令执行完成")
	} else if *enableShell {
		// 启用SSH和Telnet模式
		logger.Info("开始为 %s 路由器启用SSH和Telnet...", *model)
		err = routerClient.EnableSSH()
		if err != nil {
			logger.Error("启用SSH和Telnet失败: %v", err)
			os.Exit(1)
		}
		
		// 如果提供了序列号，显示SSH连接信息
		if *serialNumber != "" {
			sshPassword := utils.CalculateSSHPassword(*serialNumber)
			if sshPassword != "" {
				fmt.Printf("\n登录凭据:\n")
				fmt.Printf("  用户名: root\n")
				fmt.Printf("  密码: %s\n", sshPassword)
				
				// 显示SSH连接命令
				fmt.Printf("\n连接命令:\n")
				fmt.Printf("  SSH: %s\n", routerClient.GetSSHCommand())
				fmt.Printf("  Telnet: %s\n", routerClient.GetTelnetCommand())
			}
		} else {
			fmt.Printf("\n提示: 如果您知道路由器序列号，可以使用 -sn 参数计算SSH密码\n")
			fmt.Printf("例如: %s -sn YOUR_SERIAL_NUMBER -calc-password\n", os.Args[0])
		}
	} else if *disableShell {
		// 关闭SSH和Telnet模式
		logger.Info("开始为 %s 路由器关闭SSH和Telnet...", *model)
		err = routerClient.DisableSSH()
		if err != nil {
			logger.Error("关闭SSH和Telnet失败: %v", err)
			os.Exit(1)
		}
		logger.Info("SSH和Telnet关闭操作完成")
	} else {
		// 如果没有指定具体操作，显示帮助信息
		fmt.Println("请指定要执行的操作: -enable_shell, -disable_shell, -shell_status 或 -exec 命令")
		fmt.Println("使用 -h 查看帮助信息")
		os.Exit(1)
	}
}