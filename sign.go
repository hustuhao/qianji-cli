package qianji

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 钱迹 API 签名常量，来自 libfabricsuffer.so 逆向
const (
	saltRequestID    = "free20170908&x_*1127"
	saltEncReqID     = "michaeljackson"
	saltTok          = "1172020"
	versionCode      = 1170
	packageName      = "com.mutangtech.qianji"
	versionName      = "4.3.7"
	deviceBrand      = "XIAOMI"
	deviceModel      = "CLI"
	osVersion        = "32" // Android 12
	timezoneOffset   = "8"  // UTC+8
	language         = "zh"
	region           = "CN"
	market           = "cli"
)

var (
	deviceID string
)

func init() {
	// 持久化设备 ID，确保服务端识别为同一设备
	home, _ := os.UserHomeDir()
	devIDFile := filepath.Join(home, ".qianji", "device_id")
	data, err := os.ReadFile(devIDFile)
	if err == nil && len(data) > 0 {
		deviceID = strings.TrimSpace(string(data))
		return
	}

	// 首次生成
	uuid := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		rand.Uint32(),
		rand.Uint32()&0xffff,
		(rand.Uint32()&0x0fff)|0x4000,
		(rand.Uint32()&0x3fff)|0x8000,
		rand.Uint64()&0xffffffffffff,
	)
	hash := md5.Sum([]byte(uuid))
	deviceID = fmt.Sprintf("%x", hash)
	os.MkdirAll(filepath.Dir(devIDFile), 0700)
	os.WriteFile(devIDFile, []byte(deviceID), 0600)
}

// computeSign 计算 reqidv2 和 tok，完全模仿 libfabricsuffer.so 逻辑
func computeSign(ctrl, act string) (reqID, tok string) {
	millis := time.Now().UnixMilli()

	// Step 1: combined = versionCode + millis + packageName + ctrl
	combined := fmt.Sprintf("%d%d%s%s", versionCode, millis, packageName, ctrl)

	// Step 2: requestId = MD5(combined + saltRequestID)
	hash := md5.Sum([]byte(combined + saltRequestID))
	reqID = fmt.Sprintf("%x", hash)

	// Step 3: encReqid = MD5(reqID + saltEncReqID)
	hash2 := md5.Sum([]byte(reqID + saltEncReqID))
	encReqID := fmt.Sprintf("%x", hash2)

	// Step 4: tok = MD5(reqID + saltTok + ctrl + encReqID + act)
	tokInput := reqID + saltTok + ctrl + encReqID + act
	hash3 := md5.Sum([]byte(tokInput))
	tok = fmt.Sprintf("%x", hash3)

	return reqID, tok
}

// signHeaders 返回所有设备相关的 HTTP header
func signHeaders() map[string]string {
	return map[string]string{
		"os":             "1",
		"osvs":           osVersion,
		"devbrand":       deviceBrand,
		"devname":        deviceModel,
		"devid":          deviceID,
		"vs":             fmt.Sprintf("%d", versionCode),
		"pkg":            packageName,
		"vsn":            versionName,
		"timezoneoffset": timezoneOffset,
		"clang":          language,
		"cregion":        region,
		"mk":             market,
	}
}
