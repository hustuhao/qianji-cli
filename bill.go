package qianji

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	TypeExpense  = 0  // 支出
	TypeIncome   = 1  // 收入
	TypeTransfer = 2  // 转账
	TypeRefund   = 20 // 退款

	PlatformManual    = 0   // 手动
	PlatformRepeating = 120 // 重复任务
	PlatformAuto      = 122 // 自动记账

	StatusDeleted = 0 // 已删除
	StatusSynced  = 1 // 已同步
	StatusNotSync = 2 // 待同步
)

// Bill 表示一笔记账。
type Bill struct {
	ID         int64    `json:"id,omitempty"`
	AssetID    int64    `json:"assetid"`
	BookID     int64    `json:"bookid"`
	CateID     int64    `json:"cateid"`
	CreateTime int64    `json:"createtime"`
	DescInfo   string   `json:"descinfo,omitempty"`
	FromID     int64    `json:"fromid"`
	Images     []string `json:"images,omitempty"`
	Money      float64  `json:"money"`
	Platform   int      `json:"platform"`
	Remark     string   `json:"remark"`
	Status     int      `json:"status"`
	TargetID   int64    `json:"targetid"`
	TimeInSec  int64    `json:"time"`
	Type       int      `json:"type"`
	UpdateTime int64    `json:"updatetime"`
	UserID     string   `json:"userid"`
	Username   string   `json:"username,omitempty"`
	CateName   string   `json:"catename,omitempty"`
	AssetName  string   `json:"assetname,omitempty"`
	BookName   string   `json:"bookname,omitempty"`
}

// Time 返回账单时间的 time.Time。
func (b Bill) Time() time.Time {
	return time.Unix(b.TimeInSec, 0)
}

// IsExpense 是否为支出。
func (b Bill) IsExpense() bool { return b.Type == TypeExpense }

// IsIncome 是否为收入。
func (b Bill) IsIncome() bool { return b.Type == TypeIncome }

// SyncPayload 是 syncall 请求体。
type SyncPayload struct {
	Bills SyncBody `json:"bills"`
}

// SyncBody 包含待同步和待删除的账单。
type SyncBody struct {
	ChangeList []Bill  `json:"changelist,omitempty"`
	DelList    []int64 `json:"dellist,omitempty"`
}

// NewBill 创建一笔基础支出账单。
func NewBill(bookID int64, money float64, remark string) Bill {
	now := time.Now().Unix()
	return Bill{
		BookID:     bookID,
		Money:      money,
		Remark:     remark,
		Type:       TypeExpense,
		Platform:   PlatformManual,
		Status:     StatusNotSync,
		TimeInSec:  now,
		CreateTime: now,
		UpdateTime: now,
		Images:     []string{},
	}
}

// NewIncome 创建一笔收入账单。
func NewIncome(bookID int64, money float64, remark string) Bill {
	b := NewBill(bookID, money, remark)
	b.Type = TypeIncome
	return b
}

// WithCategory 设置分类。
func (b Bill) WithCategory(cateID int64) Bill {
	b.CateID = cateID
	return b
}

// WithAsset 设置资产账户。
func (b Bill) WithAsset(assetID int64) Bill {
	b.AssetID = assetID
	return b
}

// WithTime 设置账单时间。
func (b Bill) WithTime(t time.Time) Bill {
	b.TimeInSec = t.Unix()
	b.CreateTime = t.Unix()
	b.UpdateTime = t.Unix()
	return b
}

// SyncBills 同步账单到服务端，同时拉回服务端数据。
func (s *Session) SyncBills(changes []Bill, deletes []int64) ([]Bill, error) {
	payload := SyncPayload{
		Bills: SyncBody{
			ChangeList: changes,
			DelList:    deletes,
		},
	}

	vJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	params := url.Values{}
	params.Set("uid", s.UserID)
	params.Set("v", string(vJSON))

	data, err := s.Client.doPost("bill", "syncall", params, s.Token)
	if err != nil {
		return nil, fmt.Errorf("sync bills: %w", err)
	}

	return parseSyncResponse(data)
}

// parseSyncResponse 解析 syncall 响应，格式:
//
//	{"ec":200,"em":"","data":{"sync_result":{"bill":{"success":[...]},"sync_time":...}}}
func parseSyncResponse(data []byte) ([]Bill, error) {
	var raw struct {
		Ec   int `json:"ec"`
		Data struct {
			SyncResult struct {
				Bill struct {
					Success []Bill `json:"success"`
				} `json:"bill"`
			} `json:"sync_result"`
		} `json:"data"`
		Em string `json:"em"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse sync response: %w", err)
	}
	if raw.Ec != 200 {
		msg := raw.Em
		if strings.HasPrefix(msg, "{") {
			var emMsg struct{ Msg string `json:"msg"` }
			if json.Unmarshal([]byte(msg), &emMsg) == nil && emMsg.Msg != "" {
				msg = emMsg.Msg
			}
		}
		return nil, fmt.Errorf("sync failed (ec=%d): %s", raw.Ec, msg)
	}
	return raw.Data.SyncResult.Bill.Success, nil
}

// ListBills 拉取全部账单（发送空 sync，获取服务端全量）。
func (s *Session) ListBills() ([]Bill, error) {
	return s.SyncBills(nil, nil)
}

// BillsForDate 筛选指定日期的账单（基于 time 字段）。
func BillsForDate(bills []Bill, t time.Time) []Bill {
	y, m, d := t.Date()
	var result []Bill
	for _, b := range bills {
		by, bm, bd := b.Time().Date()
		if by == y && bm == m && bd == d {
			result = append(result, b)
		}
	}
	return result
}

// TotalMoney 计算账单列表的总金额。
func TotalMoney(bills []Bill) float64 {
	var total float64
	for _, b := range bills {
		total += b.Money
	}
	return total
}

// AddBill 添加一笔账单。
func (s *Session) AddBill(bill Bill) (*AddBillResult, error) {
	bills, err := s.SyncBills([]Bill{bill}, nil)
	if err != nil {
		return nil, err
	}
	return &AddBillResult{Bills: bills}, nil
}

// PullResult 是 syncv2/pull 响应。
type PullResult struct {
	Changes  []Bill   `json:"changes"`
	Deletes  []int64  `json:"deletes"`
	PageInfo PullPage `json:"-"`
}

// PullPage 分页信息。
type PullPage struct {
	BookID    int64  `json:"bookid"`
	PageOff   int64  `json:"pageoffset"`
	HasMore   bool   `json:"hasmore"`
	Count     int    `json:"count"`
	PageSign  string `json:"pagesign"`
}

// PullBills 从服务端拉取其他设备的账单（syncv2/pull）。
// bookID=-1 表示拉全部账本。lastTimes 可追溯上次同步时间实现增量拉取。
func (s *Session) PullBills(bookID int64, lastTimes string, pageOff int64, pageSign string) (*PullResult, error) {
	params := url.Values{}
	params.Set("uid", s.UserID)
	params.Set("bookid", fmt.Sprintf("%d", bookID))
	if lastTimes != "" {
		params.Set("lasttimes", lastTimes)
	}
	params.Set("pageoffset", fmt.Sprintf("%d", pageOff))
	params.Set("pagesign", pageSign)

	data, err := s.Client.doPost("syncv2", "pull", params, s.Token)
	if err != nil {
		return nil, fmt.Errorf("pull: %w", err)
	}

	var raw struct {
		Ec   int      `json:"ec"`
		Em   string   `json:"em"`
		Data PullResult `json:"data"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse pull: %w", err)
	}
	if raw.Ec != 200 {
		return nil, fmt.Errorf("pull failed (ec=%d): %s", raw.Ec, raw.Em)
	}

	result := &raw.Data

	// 解析内嵌的 pageParams
	var page struct {
		BookID   int64  `json:"bookid"`
		PageOff  int64  `json:"pageoffset"`
		HasMore  int    `json:"hasmore"`
		Count    int    `json:"count"`
		PageSign string `json:"pagesign"`
	}
	var rawWrapper struct {
		Data struct {
			BookID   int64  `json:"bookid"`
			PageOff  int64  `json:"pageoffset"`
			HasMore  int    `json:"hasmore"`
			Count    int    `json:"count"`
			PageSign string `json:"pagesign"`
		} `json:"data"`
	}
	// pageParams 在响应的 data 级别（与 changes/deletes 平级）
	if err := json.Unmarshal(data, &rawWrapper); err == nil {
		page.BookID = rawWrapper.Data.BookID
		page.PageOff = rawWrapper.Data.PageOff
		page.HasMore = rawWrapper.Data.HasMore
		page.Count = rawWrapper.Data.Count
		page.PageSign = rawWrapper.Data.PageSign
	}
	result.PageInfo = PullPage{
		BookID:   page.BookID,
		PageOff:  page.PageOff,
		HasMore:  page.HasMore == 1,
		Count:    page.Count,
		PageSign: page.PageSign,
	}
	return result, nil
}

// FullSync 完整同步：先拉取其他设备账单，再推送本地待同步。
func (s *Session) FullSync(pending []Bill) (pulledBills []Bill, err error) {
	// 1. PULL: 循环拉直到 hasMore=false
	pageOff := int64(0)
	pageSign := ""
	for {
		pr, err := s.PullBills(-1, "", pageOff, pageSign)
		if err != nil {
			return pulledBills, fmt.Errorf("pull page: %w", err)
		}
		pulledBills = append(pulledBills, pr.Changes...)
		// 处理删除
		for _, delID := range pr.Deletes {
			// 本地标记为已删除
			SaveBills([]Bill{{ID: delID, Status: 0}})
		}
		if !pr.PageInfo.HasMore {
			break
		}
		pageOff = pr.PageInfo.PageOff
		pageSign = pr.PageInfo.PageSign
	}

	// 2. PUSH: 推送本地待同步
	if len(pending) > 0 {
		s.SyncBills(pending, nil)
	}
	return pulledBills, nil
}
type AddBillResult struct {
	Bills []Bill
}

// DeleteBill 删除一笔账单。
func (s *Session) DeleteBill(billID int64) error {
	_, err := s.SyncBills(nil, []int64{billID})
	return err
}
