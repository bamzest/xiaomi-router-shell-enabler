package routers

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bamzest/xiaomi-router-shell-enabler/pkg/logger"
)

// BaseRouterClient 路由器客户端基类
type BaseRouterClient struct {
	Host  string
	Token string
	Model string
}

// ShellStatusResult 存储Shell状态检查的结果
type ShellStatusResult struct {
	SSHEnabled     bool // API返回的SSH状态
	TelnetEnabled  bool // API返回的Telnet状态
	SSHPortOpen    bool // 22端口是否开放
	TelnetPortOpen bool // 23端口是否开放
}

// HTTP GET请求
func (c *BaseRouterClient) Get(apiPath string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/cgi-bin/luci/;stok=%s/%s", c.Host, c.Token, apiPath)

	logger.Debug("发送GET请求: %s", url)

	client := &http.Client{
		Timeout: 30 * time.Second, // 增加超时时间到30秒
	}
	resp, err := client.Get(url)
	if err != nil {
		logger.Debug("GET请求失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Debug("读取响应失败: %v", err)
		return nil, err
	}

	logger.Debug("收到GET响应: [状态码: %d] %s", resp.StatusCode, string(body))
	return body, nil
}

// HTTP POST请求
func (c *BaseRouterClient) Post(apiPath string, data string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/cgi-bin/luci/;stok=%s/%s", c.Host, c.Token, apiPath)

	logger.Debug("发送POST请求: %s", url)
	logger.Debug("POST请求数据: %s", data)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
	if err != nil {
		logger.Debug("创建POST请求失败: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 30 * time.Second, // 增加超时时间到30秒
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Debug("POST请求失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Debug("读取响应失败: %v", err)
		return nil, err
	}

	logger.Debug("收到POST响应: [状态码: %d] %s", resp.StatusCode, string(body))
	return body, nil
}

// CheckPortOpen 检查指定端口是否开放
func (c *BaseRouterClient) CheckPortOpen(port int) bool {
	address := fmt.Sprintf("%s:%d", c.Host, port)
	logger.Debug("检查端口是否开放: %s", address)

	// 设置较短的超时时间，避免长时间等待
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		logger.Debug("端口 %d 未开放: %v", port, err)
		return false
	}
	defer conn.Close()

	logger.Debug("端口 %d 已开放", port)
	return true
}

// SetSystemTime 设置系统时间 (通用实现)
func (c *BaseRouterClient) SetSystemTime() error {
	logger.Info("设置系统时间...")

	now := time.Now()
	timeStr := now.Format("2006-1-2 15:4:5")

	apiPath := fmt.Sprintf("api/misystem/set_sys_time?time=%s&timezone=CST-8", timeStr)
	respBody, err := c.Get(apiPath)
	if err != nil {
		return fmt.Errorf("设置系统时间失败: %v", err)
	}

	// 简单检查响应是否包含成功信息
	responseStr := string(respBody)
	if !strings.Contains(responseStr, `"code":0`) && !strings.Contains(responseStr, `"success":true`) {
		logger.Warn("设置系统时间可能失败，响应: %s", responseStr)
		// 继续执行，因为这可能只是响应格式不符合预期，但操作可能已成功
	}

	logger.Info("系统时间已设置")
	return nil
}

// GetSSHCommand 获取适用于此型号的SSH连接命令 (基本实现，子类可覆写)
func (c *BaseRouterClient) GetSSHCommand() string {
	// 默认的SSH连接命令
	return fmt.Sprintf("ssh root@%s", c.Host)
}

// GetTelnetCommand 获取适用于此型号的Telnet连接命令 (基本实现，子类可覆写)
func (c *BaseRouterClient) GetTelnetCommand() string {
	// 默认的Telnet连接命令
	return fmt.Sprintf("telnet %s", c.Host)
}

// ExecuteCustomCommand 执行自定义命令 (需要子类实现)
func (c *BaseRouterClient) ExecuteCustomCommand(command string) error {
	return fmt.Errorf("此路由器型号不支持执行自定义命令")
}

// EnableSSH 启用SSH (需要子类实现)
func (c *BaseRouterClient) EnableSSH() error {
	return fmt.Errorf("此路由器型号不支持启用SSH")
}

// DisableSSH 关闭SSH (需要子类实现)
func (c *BaseRouterClient) DisableSSH() error {
	return fmt.Errorf("此路由器型号不支持关闭SSH")
}

// VerifySSHStatus 验证SSH状态 (需要子类实现)
func (c *BaseRouterClient) VerifySSHStatus() (bool, error) {
	return false, fmt.Errorf("此路由器型号不支持验证SSH状态")
}

// CheckShellStatus 检查SSH和Telnet状态 (基本实现，子类可覆写)
func (c *BaseRouterClient) CheckShellStatus() (bool, string, error) {
	logger.Info("检查SSH和Telnet状态...")
	
	// 创建结果结构体
	result := &ShellStatusResult{
		SSHEnabled:     false,
		TelnetEnabled:  false,
		SSHPortOpen:    false,
		TelnetPortOpen: false,
	}
	
	// 检查SSH端口(22)是否开放
	logger.Info("检查SSH端口(22)是否开放...")
	result.SSHPortOpen = c.CheckPortOpen(22)
	
	// 检查Telnet端口(23)是否开放
	logger.Info("检查Telnet端口(23)是否开放...")
	result.TelnetPortOpen = c.CheckPortOpen(23)
	
	// 生成详细状态报告
	var detailsBuilder strings.Builder
	detailsBuilder.WriteString("SSH状态:\n")
	detailsBuilder.WriteString(fmt.Sprintf("  - 端口22开放: %v\n", result.SSHPortOpen))
	detailsBuilder.WriteString("\nTelnet状态:\n")
	detailsBuilder.WriteString(fmt.Sprintf("  - 端口23开放: %v\n", result.TelnetPortOpen))
	
	// 如果端口开放，显示连接命令
	detailsBuilder.WriteString("\n连接信息:\n")
	if result.SSHPortOpen {
		detailsBuilder.WriteString(fmt.Sprintf("  - SSH连接命令: %s\n", c.GetSSHCommand()))
	}
	if result.TelnetPortOpen {
		detailsBuilder.WriteString(fmt.Sprintf("  - Telnet连接命令: %s\n", c.GetTelnetCommand()))
	}
	
	// 基本实现只检查端口，子类可以覆写此方法以提供更详细的状态检查
	overallStatus := result.SSHPortOpen || result.TelnetPortOpen
	
	return overallStatus, detailsBuilder.String(), nil
}