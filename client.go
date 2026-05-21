// Package qianji 提供了与钱迹云 API 交互的 Go 客户端
//
// API 签名系统已通过逆向 libfabricsuffer.so 完全破解。
// 使用前需通过 Login() 获取 Session。
package qianji

import (
	"encoding/json"
)

const DefaultHost = "https://api.qianjiapp.com"

// Client 封装 HTTP 连接池。
type Client struct {
	Host string
}

// Session 持有一个登录后的认证令牌和用户信息。
type Session struct {
	Client *Client
	Token  string
	UserID string
}

// NewClient 创建一个新的 API 客户端。
func NewClient(host string) *Client {
	if host == "" {
		host = DefaultHost
	}
	return &Client{Host: host}
}

// JSONResponse 是标准 API 响应的通用结构。
type JSONResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}
