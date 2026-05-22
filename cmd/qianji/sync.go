package main

import (
	"fmt"

	"github.com/wepie/qianji"
)

func cmdSync(args []string) {
	s := mustSession()

	fmt.Println("从服务端拉取...")
	// 1. 拉取其他设备账单
	pending, _ := qianji.QueryPendingBills()
	pulled, err := s.FullSync(pending)
	if err != nil {
		fmt.Printf("拉取错误: %v\n", err)
	}
	fmt.Printf("拉取 %d 条账单\n", len(pulled))

	// 2. 存入本地
	if len(pulled) > 0 {
		for i := range pulled {
			pulled[i].Status = 1 // 服务端来的直接标已同步
		}
		qianji.SaveBills(pulled)
	}

	// 3. 标记本地待同步为已同步（刚刚推过了）
	ids := make([]int64, len(pending))
	for i, b := range pending {
		ids[i] = b.ID
	}
	qianji.MarkSynced(ids)

	fmt.Printf("同步完成。本地共 %d 条账单。\n", qianji.CountBills())
}
