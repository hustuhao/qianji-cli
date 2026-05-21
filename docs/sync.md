# 钱迹同步机制文档

> 基于 APK 逆向分析 (jadx) + 实际测试验证

---

## 一、架构概览

钱迹采用 **客户端优先** 的同步架构：

```
┌──────────────┐     syncall (PUSH)     ┌──────────────┐
│  手机 App     │ ◄──────────────────► │  服务端       │
│  (SQLite)    │     ID 确认            │  (仅存 delta) │
└──────────────┘                        └──────────────┘
       │
       │ adb pull (qianji sync)
       ▼
┌──────────────┐
│  CLI          │
│  (SQLite)    │
└──────────────┘
```

**核心特征：**

- 服务端不存储账单全文，只接收和确认客户端推送的 delta
- 每个客户端独立维护完整的 SQLite 数据库副本
- `bill/syncall` 是 PUSH 协议：客户端推、服务端确认 ID
- 不存在"拉取全部账单"的 API
- 多设备同步通过各自维护本地库 + 定期合并实现

---

## 二、客户端数据库

### 2.1 表结构

表名: `user_bill` (由 `BillDao.java` 定义，GreenDAO ORM)

```sql
CREATE TABLE user_bill (
    _id             INTEGER PRIMARY KEY AUTOINCREMENT,
    billid          INTEGER NOT NULL,          -- 服务端唯一 ID
    USERID          TEXT,                      -- 用户 ID
    TIME            INTEGER NOT NULL,          -- 账单时间 (Unix 秒)
    TYPE            INTEGER NOT NULL,          -- 0=支出 1=收入 2=转账 3=信用卡 5=报销...
    REMARK          TEXT,                      -- 备注
    MONEY           REAL NOT NULL,             -- 金额
    STATUS          INTEGER NOT NULL,          -- 0=已删除 1=已同步 2=待同步
    CATEGORY_ID     INTEGER NOT NULL,          -- 分类 ID
    IMAGES          TEXT,                      -- 图片 JSON 数组
    PAYTYPE         INTEGER NOT NULL,          -- 支付方式
    updatetime      INTEGER NOT NULL,          -- 更新时间 (Unix 秒)
    createtime      INTEGER NOT NULL,          -- 创建时间 (Unix 秒)
    PLATFORM        INTEGER NOT NULL,          -- 0=手动 120=重复任务 121=分期 122=自动
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

CREATE UNIQUE INDEX idx_user_bill_billid ON user_bill (billid);
```

### 2.2 STATUS 字段

| 值 | 含义 | 说明 |
|----|------|------|
| 0 | 已删除 | 逻辑删除，syncall 的 dellist 推送后服务端也删 |
| 1 | 已同步 | 服务端已确认，本地与云端一致 |
| 2 | 待同步 | 本地新建或修改，下次 syncall 推送到服务端 |

### 2.3 TYPE 字段

| 值 | 含义 |
|----|------|
| 0 | 支出 |
| 1 | 收入 |
| 2 | 转账 |
| 3 | 信用卡还款 |
| 5 | 报销 |
| 6 | 债务-DEBT |
| 7 | 债务-LOAN |
| 20 | 退款 |

### 2.4 billid 生成

```java
// n8.j.f()
public static long f() {
    return System.currentTimeMillis() * 1000 + random.nextInt(1000);
}
```

格式: `毫秒时间戳 * 1000 + 随机数(0-999)`，约 17-18 位长整数，全局唯一概率极高。

---

## 三、syncall 协议详解

### 3.1 请求

```
POST /bill/syncall
Header: utoken: <token>
Header: ctrl: bill
Header: act: syncall
Header: reqidv2: <签名>
Header: tok: <签名>

Body (form-urlencoded):
  uid=<user_id>
  v=<JSON>
```

`v` 参数 JSON 结构:

```json
{
  "bills": {
    "changelist": [
      {
        "id": 1700746925900113295,
        "assetid": 1001,
        "bookid": 1,
        "cateid": 89693180,
        "createtime": 1700746925,
        "descinfo": "",
        "extra": {},
        "fromid": 0,
        "images": [],
        "money": 13.0,
        "platform": 0,
        "remark": "中饭",
        "status": 2,
        "targetid": 0,
        "time": 1700746877,
        "type": 0,
        "updatetime": 1700746925,
        "userid": "23112360655f55d36b693",
        "username": ""
      }
    ],
    "dellist": [1700746925900113295]
  }
}
```

**changelist** 中的字段来自 `Bill.toSyncJson()`，移除了以下无关字段:
- `category` (分类对象)
- `fromact` (转出账户名)
- `targetact` (转入账户名)
- `paytype` (支付方式)
- `bookname` (账本名)
- `images` (图片列表 — 部分情况)

### 3.2 响应

```json
{
  "ec": 200,
  "em": "",
  "data": {
    "sync_result": {
      "bill": {
        "change_total": 0,
        "conf_ids": [],
        "del_ids": [],
        "del_total": 0,
        "has_failed": false,
        "new_count": 0,
        "new_ids": [],
        "update_count": 0,
        "update_ids": [],
        "userid": "23112360655f55d36b693"
      }
    }
  }
}
```

### 3.3 响应字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `ec` | int | 200=成功 |
| `data.sync_result.bill` | object | 服务端处理结果 |
| `bill.new_ids` | long[] | 新创建的账单 ID（服务端分配的新 billid） |
| `bill.update_ids` | long[] | 已更新的账单 ID |
| `bill.conf_ids` | long[] | ID 冲突的账单（需要重新分配 billid） |
| `bill.del_ids` | long[] | 服务端要求删除的账单 ID |
| `bill.has_failed` | bool | 是否有失败项 |
| `bill.change_total` | int | changelist 总数 |
| `bill.new_count` | int | 新建数量 |

**关键发现:** 服务端不返回账单正文。`sync_result` 只包含 ID 数组，不含完整 Bill 对象。

### 3.4 客户端处理响应 (`saveSyncedResult`)

```java
// l.java: saveSyncedResult(BillSyncResult)
public void saveSyncedResult(BillSyncResult result) {
    // 1. new_ids: 本地 bill 状态改为 1（已同步）
    for (long id : result.new_ids) {
        Bill bill = findByBillId(id);
        bill.setStatus(1);
        save(bill);
    }

    // 2. update_ids: 同上，标记已同步
    for (long id : result.update_ids) {
        Bill bill = findByBillId(id);
        bill.setStatus(1);
        save(bill);
    }

    // 3. conf_ids: ID 冲突，分配新的 billid
    for (long id : result.conf_ids) {
        Bill bill = findByBillId(id);
        bill.setBillid(generateNewId());
        save(bill);
    }

    // 4. del_ids: 本地删除
    for (long id : result.del_ids) {
        deleteByBillId(id);
    }
}
```

---

## 四、完整同步流程

### 4.1 手机 App 的同步循环

```
入口: ne.a.start(context, false)

循环 (最多 5 次):
  1. e(): 从本地 DB 查 status=2 的 bill
          拆分为 changelist (status=2) + dellist (status=0)
  
  2. a(): 构建 JSON payload {"bills":{"changelist":[...],"dellist":[...]}}
  
  3. d(): POST /bill/syncall → 获得响应
  
  4. saveSyncedResult():
     - 将 new_ids/update_ids 中的 bill 标记为 status=1
     - 处理 conf_ids (重新分配 billid)
     - 删除 del_ids 中的 bill

  5. 如果还有 status=2 的 bill (新产生的), 继续循环
     最多 5 次, 每次最多 200 条
```

### 4.2 CLI 的同步流程

```
1. add 命令:
   本地创建 Bill (status=2)
   → SyncBills(changelist=[bill])
   → POST /bill/syncall
   → MarkSynced([bill.billid])

2. sync 命令:
   推送本地 status=2 的 bill → 服务端
   adb pull 手机库 → INSERT OR REPLACE 进本地库

3. list 命令:
   查询本地 DB WHERE TIME 在今天范围内
```

---

## 五、多设备同步限制

### 5.1 已知限制

| 场景 | 是否支持 | 说明 |
|------|---------|------|
| 设备 A 记账，设备 A 查询 | ✅ | 本地 DB 直接查 |
| 设备 A 记账，CLI 查询 | ❌ | syncall 不返回他设备账单 |
| CLI 记账，手机查询 | ❌ | 同上 |
| 设备 A 记账→sync→设备 B sync | ❌ | 服务端无跨设备推送 |

### 5.2 根本原因

服务端不区分客户端实例。它认为所有请求来自同一个客户端（同一个 token）。
`syncall` 的"服务端变更"只针对当前客户端正在推送的批次——如果服务端
检测到 ID 冲突或需要合并，才返回 `conf_ids`。

不存在"设备 A 推了一条账单，服务端记住它，设备 B 拉取时返回"这种机制。

### 5.3 解决方案

CLI 通过 `qianji sync` 命令从手机 adb pull 数据库来合并且合并数据:

```
每运行一次 qianji sync:
  1. 推送 CLI 本地待同步账单 → 服务端
  2. adb root + pull 手机 SQLite 文件
  3. INSERT OR REPLACE 所有手机账单 (status=1) 进本地库
```

合并后 CLI 本地库包含两端的全部账单。

---

## 六、签名机制

所有 API 请求必须携带签名头 `reqidv2` 和 `tok`。

```
combined  = sprintf("%d%d%s%s", versionCode, millis, pkg, ctrl)
reqId     = MD5(combined + "free20170908&x_*1127")
encReqId  = MD5(reqId + "michaeljackson")
tok       = MD5(reqId + "1172020" + ctrl + encReqId + act)

Header:
  reqidv2: <reqId>
  tok:     <tok>
```

实现文件: `libfabricsuffer.so` → `sign.go`

---

## 七、API 端点清单

### 7.1 已实现的端点

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | account/login | 登录，返回 token + user |
| POST | bill/syncall | 核心同步协议 |
| POST | category/listv2 | 分类列表 |
| POST | book/list | 账本列表 |
| POST | asset/list | 资产账户列表 |

### 7.2 未实现但存在的重要端点

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | bill/refund2 | 退款 |
| POST | billbatch/deletev2 | 批量删除 |
| POST | budget/submitv2 | 预算管理 |
| POST | tag/syncall | 标签同步 |
| POST | repeat_task/* | 周期记账 |
| POST | assetline/init | 资产流水初始化 |

---

## 八、文件索引

| 项目文件 | 对应 APK 源码 |
|---------|-------------|
| `sign.go` | `JNIHelper.java` + `libfabricsuffer.so` |
| `db.go` / `db.sql` | `BillDao.java` |
| `client.go` | `g8/a.java` |
| `curl.go` | Volley `vi/c.java` |
| `bill.go` | `com/mutangtech/qianji/network/api/bill/d.java` |
| `login.go` | `com/mutangtech/qianji/network/api/account/a.java` |
