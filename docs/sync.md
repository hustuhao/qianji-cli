# 钱迹同步机制文档

> 基于 APK 逆向分析 + 实际测试验证

---

## 一、架构概览

钱迹采用 **客户端优先 + 双向同步** 架构：

```
┌──────────────┐                              ┌──────────────┐
│  手机 App     │                              │  CLI          │
│  (SQLite)    │                              │  (SQLite)    │
└──────┬───────┘                              └──────┬───────┘
       │                                             │
       │  PUSH: bill/syncall                          │
       │  ──────────────────►  ┌──────────┐  ◄────── │
       │                      │  服务端    │          │
       │  PULL: syncv2/pull   │           │          │
       │  ◄────────────────── │           │ ──────► │
       │                      └──────────┘          │
       │                                             │
```

**两个同步端点：**

| 端点 | 方向 | 作用 | 实现 |
|------|------|------|------|
| `bill/syncall` | 客户端 → 服务端 | PUSH: 推送本地变更，返回 ID 确认 | `ne.a` |
| `syncv2/pull` | 服务端 → 客户端 | PULL: 拉取其他设备的账单、分类、删除 | `le.a` |

**完整同步流程**（`je.c` 协调）：

```
FullSync:
  1. le.a.start()  → syncv2/pull  → 循环拉取（分页）
  2. ne.a.start()  → bill/syncall → 推送本地变更
  3. onSyncFinished(pullCount + pushCount)
```

---

## 二、PULL 协议: `syncv2/pull`

### 2.1 请求

```
POST /syncv2/pull
Header: utoken: <token>

Body (form-urlencoded):
  uid=<user_id>
  bookid=-1                         // -1=全部账本
  lasttimes=<json>                  // 可选，{bookId: lastSyncTimestamp} 增量同步
  pageoffset=<int64>                // 分页偏移
  pagesign=<string>                 // 分页游标
```

### 2.2 响应

```json
{
  "ec": 200,
  "em": "",
  "data": {
    "changes": [
      {
        "id": 1700746925900113295,
        "billid": 1700746925900113295,
        "userid": "23112360655f55d36b693",
        "time": 1700746877,
        "type": 0,
        "remark": "中饭",
        "money": 13.0,
        "status": 1,
        "cateid": 89693180,
        "platform": 0,
        "bookid": -1,
        "assetid": -1,
        "fromid": 0,
        "targetid": 0,
        "descinfo": "",
        "images": [],
        "username": "",
        "catename": "",
        "assetname": "",
        "bookname": ""
      }
    ],
    "deletes": [],
    "categories": [],
    "bookid": -1,
    "pageoffset": 0,
    "pagesign": "",
    "hasmore": 0,
    "count": 187
  }
}
```

### 2.3 分页机制

```
首次请求: pageoffset=0, pagesign=""
循环: 直到 hasmore=0
  - 每次拉取一页数据
  - pageoffset = 响应中的 pageoffset
  - pagesign = 响应中的 pagesign
  - count = 本页条数
```

### 2.4 增量同步

`lasttimes` 参数可追溯上次同步时间，实现增量拉取：

```json
{"-1": 1779300000000}
```

首次同步不传 `lasttimes`，全量拉取。

### 2.5 客户端保存

```java
// l.java: savePullResult()
public void savePullResult(le.c cVar) {
    saveList(cVar.changes);      // INSERT OR REPLACE user_bill
    deleteListByPK(cVar.deletes); // DELETE FROM user_bill
    saveList(cVar.categoryList);  // 更新分类
}
```

---

## 三、PUSH 协议: `bill/syncall`

### 3.1 请求

```
POST /bill/syncall
Body:
  uid=<user_id>
  v={"bills":{"changelist":[...],"dellist":[...]}}
```

### 3.2 响应

```json
{
  "ec": 200,
  "data": {
    "sync_result": {
      "bill": {
        "new_ids": [],
        "update_ids": [],
        "conf_ids": [],
        "del_ids": [],
        "new_count": 0,
        "change_total": 0
      }
    }
  }
}
```

**关键：** syncall 只返回 ID 数组，不返回账单正文。

### 3.3 客户端处理

```java
// l.java: saveSyncedResult()
public void saveSyncedResult(BillSyncResult result) {
    // new_ids: 标记 status=1
    // update_ids: 标记 status=1
    // conf_ids: 重新分配 billid
    // del_ids: DELETE FROM user_bill
}
```

---

## 四、客户端数据库

表: `user_bill` (26 字段，由 `BillDao.createTable()` 定义)

### STATUS 字段

| 值 | 含义 |
|----|------|
| 0 | 已删除 |
| 1 | 已同步 |
| 2 | 待同步 |

### billid

格式: `毫秒时间戳 * 1000 + 随机数(0-999)`，约 17-18 位。

---

## 五、CLI 实现

### 5.1 qianji sync

```
1. PullBills(-1, "", 0, "") → 循环分页拉取
2. QueryPendingBills() → 获取本地 status=2
3. SyncBills(pending, nil) → 推送本地到服务端
4. SaveBills(pulled) → 拉取的存入本地
5. MarkSynced(pending) → 标记本地已同步
```

### 5.2 qianji add

```
1. NewBill() → 构建 Bill (status=2)
2. SyncBills([bill], nil) → 推送到服务端
3. SaveBills([bill]) → 存入本地
4. MarkSynced([bill.id]) → 标记已同步
```

### 5.3 qianji list

```
QueryBillsByDate() → 从本地 SQLite 按日期查询
```

### 5.4 文件职责

| 文件 | 职责 |
|------|------|
| `bill.go` | Bill 模型, SyncBills, PullBills, FullSync |
| `db.go` / `db.sql` | 本地 SQLite (user_bill 表) |
| `sign.go` | API 签名 (libfabricsuffer.so 逆向) |
| `curl.go` | HTTP 请求后端 |
| `cmd/qianji/main.go` | CLI 入口 |
| `cmd/qianji/sync.go` | sync 命令 |

---

## 六、API 端点清单

| 端点 | 方法 | 说明 |
|------|------|------|
| account/login | POST | 登录 |
| bill/syncall | POST | PUSH 同步 |
| **syncv2/pull** | POST | **PULL 同步** |
| category/listv2 | POST | 分类列表 |
| book/list | POST | 账本列表 |
| asset/list | POST | 资产账户 |

---

## 七、签名机制

```
combined  = sprintf("%d%d%s%s", versionCode, millis, pkg, ctrl)
reqId     = MD5(combined + "free20170908&x_*1127")
encReqId  = MD5(reqId + "michaeljackson")
tok       = MD5(reqId + "1172020" + ctrl + encReqId + act)
```

---

## 八、关键源码文件索引

| APK 源码 | 作用 |
|---------|------|
| `je/c.java` | 同步总协调器 (PULL + PUSH) |
| `le/a.java` | PULL 协调器 (syncv2/pull) |
| `ne/a.java` | PUSH 协调器 (bill/syncall) |
| `pc/b.java` | PULL API 构建器 |
| `com/.../api/bill/d.java` | syncall API 构建器 |
| `com/.../api/bill/g.java` | PULL 响应解析器 |
| `com/.../api/bill/h.java` | syncall 响应解析器 |
| `le/c.java` | PULL 响应模型 (changes+deletes) |
| `le/d.java` | syncall 响应模型 (BillSyncResult) |
| `com/.../data/model/BillDao.java` | user_bill 表定义 |
| `com/.../data/db/dbhelper/l.java` | savePullResult + saveSyncedResult |
| `com/.../arc/http/JNIHelper.java` | 签名入口 |
| `lib/arm64-v8a/libfabricsuffer.so` | 签名算法实现 |
