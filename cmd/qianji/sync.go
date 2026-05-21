package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/wepie/qianji"
)

func cmdSync(args []string) {
	mustSession() // 确保已登录

	device := pickDevice(args)
	if device == "" {
		fmt.Println("未找到已连接的 Android 设备。")
		fmt.Println("请用 USB 连接手机或确保 adb 已连接，然后重试。")
		fmt.Println("用法: qianji sync [device_id]")
		os.Exit(1)
	}

	fmt.Printf("设备: %s\n", device)

	// 1. adb root
	adb(device, "root")

	// 2. adb pull 手机数据库
	tmpFile := os.TempDir() + "/qianji_sync.db"
	fmt.Println("拉取手机数据库...")
	out, err := exec.Command("adb", "-s", device, "pull",
		"/data/data/com.mutangtech.qianji/databases/qianjiapp", tmpFile).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "adb pull 失败: %s\n", string(out))
		os.Exit(1)
	}
	defer os.Remove(tmpFile)

	// 3. 推送本地待同步账单到服务端
	s := mustSession()
	localPending, _ := qianji.QueryPendingBills()
	if len(localPending) > 0 {
		fmt.Printf("推送 %d 条本地待同步...\n", len(localPending))
		serverBills, err := s.SyncBills(localPending, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "推送失败: %v\n", err)
		} else {
			qianji.SaveBills(serverBills)
			// 标记已同步
			ids := make([]int64, len(serverBills))
			for i, b := range serverBills {
				ids[i] = b.ID
			}
			qianji.MarkSynced(ids)
			fmt.Printf("推送完成: %d 条\n", len(serverBills))
		}
	}

	// 4. 合并手机数据（INSERT OR REPLACE by billid）
	phoneBills := readPhoneBills(tmpFile)
	fmt.Printf("手机有 %d 条账单, 合并到本地...\n", len(phoneBills))
	if len(phoneBills) > 0 {
		// 全部标记为已同步（来自手机的数据都是 status=1）
		for i := range phoneBills {
			phoneBills[i].Status = 1
		}
		qianji.SaveBills(phoneBills)
	}

	fmt.Printf("同步完成。本地共 %d 条账单。\n", qianji.CountBills())
}

func pickDevice(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	out, err := exec.Command("adb", "devices").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "\tdevice") && !strings.Contains(line, "List") {
			return strings.Fields(line)[0]
		}
	}
	return ""
}

func adb(device string, args ...string) {
	all := append([]string{"-s", device}, args...)
	exec.Command("adb", all...).Run()
}

func readPhoneBills(path string) []qianji.Bill {
	// 直接读 SQLite 文件
	tmpDB, err := qianji.OpenReadOnly(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取手机库: %v\n", err)
		return nil
	}
	defer tmpDB.Close()

	bills, err := qianji.QueryAllFrom(tmpDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询手机库: %v\n", err)
		return nil
	}
	return bills
}
