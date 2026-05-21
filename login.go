package qianji

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Login 使用账密登录。account 可以是手机号或邮箱，password 为明文（自动 MD5）。
func (c *Client) Login(account, password string) (*Session, error) {
	// 密码 MD5 用小写十六进制输出（对应 n8.k.b 行为）
	hash := md5.Sum([]byte(password))
	pwdMD5 := fmt.Sprintf("%x", hash)

	params := url.Values{}
	params.Set("v", account)
	params.Set("pwd", pwdMD5)

	data, err := c.doPost("account", "login", params, "")
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	var raw struct {
		Ec   int `json:"ec"`
		Data struct {
			Token string `json:"token"`
			User  struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
		Em string `json:"em"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	if raw.Ec != 200 {
		// 解析 em 中的 JSON 消息（格式可能是嵌套 JSON 字符串）
		var emMsg struct{ Msg string `json:"msg"` }
		msg := raw.Em
		if json.Unmarshal([]byte(raw.Em), &emMsg) == nil && emMsg.Msg != "" {
			msg = emMsg.Msg
		}
		return nil, fmt.Errorf("login failed (ec=%d): %s", raw.Ec, msg)
	}

	return &Session{
		Client: c,
		Token:  raw.Data.Token,
		UserID: raw.Data.User.ID,
	}, nil
}

// Logout 登出。
func (s *Session) Logout() error {
	params := url.Values{}
	params.Set("uid", s.UserID)

	_, err := s.Client.doPost("account", "logout", params, s.Token)
	return err
}

// GetCategories 获取所有分类。
func (s *Session) GetCategories() ([]Category, error) {
	params := url.Values{}
	params.Set("uid", s.UserID)

	data, err := s.Client.doPost("category", "listv2", params, s.Token)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Ec   int `json:"ec"`
		Data struct {
			List []Category `json:"list"`
		} `json:"data"`
		Em string `json:"em"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if raw.Ec != 200 {
		return nil, fmt.Errorf("get categories failed (ec=%d): %s", raw.Ec, raw.Em)
	}
	return raw.Data.List, nil
}

// Category 分类模型。
type Category struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	ParentID int64  `json:"parent_id"`
	BookID   int64  `json:"bookid"`
	Level    int    `json:"level"`
	Type     int    `json:"type"`           // 0=expense 1=income
	Sort     int    `json:"sort,omitempty"` // 支出 / 收入
}

// GetBooks 获取账本列表。
func (s *Session) GetBooks() ([]Book, error) {
	params := url.Values{}
	params.Set("uid", s.UserID)

	data, err := s.Client.doPost("book", "list", params, s.Token)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Ec   int `json:"ec"`
		Data struct {
			List []Book `json:"list"`
		} `json:"data"`
		Em string `json:"em"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if raw.Ec != 200 {
		return nil, fmt.Errorf("get books failed (ec=%d): %s", raw.Ec, raw.Em)
	}
	return raw.Data.List, nil
}

// Book 账本模型。
type Book struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	UserID string `json:"userid"`
}

// GetAssets 获取资产账户列表。
func (s *Session) GetAssets() ([]AssetAccount, error) {
	params := url.Values{}
	params.Set("uid", s.UserID)

	data, err := s.Client.doPost("asset", "list", params, s.Token)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Ec   int `json:"ec"`
		Data struct {
			List []AssetAccount `json:"list"`
		} `json:"data"`
		Em string `json:"em"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if raw.Ec != 200 {
		return nil, fmt.Errorf("get assets failed (ec=%d): %s", raw.Ec, raw.Em)
	}
	return raw.Data.List, nil
}

// AssetAccount 资产账户模型。
type AssetAccount struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Type   int    `json:"type"`
	Amount float64 `json:"amount"`
}

// indentJSON 格式化 JSON 输出。
func indentJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return strings.TrimSpace(string(b))
}
