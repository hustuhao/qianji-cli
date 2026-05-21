// curl 后端 — 绕过 Go TLS 兼容性问题
package qianji

import (
	"fmt"
	"net/url"
	"os/exec"
)

// doPostCurl 通过 curl 发送 POST 请求（参数放 body）
func (c *Client) doPostCurl(ctrl, act string, params url.Values, token string) ([]byte, error) {
	return c.doRequest(ctrl, act, params, token, true)
}

// doGetCurl 通过 curl 发送 GET 请求（参数放 query string）
func (c *Client) doGetCurl(ctrl, act string, params url.Values, token string) ([]byte, error) {
	return c.doRequest(ctrl, act, params, token, false)
}

func (c *Client) doRequest(ctrl, act string, params url.Values, token string, isPost bool) ([]byte, error) {
	u := c.Host + "/" + ctrl + "/" + act

	args := []string{"-s", "--connect-timeout", "10"}

	// POST: 参数放 body (-d)，GET: 参数放 query string
	if isPost {
		args = append(args, "-X", "POST")
		if len(params) > 0 {
			args = append(args, "-d", params.Encode())
		}
	} else {
		if len(params) > 0 {
			u += "?" + params.Encode()
		}
	}

	args = append(args, u)

	// 设备头
	for k, v := range signHeaders() {
		args = append(args, "-H", fmt.Sprintf("%s: %s", k, v))
	}

	// 签名头
	reqID, tok := computeSign(ctrl, act)
	args = append(args,
		"-H", fmt.Sprintf("reqidv2: %s", reqID),
		"-H", fmt.Sprintf("tok: %s", tok),
		"-H", "content-type: application/x-www-form-urlencoded",
	)

	args = append(args,
		"-H", fmt.Sprintf("ctrl: %s", ctrl),
		"-H", fmt.Sprintf("act: %s", act),
	)

	if token != "" {
		args = append(args, "-H", fmt.Sprintf("utoken: %s", token))
	}

	cmd := exec.Command("curl", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("curl: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("curl: %w", err)
	}
	return out, nil
}

// 对接 doPost/doGet 接口（供 login.go / bill.go 调用）
func (c *Client) doPost(ctrl, act string, params url.Values, token string) ([]byte, error) {
	return c.doPostCurl(ctrl, act, params, token)
}

func (c *Client) doGet(ctrl, act string, params url.Values, token string) ([]byte, error) {
	return c.doGetCurl(ctrl, act, params, token)
}
