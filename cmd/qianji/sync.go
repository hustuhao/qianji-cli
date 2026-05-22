package main

import (
	"fmt"

	"github.com/wepie/qianji"
)

func cmdSync(args []string) {
	s := mustSession()

	fmt.Println("从服务端拉取...")
	pending, _ := qianji.QueryPendingBills()
	pulled, err := s.FullSync(pending)
	if err != nil {
		fmt.Printf("拉取错误: %v\n", err)
	}
	fmt.Printf("拉取 %d 条账单\n", len(pulled))

	// 拉回的存入本地
	if len(pulled) > 0 {
		for i := range pulled {
			pulled[i].Status = 1
		}
		qianji.SaveBills(pulled)
	}

	fmt.Printf("同步完成。本地共 %d 条账单。\n", qianji.CountBills())
}
