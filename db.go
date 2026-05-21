// 本地 SQLite 存储 — 完全模仿钱迹 Android 客户端 user_bill 表结构
// 使用 modernc.org/sqlite (纯 Go，无 CGO 依赖)
//
// 表结构来自 BillDao.createTable() 逆向:
//
//	CREATE TABLE "user_bill" (
//	  "_id"             INTEGER PRIMARY KEY AUTOINCREMENT,
//	  "billid"          INTEGER NOT NULL,        -- 服务端唯一 ID
//	  "USERID"          TEXT,
//	  "TIME"            INTEGER NOT NULL,        -- 账单时间 (Unix 秒)
//	  "TYPE"            INTEGER NOT NULL,        -- 0=支出 1=收入 2=转账...
//	  "REMARK"          TEXT,
//	  "MONEY"           REAL NOT NULL,
//	  "STATUS"          INTEGER NOT NULL,        -- 0=deleted 1=synced 2=pending
//	  "CATEGORY_ID"     INTEGER NOT NULL,
//	  "IMAGES"          TEXT,                    -- JSON 数组
//	  "PAYTYPE"         INTEGER NOT NULL,
//	  "updatetime"      INTEGER NOT NULL,
//	  "createtime"      INTEGER NOT NULL,
//	  "PLATFORM"        INTEGER NOT NULL,
//	  "ASSETID"         INTEGER NOT NULL,
//	  "FROMID"          INTEGER NOT NULL,
//	  "TARGETID"        INTEGER NOT NULL,
//	  "EXTRA"           TEXT,                    -- JSON
//	  "DESCINFO"        TEXT,
//	  "bookid"          INTEGER NOT NULL,
//	  "USERNAME"        TEXT,
//	  "FROMACT"         TEXT,
//	  "TARGETACT"       TEXT,
//	  "IMPORT_PACK_ID"  INTEGER NOT NULL,
//	  "BOOK_NAME"       TEXT
//	)
package qianji

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// InitDB 初始化本地数据库。dbPath 为空则用 ~/.qianji/qianji.db。
func InitDB(dbPath string) error {
	if dbPath == "" {
		home, _ := os.UserHomeDir()
		dbPath = filepath.Join(home, ".qianji", "qianji.db")
	}
	os.MkdirAll(filepath.Dir(dbPath), 0700)

	var err error
	db, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	return migrate()
}

func migrate() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_bill (
			_id             INTEGER PRIMARY KEY AUTOINCREMENT,
			billid          INTEGER NOT NULL,
			USERID          TEXT,
			TIME            INTEGER NOT NULL,
			TYPE            INTEGER NOT NULL,
			REMARK          TEXT,
			MONEY           REAL NOT NULL,
			STATUS          INTEGER NOT NULL,
			CATEGORY_ID     INTEGER NOT NULL,
			IMAGES          TEXT,
			PAYTYPE         INTEGER NOT NULL,
			updatetime      INTEGER NOT NULL,
			createtime      INTEGER NOT NULL,
			PLATFORM        INTEGER NOT NULL,
			ASSETID         INTEGER NOT NULL,
			FROMID          INTEGER NOT NULL,
			TARGETID        INTEGER NOT NULL,
			EXTRA           TEXT,
			DESCINFO        TEXT,
			bookid          INTEGER NOT NULL,
			USERNAME        TEXT,
			FROMACT         TEXT,
			TARGETACT       TEXT,
			IMPORT_PACK_ID  INTEGER NOT NULL,
			BOOK_NAME       TEXT
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_user_bill_billid ON user_bill (billid);
	`)
	return err
}

// SaveBills 批量存入本地 DB（upsert by billid）。
func SaveBills(bills []Bill) error {
	if db == nil {
		return fmt.Errorf("db not initialized")
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO user_bill
		(billid, USERID, TIME, TYPE, REMARK, MONEY, STATUS, CATEGORY_ID,
		 IMAGES, PAYTYPE, updatetime, createtime, PLATFORM, ASSETID, FROMID, TARGETID,
		 EXTRA, DESCINFO, bookid, USERNAME, FROMACT, TARGETACT, IMPORT_PACK_ID, BOOK_NAME)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, b := range bills {
		imagesJSON, _ := json.Marshal(b.Images)
		_, err = stmt.Exec(
			b.ID, b.UserID, b.TimeInSec, b.Type, b.Remark, b.Money, b.Status, b.CateID,
			string(imagesJSON), 0, b.UpdateTime, b.CreateTime, b.Platform, b.AssetID, b.FromID, b.TargetID,
			"", b.DescInfo, b.BookID, b.Username, "", "", 0, b.BookName,
		)
		if err != nil {
			return fmt.Errorf("insert bill %d: %w", b.ID, err)
		}
	}
	return tx.Commit()
}

// QueryBillsByDate 按日期查询本地账单（基于 TIME 字段）。
func QueryBillsByDate(t time.Time) ([]Bill, error) {
	if db == nil {
		return nil, fmt.Errorf("db not initialized")
	}
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	endOfDay := startOfDay + 86400

	rows, err := db.Query(`
		SELECT billid, USERID, TIME, TYPE, REMARK, MONEY, STATUS, CATEGORY_ID,
		       IMAGES, updatetime, createtime, PLATFORM, ASSETID, FROMID, TARGETID,
		       DESCINFO, bookid, USERNAME, BOOK_NAME
		FROM user_bill
		WHERE TIME >= ? AND TIME < ? AND STATUS != 0
		ORDER BY TIME DESC
	`, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bills []Bill
	for rows.Next() {
		var b Bill
		var imagesJSON, descInfo, bookName, username, userID, remark sql.NullString
		err := rows.Scan(
			&b.ID, &userID, &b.TimeInSec, &b.Type, &remark, &b.Money, &b.Status, &b.CateID,
			&imagesJSON, &b.UpdateTime, &b.CreateTime, &b.Platform, &b.AssetID, &b.FromID, &b.TargetID,
			&descInfo, &b.BookID, &username, &bookName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		if userID.Valid {
			b.UserID = userID.String
		}
		if remark.Valid {
			b.Remark = remark.String
		}
		if imagesJSON.Valid && imagesJSON.String != "" && imagesJSON.String != "null" {
			json.Unmarshal([]byte(imagesJSON.String), &b.Images)
		}
		if descInfo.Valid {
			b.DescInfo = descInfo.String
		}
		if username.Valid {
			b.Username = username.String
		}
		if bookName.Valid {
			b.BookName = bookName.String
		}
		b.CateName = fmt.Sprintf("#%d", b.CateID)
		bills = append(bills, b)
	}
	return bills, nil
}

// CountBills 返回本地账单总数（用于调试）。
func CountBills() int {
	if db == nil {
		return 0
	}
	var c int
	db.QueryRow("SELECT COUNT(*) FROM user_bill WHERE STATUS != 0").Scan(&c)
	return c
}
