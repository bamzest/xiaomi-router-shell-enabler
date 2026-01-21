package auth

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bamzest/xiaomi-router-shell-enabler/pkg/logger"
)

const (
	// DefaultRouterIP 默认路由器IP地址
	DefaultRouterIP = "192.168.31.1"

	// Key 小米路由器加密密钥
	Key = "a2ffa5c9be07488bbb04a3a47d3c5f6a"
)

// LoginResponse 登录响应结构
type LoginResponse struct {
	Code  int    `json:"code"`
	Token string `json:"token"`
	URL   string `json:"url"`
}

// 生成 nonce
func generateNonce() string {
	rand.Seed(time.Now().UnixNano())
	timestamp := time.Now().Unix()
	random := rand.Intn(10000)
	return fmt.Sprintf("0_%s_%d_%d", "", timestamp, random)
}

// SHA1 加密
func sha1Sum(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// SHA256 加密
func sha256Sum(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// 使用 SHA1 加密密码
func encryptPasswordSHA1(password, nonce string) string {
	// 第一次 SHA1: 密码+key
	firstHash := sha1Sum(password + Key)
	// 第二次 SHA1: nonce + 第一次哈希结果
	secondHash := sha1Sum(nonce + firstHash)
	return secondHash
}

// 使用 SHA256 加密密码
func encryptPasswordSHA256(password, nonce string) string {
	// 第一次 SHA256: 密码+key
	firstHash := sha256Sum(password + Key)
	// 第二次 SHA256: nonce + 第一次哈希结果
	secondHash := sha256Sum(nonce + firstHash)
	return secondHash
}

// GetStok 获取路由器的stok
func GetStok(routerIP, password string, useSHA256 bool) (string, error) {
	// 生成 nonce
	nonce := generateNonce()

	// 加密密码
	var encryptedPassword string
	if useSHA256 {
		encryptedPassword = encryptPasswordSHA256(password, nonce)
	} else {
		encryptedPassword = encryptPasswordSHA1(password, nonce)
	}

	// 构建表单数据
	formData := url.Values{}
	formData.Set("username", "admin")
	formData.Set("password", encryptedPassword)
	formData.Set("logtype", "2")
	formData.Set("nonce", nonce)

	// 构建登录 URL
	loginURL := fmt.Sprintf("http://%s/cgi-bin/luci/api/xqsystem/login", routerIP)

	logger.Debug("发送登录请求到: %s", loginURL)
	logger.Debug("使用 nonce: %s", nonce)
	logger.Debug("加密后的密码: %s", encryptedPassword)

	// 创建请求
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	logger.Debug("原始响应: %s", string(body))

	// 解析 JSON 响应
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("解析 JSON 失败: %v", err)
	}

	// 检查登录结果
	if loginResp.Code == 0 {
		logger.Info("登录成功! 获取到的 stok: %s", loginResp.Token)
		return loginResp.Token, nil
	} else {
		return "", fmt.Errorf("登录失败, 错误代码: %d, 消息: %s", loginResp.Code, loginResp.URL)
	}
}
