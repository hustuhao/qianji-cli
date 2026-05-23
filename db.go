// 本地 SQLite 存储 — 完全模仿钱迹 Android 客户端 user_bill 表结构
// schema 由 db.sql 维护，编译时通过 embed 嵌入二进制。
package qianji

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed db.sql
var dbSQL string

var db *sql.DB

// InitDB 初始化本地数据库。dbPath 为空则用 ~/.qianji/qianji.db。
func InitDB(dbPath string) error {
	if dbPath == "" {
		home, _ := os.UserHomeDir()
		dbPath = filepath.Join(home, ".qianji", "qianji.db")
	}
	os.MkdirAll(filepath.Dir(dbPath), 0700)
	return initDBWithDSN(dbPath + "?_journal_mode=WAL")
}

func initDBWithDSN(dsn string) error {
	var err error
	db, err = sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	_, err = db.Exec(dbSQL)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
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
		INSERT INTO user_bill
		(billid, USERID, TIME, TYPE, REMARK, MONEY, STATUS, CATEGORY_ID,
		 IMAGES, PAYTYPE, updatetime, createtime, PLATFORM, ASSETID, FROMID, TARGETID,
		 EXTRA, DESCINFO, bookid, USERNAME, FROMACT, TARGETACT, IMPORT_PACK_ID, BOOK_NAME)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(billid) DO UPDATE SET
			USERID=excluded.USERID, TIME=excluded.TIME, TYPE=excluded.TYPE,
			REMARK=excluded.REMARK, MONEY=excluded.MONEY, STATUS=excluded.STATUS,
			CATEGORY_ID=excluded.CATEGORY_ID, IMAGES=excluded.IMAGES,
			updatetime=excluded.updatetime, createtime=excluded.createtime,
			PLATFORM=excluded.PLATFORM, ASSETID=excluded.ASSETID,
			FROMID=excluded.FROMID, TARGETID=excluded.TARGETID,
			EXTRA=excluded.EXTRA, DESCINFO=excluded.DESCINFO,
			bookid=excluded.bookid, USERNAME=excluded.USERNAME,
			BOOK_NAME=excluded.BOOK_NAME
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

// ---- 查询 ----

const billColumns = `billid, USERID, TIME, TYPE, REMARK, MONEY, STATUS, CATEGORY_ID,
	IMAGES, updatetime, createtime, PLATFORM, ASSETID, FROMID, TARGETID,
	DESCINFO, bookid, USERNAME, BOOK_NAME`

func scanBills(rows *sql.Rows) ([]Bill, error) {
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
		bills = append(bills, b)
	}
	return bills, rows.Err()
}

// QueryBillsByDate 按日期查询本地账单。
func QueryBillsByDate(t time.Time) ([]Bill, error) {
	if db == nil {
		return nil, fmt.Errorf("db not initialized")
	}
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
	end := start + 86400

	rows, err := db.Query(`
		SELECT `+billColumns+`
		FROM user_bill WHERE TIME >= ? AND TIME < ? AND STATUS != 0
		ORDER BY TIME DESC
	`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBills(rows)
}

// QueryAllBills 返回本地全部非删除账单。
func QueryAllBills() ([]Bill, error) {
	if db == nil {
		return nil, fmt.Errorf("db not initialized")
	}
	rows, err := db.Query(`
		SELECT ` + billColumns + `
		FROM user_bill WHERE STATUS != 0 ORDER BY TIME DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBills(rows)
}

// QueryPendingBills 返回本地待同步账单（status=2）。
func QueryPendingBills() ([]Bill, error) {
	if db == nil {
		return nil, fmt.Errorf("db not initialized")
	}
	rows, err := db.Query(`
		SELECT ` + billColumns + `
		FROM user_bill WHERE STATUS = 2 ORDER BY TIME
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBills(rows)
}

// QueryBillByID 按 billid 查询单笔账单。
func QueryBillByID(billID int64) (*Bill, error) {
	if db == nil {
		return nil, fmt.Errorf("db not initialized")
	}
	rows, err := db.Query(`
		SELECT `+billColumns+` FROM user_bill WHERE billid = ?
	`, billID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bills, err := scanBills(rows)
	if err != nil {
		return nil, err
	}
	if len(bills) == 0 {
		return nil, fmt.Errorf("bill %d not found", billID)
	}
	return &bills[0], nil
}

// MarkSynced 将指定账单 ID 标记为已同步（status=1）。
func MarkSynced(billIDs []int64) error {
	if db == nil || len(billIDs) == 0 {
		return nil
	}
	q := "UPDATE user_bill SET STATUS=1 WHERE billid IN ("
	args := make([]interface{}, len(billIDs))
	for i, id := range billIDs {
		if i > 0 {
			q += ","
		}
		q += "?"
		args[i] = id
	}
	q += ")"
	_, err := db.Exec(q, args...)
	return err
}

// SaveCategories 批量保存分类（INSERT OR REPLACE）。
func SaveCategories(cats []Category) error {
	if db == nil || len(cats) == 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO category VALUES (?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, c := range cats {
		_, err = stmt.Exec(c.ID, c.Name, c.Icon, c.Type, c.Sort, c.UserID, c.Editable, c.BookID, c.ParentID, c.Level)
		if err != nil {
			return fmt.Errorf("insert category %d: %w", c.ID, err)
		}
	}
	return tx.Commit()
}

// CountBills 返回本地账单总数。
func CountBills() int {
	if db == nil {
		return 0
	}
	var c int
	db.QueryRow("SELECT COUNT(*) FROM user_bill WHERE STATUS != 0").Scan(&c)
	return c
}

// ---- 辅助：读取外部数据库 ----

// OpenReadOnly 以只读模式打开一个外部 SQLite 文件。
func OpenReadOnly(path string) (*sql.DB, error) {
	return sql.Open("sqlite", "file:"+path+"?mode=ro")
}

// OpenForWrite 以读写模式打开一个外部 SQLite 文件。
func OpenForWrite(path string) (*sql.DB, error) {
	return sql.Open("sqlite", "file:"+path)
}

// QueryAllFrom 从指定连接查询全部非删除账单。
func QueryAllFrom(extDB *sql.DB) ([]Bill, error) {
	rows, err := extDB.Query(`
		SELECT ` + billColumns + `
		FROM user_bill WHERE STATUS != 0 ORDER BY TIME DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBills(rows)
}

// SaveTo 将账单写入指定的数据库连接。
func SaveTo(extDB *sql.DB, bills []Bill) error {
	tx, err := extDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO user_bill
		(billid, USERID, TIME, TYPE, REMARK, MONEY, STATUS, CATEGORY_ID,
		 IMAGES, PAYTYPE, updatetime, createtime, PLATFORM, ASSETID, FROMID, TARGETID,
		 EXTRA, DESCINFO, bookid, USERNAME, FROMACT, TARGETACT, IMPORT_PACK_ID, BOOK_NAME)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(billid) DO UPDATE SET
			USERID=excluded.USERID, TIME=excluded.TIME, TYPE=excluded.TYPE,
			REMARK=excluded.REMARK, MONEY=excluded.MONEY, STATUS=excluded.STATUS,
			CATEGORY_ID=excluded.CATEGORY_ID, IMAGES=excluded.IMAGES,
			updatetime=excluded.updatetime, createtime=excluded.createtime,
			PLATFORM=excluded.PLATFORM, ASSETID=excluded.ASSETID,
			FROMID=excluded.FROMID, TARGETID=excluded.TARGETID,
			EXTRA=excluded.EXTRA, DESCINFO=excluded.DESCINFO,
			bookid=excluded.bookid, USERNAME=excluded.USERNAME,
			BOOK_NAME=excluded.BOOK_NAME
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
