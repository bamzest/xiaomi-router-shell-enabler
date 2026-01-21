package client

import (
	"fmt"
	"strings"

	"github.com/bamzest/xiaomi-router-shell-enabler/pkg/auth"
	"github.com/bamzest/xiaomi-router-shell-enabler/pkg/logger"
	"github.com/bamzest/xiaomi-router-shell-enabler/pkg/routers"
)

// RouterClient 路由器客户端接口
type RouterClient interface {
	// SetSystemTime 设置系统时间
	SetSystemTime() error

	// EnableSSH 启用SSH
	EnableSSH() error

	// DisableSSH 关闭SSH
	DisableSSH() error

	// VerifySSHStatus 验证SSH状态
	VerifySSHStatus() (bool, error)

	// ExecuteCustomCommand 执行自定义命令
	ExecuteCustomCommand(command string) error

	// CheckShellStatus 检查SSH和Telnet状态
	// 返回值：总体状态(bool), 详细状态信息(string), 错误(error)
	CheckShellStatus() (bool, string, error)

	// GetSSHCommand 获取适用于此型号的SSH连接命令
	GetSSHCommand() string

	// GetTelnetCommand 获取适用于此型号的Telnet连接命令
	GetTelnetCommand() string
}

// 创建路由器客户端的工厂函数 - 使用密码而不是token
func NewRouterClient(host, password, model string) (RouterClient, error) {
	logger.Debug("创建路由器客户端: 型号=%s, 主机=%s", model, host)

	// 将型号转为小写，便于匹配
	modelLower := strings.ToLower(model)

	// 根据型号选择合适的认证方式
	useSHA256 := true
	if modelLower == "redmi_ax5400pro" {
		// AX5400Pro 使用 SHA256
		useSHA256 = true
	}
	// 其他型号可以在这里添加特殊的认证方式

	// 通过密码获取 stok
	token, err := auth.GetStok(host, password, useSHA256)
	if err != nil {
		return nil, fmt.Errorf("获取 stok 失败: %v", err)
	}

	logger.Info("成功获取 stok: %s", token)

	// 检查路由器型号是否支持
	switch modelLower {
	case "redmi_ax5400pro":
		return routers.NewAX5400ProClient(host, token), nil
	// 可以在这里添加更多型号的支持
	// case "xiaomi_ax3600":
	//     return routers.NewAX3600Client(host, token), nil
	default:
		return nil, fmt.Errorf("不支持的路由器型号: %s", model)
	}
}

// 获取支持的路由器型号列表
func GetSupportedModels() []string {
	// 返回所有支持的型号
	return []string{
		"redmi_ax5400pro",
		// 可以在这里添加更多支持的型号
		// "xiaomi_ax3600",
	}
}
