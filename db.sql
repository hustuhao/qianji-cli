-- 钱迹用户账单表 — 完全镜像 Android App 的 user_bill 表
-- 来源: BillDao.createTable() 逆向 (jadx)
-- 表名: user_bill (GreenDAO: BillDao.TABLENAME)

CREATE TABLE IF NOT EXISTS user_bill (
    _id             INTEGER PRIMARY KEY AUTOINCREMENT,
    billid          INTEGER NOT NULL,          -- 服务端唯一 ID
    USERID          TEXT,                      -- 用户 ID
    TIME            INTEGER NOT NULL,          -- 账单时间 (Unix 秒)
    TYPE            INTEGER NOT NULL,          -- 0=支出 1=收入 2=转账 3=信用卡 5=报销 ...
    REMARK          TEXT,                      -- 备注
    MONEY           REAL NOT NULL,             -- 金额
    STATUS          INTEGER NOT NULL,          -- 0=已删除 1=已同步 2=待同步
    CATEGORY_ID     INTEGER NOT NULL,          -- 分类 ID
    IMAGES          TEXT,                      -- 图片 JSON 数组
    PAYTYPE         INTEGER NOT NULL,          -- 支付方式
    updatetime      INTEGER NOT NULL,          -- 更新时间 (Unix 秒)
    createtime      INTEGER NOT NULL,          -- 创建时间 (Unix 秒)
    PLATFORM        INTEGER NOT NULL,          -- 0=手动 1=旧重复 120=重复 121=分期 122=自动
    ASSETID         INTEGER NOT NULL,          -- 资产账户 ID
    FROMID          INTEGER NOT NULL,          -- 转出账户 ID (转账)
    TARGETID        INTEGER NOT NULL,          -- 转入账户 ID (转账)
    EXTRA           TEXT,                      -- 额外信息 JSON (标签/报销/币种)
    DESCINFO        TEXT,                      -- 描述信息
    bookid          INTEGER NOT NULL,          -- 账本 ID
    USERNAME        TEXT,                      -- 用户名
    FROMACT         TEXT,                      -- 转出账户名
    TARGETACT       TEXT,                      -- 转入账户名
    IMPORT_PACK_ID  INTEGER NOT NULL,          -- 导入批次 ID
    BOOK_NAME       TEXT                       -- 账本名称
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_bill_billid ON user_bill (billid);

-- 辅助索引：按时间 + 状态查询（list 命令使用）
CREATE INDEX IF NOT EXISTS idx_user_bill_time_status ON user_bill (TIME, STATUS);
