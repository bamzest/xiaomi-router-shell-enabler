package utils

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/bamzest/xiaomi-router-shell-enabler/pkg/logger"
)

// 路由器固件解包后，从 /bin/mkxqimage 中提取的盐
var salt = map[string]string{
	"r1d":    "A2E371B0-B34B-48A5-8C40-A7133F3B5D88",
	"others": "d44fb0960aa0-a5e6-4a30-250f-6d2df50a",
}

// CalculateSSHPassword 计算小米路由器SSH密码
// 密码算法：原始 SN 拼接反转后的盐，做 md5 运算取前 8 个字符
func CalculateSSHPassword(sn string) string {
	if sn == "" {
		logger.Warn("序列号为空，无法计算SSH密码")
		return ""
	}
	
	logger.Debug("计算SSH密码，序列号: %s", sn)
	saltValue := getSalt(sn)
	logger.Debug("使用盐值: %s", saltValue)
	
	// 计算MD5
	md5sum := md5.Sum([]byte(sn + saltValue))
	password := fmt.Sprintf("%x", md5sum)[:8]
	
	logger.Debug("计算得到的密码: %s", password)
	return password
}

// getSalt 根据序列号选择合适的盐值
// SN 中不含 '/' 则为 r1d
func getSalt(sn string) string {
	if !strings.Contains(sn, "/") {
		logger.Debug("检测到R1D路由器")
		return salt["r1d"]
	} else {
		logger.Debug("检测到非R1D路由器")
		return swapSalt(salt["others"])
	}
}

// swapSalt 非R1D盐要反转后才能使用
func swapSalt(salt string) string {
	parts := strings.Split(salt, "-")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, "-")
}