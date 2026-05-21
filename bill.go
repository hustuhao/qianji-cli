package qianji

import (
	"encoding/json"
	"fmt"
	"net/url"
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

	StatusDeleted  = 0 // 已删除
	StatusSynced   = 1 // 已同步
	StatusNotSync  = 2 // 待同步
)

// Bill 表示一笔记账。
type Bill struct {
	ID          int64   `json:"id,omitempty"`
	AssetID     int64   `json:"assetid"`
	BookID      int64   `json:"bookid"`
	CateID      int64   `json:"cateid"`
	CreateTime  int64   `json:"createtime"`
	DescInfo    string  `json:"descinfo,omitempty"`
	FromID      int64   `json:"fromid"`
	Images      []string `json:"images,omitempty"`
	Money       float64 `json:"money"`
	Platform    int     `json:"platform"`
	Remark      string  `json:"remark"`
	Status      int     `json:"status"`
	TargetID    int64   `json:"targetid"`
	TimeInSec   int64   `json:"time"`
	Type        int     `json:"type"`
	UpdateTime  int64   `json:"updatetime"`
	UserID      string  `json:"userid"`
	Username    string  `json:"username,omitempty"`
}

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

// SyncBills 将账单增量同步到云端。changelist 中的账单会创建或更新，dellist 中的 ID 会被删除。
func (s *Session) SyncBills(changes []Bill, deletes []int64) (*SyncResult, error) {
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

	var result SyncResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse sync result: %w", err)
	}
	if result.Code != 0 {
		return &result, fmt.Errorf("sync failed (code=%d): %s", result.Code, result.Msg)
	}
	return &result, nil
}

// SyncResult 是同步操作的返回结构。
type SyncResult struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// AddBill 添加一笔账单（创建 + 同步）。
func (s *Session) AddBill(bill Bill) (*SyncResult, error) {
	return s.SyncBills([]Bill{bill}, nil)
}

// DeleteBill 删除一笔账单。
func (s *Session) DeleteBill(billID int64) (*SyncResult, error) {
	return s.SyncBills(nil, []int64{billID})
}
