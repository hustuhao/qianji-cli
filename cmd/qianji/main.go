// 钱迹 CLI — 通过云 API 记账、查账、管理
//
// 用法示例:
//
//	qianji login         交互式登录
//	qianji add 26.5 咖啡 快速记账
//	qianji cats          列出分类
//	qianji books         列出账本
//	qianji assets        列出资产账户
//
// 首次使用需登录，token 缓存在 ~/.qianji_token.json。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wepie/qianji"
)

var tokenFile = filepath.Join(os.Getenv("HOME"), ".qianji_token.json")

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "login":
		cmdLogin(args)
	case "add":
		cmdAdd(args)
	case "list", "ls":
		cmdList(args)
	case "cats", "cat", "categories":
		cmdCats(args)
	case "books":
		cmdBooks(args)
	case "assets":
		cmdAssets(args)
	case "logout":
		cmdLogout()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`钱迹 CLI (qianji)

命令:
  login          登录到钱迹账号
  add <金额> <备注>  快速添加一笔支出
  list            列出今日账单
  cats           列出所有分类
  books          列出所有账本
  assets         列出资产账户
  logout         登出

用法:
  qianji login                 交互式登录
  qianji add 26.5 咖啡         支出一笔 26.5 元"咖啡"
  qianji add -i 1000 工资      收入一笔 1000 元"工资"
  qianji add -c 5 35 午餐      指定分类 ID 为 5
  qianji list                  列出今天的账单
  qianji list -d 5.20          列出 5 月 20 日的账单`)
}

// ---- login ----

func cmdLogin(args []string) {
	client := qianji.NewClient("")

	var account, password string
	if len(args) >= 2 {
		account = args[0]
		password = args[1]
	} else {
		fmt.Print("手机号/邮箱: ")
		fmt.Scanln(&account)
		fmt.Print("密码: ")
		fmt.Scanln(&password)
	}

	session, err := client.Login(account, password)
	if err != nil {
		fatalf("登录失败: %v", err)
	}

	saveToken(session.Token, session.UserID)
	fmt.Printf("登录成功! UID=%s\n", session.UserID)
}

func saveToken(token, userID string) {
	data, _ := json.Marshal(map[string]string{
		"token":   token,
		"user_id": userID,
	})
	if err := os.WriteFile(tokenFile, data, 0600); err != nil {
		fatalf("保存 token 失败: %v", err)
	}
}

func loadToken() (token, userID string, ok bool) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", "", false
	}
	var m map[string]string
	if json.Unmarshal(data, &m) != nil {
		return "", "", false
	}
	return m["token"], m["user_id"], true
}

func mustSession() *qianji.Session {
	token, userID, ok := loadToken()
	if !ok {
		fatalf("未登录，请先执行: qianji login")
	}
	return &qianji.Session{
		Client: qianji.NewClient(""),
		Token:  token,
		UserID: userID,
	}
}

// ---- add ----

func cmdAdd(args []string) {
	if len(args) < 1 {
		fatalf("用法: qianji add <金额> <备注>  (可选: -i 收入, -c <分类ID>, -b <账本ID>, -a <资产ID>)")
	}

	var isIncome bool
	var cateID, bookID, assetID int64

	// 解析简单 flags
	var positional []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-i":
			isIncome = true
		case "-c":
			i++
			if i < len(args) {
				cateID, _ = strconv.ParseInt(args[i], 10, 64)
			}
		case "-b":
			i++
			if i < len(args) {
				bookID, _ = strconv.ParseInt(args[i], 10, 64)
			}
		case "-a":
			i++
			if i < len(args) {
				assetID, _ = strconv.ParseInt(args[i], 10, 64)
			}
		default:
			positional = append(positional, args[i])
		}
	}

	if len(positional) < 1 {
		fatalf("请提供金额")
	}

	moneyVal, err := strconv.ParseFloat(positional[0], 64)
	if err != nil {
		fatalf("金额格式错误: %s", positional[0])
	}

	remark := ""
	if len(positional) >= 2 {
		remark = positional[1]
	}

	s := mustSession()

	var bill qianji.Bill
	if isIncome {
		bill = qianji.NewIncome(bookID, moneyVal, remark)
	} else {
		bill = qianji.NewBill(bookID, moneyVal, remark)
	}

	if cateID > 0 {
		bill = bill.WithCategory(cateID)
	}
	if assetID > 0 {
		bill = bill.WithAsset(assetID)
	}

	_, err = s.AddBill(bill)
	if err != nil {
		fatalf("记账失败: %v", err)
	}

	act := "支出"
	if isIncome {
		act = "收入"
	}
	fmt.Printf("OK %s %.2f %s\n", act, moneyVal, remark)
}

// ---- cats ----

func cmdCats([]string) {
	s := mustSession()
	cats, err := s.GetCategories()
	if err != nil {
		fatalf("获取分类失败: %v", err)
	}
	for _, c := range cats {
		tag := "支出"
		if c.Type == 1 {
			tag = "收入"
		}
		indent := ""
		if c.Level > 0 {
			indent = "  "
		}
		fmt.Printf("%s[%d] %s  (%s)\n", indent, c.ID, c.Name, tag)
	}
}

// ---- books ----

func cmdBooks([]string) {
	s := mustSession()
	books, err := s.GetBooks()
	if err != nil {
		fatalf("获取账本失败: %v", err)
	}
	fmt.Printf("共 %d 个账本:\n", len(books))
	for _, b := range books {
		fmt.Printf("[%d] %s\n", b.ID, b.Name)
	}
}

// ---- assets ----

func cmdAssets([]string) {
	s := mustSession()
	assets, err := s.GetAssets()
	if err != nil {
		fatalf("获取资产失败: %v", err)
	}
	for _, a := range assets {
		fmt.Printf("[%d] %-15s  %.2f\n", a.ID, a.Name, a.Amount)
	}
}

// ---- list ----

func cmdList(args []string) {
	s := mustSession()

	// 解析日期参数 -d MM.DD
	targetDate := time.Now()
	for i := 0; i < len(args); i++ {
		if args[i] == "-d" && i+1 < len(args) {
			d, err := time.Parse("1.2", args[i+1])
			if err == nil {
				now := time.Now()
				targetDate = time.Date(now.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.Local)
			}
			break
		}
	}

	fmt.Printf("正在拉取账单...\n")
	bills, err := s.ListBills()
	if err != nil {
		fatalf("获取账单失败: %v", err)
	}

	todayBills := qianji.BillsForDate(bills, targetDate)
	if len(todayBills) == 0 {
		fmt.Printf("%s 没有账单记录\n", targetDate.Format("01月02日"))
		return
	}

	var expenseTotal, incomeTotal float64
	fmt.Printf("\n%s:\n", targetDate.Format("2006-01-02  Monday"))
	fmt.Println(strings.Repeat("-", 60))

	for _, b := range todayBills {
		tag := "支出"
		if b.IsIncome() {
			tag = "收入"
			incomeTotal += b.Money
		} else {
			expenseTotal += b.Money
		}
		t := time.Unix(b.TimeInSec, 0).Format("15:04")
		fmt.Printf("  %s  %-6s  ¥%-8.2f  %s\n", t, tag, b.Money, b.Remark)
	}

	fmt.Println(strings.Repeat("-", 60))
	if expenseTotal > 0 {
		fmt.Printf("  支出合计: ¥%.2f\n", expenseTotal)
	}
	if incomeTotal > 0 {
		fmt.Printf("  收入合计: ¥%.2f\n", incomeTotal)
	}
	fmt.Printf("  共 %d 笔\n", len(todayBills))
}

// ---- logout ----

func cmdLogout() {
	os.Remove(tokenFile)
	fmt.Println("已登出")
}

// ---- util ----

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
