package routers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bamzest/xiaomi-router-shell-enabler/pkg/logger"
)

// AX5400ProClient AX5400Pro路由器客户端
type AX5400ProClient struct {
	BaseRouterClient
}

// APIResponse 小米路由器API通用响应结构
type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 任务时间缓存文件路径
const taskTimeCacheFile = ".task_time_cache"

// NewAX5400ProClient 创建AX5400Pro客户端
func NewAX5400ProClient(host, token string) *AX5400ProClient {
	return &AX5400ProClient{
		BaseRouterClient: BaseRouterClient{
			Host:  host,
			Token: token,
			Model: "redmi_ax5400pro",
		},
	}
}

// GetSSHCommand 返回适用于此型号的SSH连接命令
func (c *AX5400ProClient) GetSSHCommand() string {
	return fmt.Sprintf("ssh -o HostKeyAlgorithms=+ssh-rsa -o PubkeyAcceptedKeyTypes=+ssh-rsa root@%s", c.Host)
}

// GetTelnetCommand 返回适用于此型号的Telnet连接命令
func (c *AX5400ProClient) GetTelnetCommand() string {
	return fmt.Sprintf("telnet %s", c.Host)
}

// 获取并递增任务时间
func getNextTaskTime() string {
	hour, minute := -1, -1 // 初始化为-1表示未设置

	// 获取程序执行目录
	exePath, err := os.Executable()
	if err != nil {
		logger.Debug("无法获取程序执行路径: %v，将使用当前目录", err)
		exePath, _ = os.Getwd()
	}

	// 使用程序所在目录
	programDir := filepath.Dir(exePath)
	cachePath := filepath.Join(programDir, taskTimeCacheFile)
	logger.Debug("任务时间缓存文件路径: %s", cachePath)

	// 尝试读取缓存文件
	content, err := ioutil.ReadFile(cachePath)
	if err == nil {
		// 文件存在，解析内容
		parts := strings.Split(string(content), ":")
		if len(parts) == 2 {
			hourVal, hourErr := strconv.Atoi(parts[0])
			minVal, minErr := strconv.Atoi(parts[1])

			if hourErr == nil && minErr == nil {
				hour = hourVal
				minute = minVal

				// 递增分钟
				minute++
				if minute >= 60 {
					minute = 0
					hour++
					if hour >= 24 {
						hour = 0
					}
				}
			} else {
				logger.Debug("解析缓存文件失败: %v, %v，将使用当前时间+1分钟", hourErr, minErr)
				hour = -1 // 重置为未设置状态
			}
		} else {
			logger.Debug("缓存文件格式不正确，将使用当前时间+1分钟")
		}
	} else {
		logger.Debug("读取缓存文件失败: %v，将使用当前时间+1分钟", err)
	}

	// 如果未设置时间（首次运行或缓存文件读取失败），使用当前时间+1分钟
	if hour == -1 || minute == -1 {
		now := time.Now().Add(1 * time.Minute)
		hour = now.Hour()
		minute = now.Minute()
		logger.Debug("使用当前时间+1分钟: %d:%d", hour, minute)
	}

	// 将新的时间写入缓存文件
	newTimeStr := fmt.Sprintf("%d:%d", hour, minute)
	err = ioutil.WriteFile(cachePath, []byte(newTimeStr), 0644)
	if err != nil {
		logger.Debug("写入缓存文件失败: %v", err)

		// 如果写入失败，尝试在当前工作目录创建
		currentDir, _ := os.Getwd()
		fallbackPath := filepath.Join(currentDir, taskTimeCacheFile)
		logger.Debug("尝试使用备用路径: %s", fallbackPath)

		err = ioutil.WriteFile(fallbackPath, []byte(newTimeStr), 0644)
		if err != nil {
			logger.Debug("写入备用缓存文件也失败: %v", err)
		}
	}

	logger.Debug("使用任务时间: %s", newTimeStr)
	return newTimeStr
}

// 设置智能控制器任务
func (c *AX5400ProClient) SetSmartControllerTask(command, taskTime string) error {
	payload := fmt.Sprintf(`{"command":"scene_setting","name":"%s","action_list":[{"thirdParty":"xmrouter","delay":17,"type":"wan_block","payload":{"command":"wan_block","mac":"00:00:00:00:00:00"}}],"launch":{"timer":{"time":"%s","repeat":"0","enabled":true}}}`, command, taskTime)
	encodedPayload := url.QueryEscape(payload)

	data := fmt.Sprintf("payload=%s", encodedPayload)

	logger.Debug("设置智能控制器任务")
	logger.Debug("原始Payload: %s", payload)
	logger.Debug("URL编码后Payload: %s", encodedPayload)

	respBody, err := c.Post("api/xqsmarthome/request_smartcontroller", data)
	if err != nil {
		logger.Debug("设置智能控制器任务请求失败: %v", err)
		return err
	}

	// 解析响应
	var resp APIResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		logger.Debug("解析响应失败: %v, 原始响应: %s", err, string(respBody))
		return fmt.Errorf("解析响应失败: %v", err)
	}

	logger.Debug("设置智能控制器任务响应: code=%d, msg=%s", resp.Code, resp.Msg)

	// 检查响应状态
	if resp.Code != 0 {
		return fmt.Errorf("API错误: %s (代码: %d)", resp.Msg, resp.Code)
	}

	return nil
}

// 启动智能控制器任务
func (c *AX5400ProClient) StartSmartControllerTask(taskTime string, week int) error {
	payload := fmt.Sprintf(`{"command":"scene_start_by_crontab","time":"%s","week":%d}`, taskTime, week)
	encodedPayload := url.QueryEscape(payload)

	data := fmt.Sprintf("payload=%s", encodedPayload)

	logger.Debug("启动智能控制器任务")
	logger.Debug("原始Payload: %s", payload)
	logger.Debug("URL编码后Payload: %s", encodedPayload)

	respBody, err := c.Post("api/xqsmarthome/request_smartcontroller", data)
	if err != nil {
		logger.Debug("启动智能控制器任务请求失败: %v", err)
		return err
	}

	// 解析响应
	var resp APIResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		logger.Debug("解析响应失败: %v, 原始响应: %s", err, string(respBody))
		return fmt.Errorf("解析响应失败: %v", err)
	}

	logger.Debug("启动智能控制器任务响应: code=%d, msg=%s", resp.Code, resp.Msg)

	// 检查响应状态
	if resp.Code != 0 {
		return fmt.Errorf("API错误: %s (代码: %d)", resp.Msg, resp.Code)
	}

	return nil
}

// ExecuteCustomCommand 执行自定义命令
func (c *AX5400ProClient) ExecuteCustomCommand(command string) error {
	logger.Info("准备执行命令: %s", command)

	// 格式化命令，确保它能在路由器上正确执行
	formattedCommand := fmt.Sprintf("'$(%s)'", command)

	// 获取任务时间
	taskTime := getNextTaskTime()

	// 设置任务
	err := c.SetSmartControllerTask(formattedCommand, taskTime)
	if err != nil {
		return fmt.Errorf("设置任务失败: %v", err)
	}

	// 执行任务
	err = c.StartSmartControllerTask(taskTime, 0)
	if err != nil {
		return fmt.Errorf("执行任务失败: %v", err)
	}

	// 等待一秒，确保命令执行
	time.Sleep(1 * time.Second)

	logger.Info("命令执行成功")
	return nil
}

// SyncRouterTime 同步路由器系统时间
func (c *AX5400ProClient) SyncRouterTime() error {
	// 获取当前时间的时间戳格式
	now := time.Now()
	timeStr := now.Format("2006.01.02-15:04:05")

	// 构建date命令
	dateCmd := fmt.Sprintf("date -s '%s'", timeStr)

	logger.Info("正在同步路由器系统时间: %s", timeStr)

	// 执行date命令
	err := c.ExecuteCustomCommand(dateCmd)
	if err != nil {
		return fmt.Errorf("同步系统时间失败: %v", err)
	}

	logger.Info("路由器系统时间同步成功")
	return nil
}

// EnableSSH 启用SSH和Telnet
func (c *AX5400ProClient) EnableSSH() error {
	// 1. 设置系统时间
	if err := c.SetSystemTime(); err != nil {
		return err
	}

	// 2. 定义启用SSH和Telnet的步骤
	steps := []struct {
		name    string
		command string
	}{
		{"解锁Dropbear配置", "sed -i s/release/debug/g /etc/init.d/dropbear"},
		{"启用SSH配置", "nvram set ssh_en=1"},
		{"启用Telnet配置", "nvram set telnet_en=1"},
		{"提交NVRAM更改", "nvram commit"},
		//{"启用Dropbear服务", "/etc/init.d/dropbear enable"},
		{"重启Dropbear服务", "/etc/init.d/dropbear restart"},
	}

	// 执行SSH和Telnet启用步骤
	for i, step := range steps {
		logger.Info("[%d/%d] %s...", i+1, len(steps), step.name)

		// 执行命令
		if err := c.ExecuteCustomCommand(step.command); err != nil {
			return fmt.Errorf("%s失败: %v", step.name, err)
		}

		logger.Info("%s完成", step.name)

		// 每个步骤之间等待2秒，确保命令执行完成
		time.Sleep(2 * time.Second)
	}

	// 3. 验证SSH和Telnet状态
	logger.Info("验证SSH和Telnet状态...")
	status, details, err := c.CheckShellStatus()
	if err != nil {
		logger.Warn("验证SSH和Telnet状态时出错: %v", err)
	} else {
		// 显示详细的状态信息
		if status {
			logger.Info("SSH和Telnet已成功启用!")
			fmt.Println("\n" + details)

			// 4. 如果SSH已成功启用，同步路由器系统时间
			logger.Info("SSH已启用，正在同步路由器系统时间...")
			if syncErr := c.SyncRouterTime(); syncErr != nil {
				logger.Warn("同步路由器系统时间失败: %v", syncErr)
			}
		} else {
			logger.Warn("SSH和Telnet可能未成功启用，请查看详细状态")
			fmt.Println("\n" + details)
		}
	}

	return nil
}

// DisableSSH 关闭SSH和Telnet
func (c *AX5400ProClient) DisableSSH() error {
	logger.Info("开始关闭SSH和Telnet服务...")

	// 定义关闭SSH和Telnet的步骤
	steps := []struct {
		name    string
		command string
	}{
		{"禁用SSH配置", "nvram set ssh_en=0"},
		{"禁用Telnet配置", "nvram set telnet_en=0"},
		{"提交NVRAM更改", "nvram commit"},
		//{"禁用Dropbear服务", "/etc/init.d/dropbear disable"},
		{"上锁Dropbear配置", "sed -i s/debug/release/g /etc/init.d/dropbear"},
		{"停止Dropbear服务", "/etc/init.d/dropbear restart"},
		// 可以添加额外的清理步骤，如果需要的话
	}

	// 执行SSH和Telnet关闭步骤
	for i, step := range steps {
		logger.Info("[%d/%d] %s...", i+1, len(steps), step.name)

		// 执行命令
		if err := c.ExecuteCustomCommand(step.command); err != nil {
			return fmt.Errorf("%s失败: %v", step.name, err)
		}

		logger.Info("%s完成", step.name)

		// 每个步骤之间等待2秒，确保命令执行完成
		time.Sleep(2 * time.Second)
	}

	// 验证SSH和Telnet状态
	logger.Info("验证SSH和Telnet是否已关闭...")
	status, details, err := c.CheckShellStatus()
	if err != nil {
		logger.Warn("验证SSH和Telnet状态时出错: %v", err)
	} else {
		if status {
			logger.Warn("SSH和Telnet可能未成功关闭，请查看详细状态")
			fmt.Println("\n" + details)
		} else {
			logger.Info("SSH和Telnet已成功关闭!")
			fmt.Println("\n" + details)
		}
	}

	return nil
}

// VerifySSHStatus 验证SSH和Telnet状态
func (c *AX5400ProClient) VerifySSHStatus() (bool, error) {
	// 创建一个状态结构体来跟踪不同的检查结果
	status := &ShellStatusResult{
		SSHEnabled:     false,
		TelnetEnabled:  false,
		SSHPortOpen:    false,
		TelnetPortOpen: false,
	}

	// 1. 首先通过API检查SSH和Telnet的启用状态
	body, err := c.Get("api/xqsystem/fac_info")
	if err != nil {
		return false, fmt.Errorf("检查SSH状态失败: %v", err)
	}

	// 输出调试信息
	logger.Debug("路由器状态信息: %s", string(body))

	// 检查API返回的状态
	c.checkAPIStatus(body, status)

	// 2. 然后检查SSH端口(22)和Telnet端口(23)是否开放
	logger.Info("检查SSH端口(22)是否开放...")
	status.SSHPortOpen = c.CheckPortOpen(22)

	logger.Info("检查Telnet端口(23)是否开放...")
	status.TelnetPortOpen = c.CheckPortOpen(23)

	// 3. 输出详细的状态信息
	logger.Info("SSH状态检查结果:")
	logger.Info("  API返回SSH已启用: %v", status.SSHEnabled)
	logger.Info("  SSH端口(22)开放: %v", status.SSHPortOpen)
	logger.Info("Telnet状态检查结果:")
	logger.Info("  API返回Telnet已启用: %v", status.TelnetEnabled)
	logger.Info("  Telnet端口(23)开放: %v", status.TelnetPortOpen)

	// 4. 综合判断SSH是否成功启用
	// 如果API返回SSH已启用，且端口22开放，则认为SSH成功启用
	sshSuccess := status.SSHEnabled && status.SSHPortOpen

	// 如果API返回Telnet已启用，且端口23开放，则认为Telnet成功启用
	telnetSuccess := status.TelnetEnabled && status.TelnetPortOpen

	if sshSuccess {
		logger.Info("SSH服务已成功启用并且端口已开放!")
	} else if status.SSHEnabled {
		logger.Warn("SSH服务已在配置中启用，但端口22未开放，请检查防火墙设置或服务状态")
	} else if status.SSHPortOpen {
		logger.Warn("SSH端口22已开放，但API未返回启用状态，可能是配置未正确保存")
	} else {
		logger.Warn("SSH服务未启用，且端口22未开放")
	}

	if telnetSuccess {
		logger.Info("Telnet服务已成功启用并且端口已开放!")
	} else if status.TelnetEnabled {
		logger.Warn("Telnet服务已在配置中启用，但端口23未开放，请检查防火墙设置或服务状态")
	} else if status.TelnetPortOpen {
		logger.Warn("Telnet端口23已开放，但API未返回启用状态，可能是配置未正确保存")
	} else {
		logger.Warn("Telnet服务未启用，且端口23未开放")
	}

	// 返回SSH的成功状态，因为这是主要功能
	return sshSuccess, nil
}

// 检查API返回的状态
func (c *AX5400ProClient) checkAPIStatus(body []byte, status *ShellStatusResult) {
	// 检查是否为标准JSON格式
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	if err != nil {
		logger.Debug("响应不是标准JSON格式: %v", err)
		// 尝试其他方法检测SSH状态
	} else {
		// 检查是否有code字段，这是标准小米API响应格式
		if code, ok := result["code"].(float64); ok {
			if code != 0 {
				logger.Debug("API返回错误码: %v, 消息: %v", code, result["msg"])
				return
			}
		}

		// 直接检查ssh字段，这是AX5400Pro的响应格式
		if ssh, ok := result["ssh"].(bool); ok {
			logger.Info("检测到SSH状态: %v", ssh)
			status.SSHEnabled = ssh
		}

		// 检查telnet字段
		if telnet, ok := result["telnet"].(bool); ok {
			logger.Info("检测到Telnet状态: %v", telnet)
			status.TelnetEnabled = telnet
		}

		// 检查是否存在ssh_en字段，这是一些其他小米路由器的响应格式
		if data, ok := result["data"].(map[string]interface{}); ok {
			if sshEn, exists := data["ssh_en"]; exists {
				if sshEnStr, ok := sshEn.(string); ok && sshEnStr == "1" {
					logger.Info("检测到SSH已启用 (ssh_en=1)")
					status.SSHEnabled = true
				} else if sshEnBool, ok := sshEn.(bool); ok && sshEnBool {
					logger.Info("检测到SSH已启用 (ssh_en=true)")
					status.SSHEnabled = true
				}
			}

			// 检查telnet_en字段
			if telnetEn, exists := data["telnet_en"]; exists {
				if telnetEnStr, ok := telnetEn.(string); ok && telnetEnStr == "1" {
					logger.Info("检测到Telnet已启用 (telnet_en=1)")
					status.TelnetEnabled = true
				} else if telnetEnBool, ok := telnetEn.(bool); ok && telnetEnBool {
					logger.Info("检测到Telnet已启用 (telnet_en=true)")
					status.TelnetEnabled = true
				}
			}
		}
	}

	// 方法2: 检查响应中是否包含SSH相关信息
	bodyStr := string(body)
	if strings.Contains(bodyStr, `"ssh":true`) ||
		strings.Contains(bodyStr, `"ssh": true`) ||
		strings.Contains(bodyStr, `"ssh_en":"1"`) ||
		strings.Contains(bodyStr, `"ssh_en": "1"`) ||
		strings.Contains(bodyStr, `"ssh_en":1`) ||
		strings.Contains(bodyStr, `"ssh_en": 1`) {
		logger.Info("检测到SSH已启用 (响应中包含SSH启用标识)")
		status.SSHEnabled = true
	}

	// 检查响应中是否包含Telnet相关信息
	if strings.Contains(bodyStr, `"telnet":true`) ||
		strings.Contains(bodyStr, `"telnet": true`) ||
		strings.Contains(bodyStr, `"telnet_en":"1"`) ||
		strings.Contains(bodyStr, `"telnet_en": "1"`) ||
		strings.Contains(bodyStr, `"telnet_en":1`) ||
		strings.Contains(bodyStr, `"telnet_en": 1`) {
		logger.Info("检测到Telnet已启用 (响应中包含Telnet启用标识)")
		status.TelnetEnabled = true
	}
}

// CheckShellStatus 检查SSH和Telnet状态 (覆写基类方法)
func (c *AX5400ProClient) CheckShellStatus() (bool, string, error) {
	logger.Info("检查 %s 路由器的SSH和Telnet状态...", c.Model)

	// 创建状态结构体
	status := &ShellStatusResult{
		SSHEnabled:     false,
		TelnetEnabled:  false,
		SSHPortOpen:    false,
		TelnetPortOpen: false,
	}

	// 1. 通过API检查SSH和Telnet的启用状态
	body, err := c.Get("api/xqsystem/fac_info")
	if err != nil {
		return false, "", fmt.Errorf("检查状态失败: %v", err)
	}

	// 检查API返回的状态
	c.checkAPIStatus(body, status)

	// 2. 检查SSH端口(22)和Telnet端口(23)是否开放
	status.SSHPortOpen = c.CheckPortOpen(22)
	status.TelnetPortOpen = c.CheckPortOpen(23)

	// 3. 生成详细状态报告
	var detailsBuilder strings.Builder
	detailsBuilder.WriteString("SSH状态:\n")
	detailsBuilder.WriteString(fmt.Sprintf("  - 配置中启用: %v\n", status.SSHEnabled))
	detailsBuilder.WriteString(fmt.Sprintf("  - 端口22开放: %v\n", status.SSHPortOpen))

	if status.SSHEnabled && status.SSHPortOpen {
		detailsBuilder.WriteString("  - 总体状态: SSH已成功启用并且可以访问\n")
	} else if status.SSHEnabled {
		detailsBuilder.WriteString("  - 总体状态: SSH在配置中已启用，但端口未开放\n")
	} else if status.SSHPortOpen {
		detailsBuilder.WriteString("  - 总体状态: SSH端口已开放，但配置中未启用\n")
	} else {
		detailsBuilder.WriteString("  - 总体状态: SSH未启用\n")
	}

	detailsBuilder.WriteString("\nTelnet状态:\n")
	detailsBuilder.WriteString(fmt.Sprintf("  - 配置中启用: %v\n", status.TelnetEnabled))
	detailsBuilder.WriteString(fmt.Sprintf("  - 端口23开放: %v\n", status.TelnetPortOpen))

	if status.TelnetEnabled && status.TelnetPortOpen {
		detailsBuilder.WriteString("  - 总体状态: Telnet已成功启用并且可以访问\n")
	} else if status.TelnetEnabled {
		detailsBuilder.WriteString("  - 总体状态: Telnet在配置中已启用，但端口未开放\n")
	} else if status.TelnetPortOpen {
		detailsBuilder.WriteString("  - 总体状态: Telnet端口已开放，但配置中未启用\n")
	} else {
		detailsBuilder.WriteString("  - 总体状态: Telnet未启用\n")
	}

	detailsBuilder.WriteString("\n连接信息:\n")
	if status.SSHEnabled && status.SSHPortOpen {
		// 使用特定于AX5400Pro的SSH连接命令
		detailsBuilder.WriteString(fmt.Sprintf("  - SSH连接命令: %s\n", c.GetSSHCommand()))
	}
	if status.TelnetEnabled && status.TelnetPortOpen {
		detailsBuilder.WriteString(fmt.Sprintf("  - Telnet连接命令: %s\n", c.GetTelnetCommand()))
	}

	// 4. 综合判断总体状态
	// 如果SSH或Telnet任一服务配置已启用且端口已开放，则认为总体状态为成功
	overallStatus := (status.SSHEnabled && status.SSHPortOpen) || (status.TelnetEnabled && status.TelnetPortOpen)

	return overallStatus, detailsBuilder.String(), nil
}
