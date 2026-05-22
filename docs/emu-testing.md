# 模拟器联合测试方法

本文档记录用 Android 模拟器和 CLI 进行联合测试的方法论，包括数据库对比、devid 隔离验证、同步状态检查等。

## 一、环境准备

### 1.1 模拟器连接

```bash
# WiFi adb 连接（TCP 模式）
adb connect 127.0.0.1:16448

# 确认设备在线
adb -s 127.0.0.1:16448 shell getprop ro.product.model

# 获取 root（模拟器通常自带）
adb -s 127.0.0.1:16448 root
```

### 1.2 关键路径

| 项目 | 路径 |
|------|------|
| 应用包名 | `com.mutangtech.qianji` |
| 数据库 | `/data/data/com.mutangtech.qianji/databases/qianjiapp` |
| 用户 MMKV | `/data/data/com.mutangtech.qianji/files/mmkv/kvu_<uid>` |
| 系统 MMKV | `/data/data/com.mutangtech.qianji/files/mmkv/kv_system` |
| CLI 数据库 | `~/.qianji/qianji.db` |
| CLI token | `~/.qianji_token.json` |
| CLI devid | `~/.qianji/device_id` |

## 二、数据库对比

### 2.1 拉取模拟器数据库

```bash
# 复制到 sdcard（需要 root）
adb -s 127.0.0.1:16448 shell \
  "cp /data/data/com.mutangtech.qianji/databases/qianjiapp /sdcard/qianji_emu.db"

# 拉到本地
adb -s 127.0.0.1:16448 pull /sdcard/qianji_emu.db /tmp/qianji_emu.db
```

注意：数据库文件名是 `qianjiapp`（无后缀），不是 `qianji.db`。

### 2.2 常用 SQL 查询

```sql
-- 总条数对比
SELECT count(*) FROM user_bill;

-- bookid 分布（发现默认值不匹配）
SELECT bookid, count(*) FROM user_bill GROUP BY bookid;

-- 搜索测试账单（确认同步结果）
SELECT bookid, REMARK, MONEY, CATEGORY_ID, billid 
FROM user_bill WHERE REMARK='打车';

-- 最新 5 笔
SELECT bookid, REMARK, MONEY FROM user_bill 
ORDER BY billid DESC LIMIT 5;
```

### 2.3 对比方法

```bash
# 两边各自搜索测试备注
sqlite3 ~/.qianji/qianji.db "SELECT REMARK, MONEY, CATEGORY_ID FROM user_bill WHERE REMARK='打车';"
sqlite3 /tmp/qianji_emu.db "SELECT REMARK, MONEY, CATEGORY_ID FROM user_bill WHERE REMARK='打车';"

# 同步后检查条数
echo "CLI: $(sqlite3 ~/.qianji/qianji.db 'SELECT count(*) FROM user_bill')"
echo "EMU: $(sqlite3 /tmp/qianji_emu.db 'SELECT count(*) FROM user_bill')"
```

## 三、同步状态检查

### 3.1 MMKV 提取

模拟器的同步状态不存储在 SQLite 中，而是存在 MMKV（内存映射键值存储）。

```bash
# 拉取用户级 MMKV
adb -s 127.0.0.1:16448 pull \
  /data/data/com.mutangtech.qianji/files/mmkv/kvu_<uid> /tmp/kvu_emu

# 拉取系统级 MMKV（含 devid）
adb -s 127.0.0.1:16448 pull \
  /data/data/com.mutangtech.qianji/files/mmkv/kv_system /tmp/kv_system_emu

# 提取可读字符串
strings /tmp/kvu_emu | grep 'syncbook_times_'
strings /tmp/kv_system_emu | grep 'app_user_device_id'
```

### 3.2 关键 MMKV 键

| 键名 | 含义 | 示例值 |
|------|------|--------|
| `syncbook_times_` | 各账本上次同步时间（JSON） | `{"-1":1779440758}` |
| `app_user_device_id` | 设备唯一 ID | `88554fd02d3b51af7cdb00f0ab80e879` |
| `tags_sync_time` | 标签同步时间 | 时间戳 |
| `last_update_catelist_<bookid>` | 分类列表更新时间 | 时间戳 |

### 3.3 lasttimes 时间戳解读

```bash
# 转换 MMKV 中的时间戳
date -r 1779440758
# → Fri May 22 17:05:58 CST 2026
```

`syncbook_times_` 存储各 bookid 的上次同步时间，用于增量同步的 `lasttimes` 参数。服务端按 `updatetime > lasttimes` 过滤，因此：
- 账单时间 **早于** lasttimes → 被过滤，不可见
- 账单时间 **晚于** lasttimes → 被返回

**典型陷阱**：使用 `qianji add -t "过去时间"` 创建的账单会被增量同步过滤掉。验证跨设备同步时应使用当前时间。

## 四、DevID 隔离验证

### 4.1 提取双方 devid

```bash
# CLI devid
cat ~/.qianji/device_id
# → 430602e1fcc17f8c4f145639827ccae4

# 模拟器 devid
strings /tmp/kv_system_emu | grep 'app_user_device_id'
# → app_user_device_id! 88554fd02d3b51af7cdb00f0ab80e879
```

### 4.2 跨 devid Pull 测试

服务端按 devid 隔离数据，但 pull 可以跨 devid 返回账单。用 Go 测试脚本直接发起 HTTP 请求：

```go
// 用模拟器的 devid 发 pull 请求
emuDevID := "88554fd02d3b51af7cdb00f0ab80e879"
resp, _ := client.DoPostCustom("syncv2", "pull", params, token, emuDevID)
```

验证方法：

1. CLI 用 `./qianji add -c 89693183 金额 备注` 推送一笔
2. 分别用 CLI devid 和模拟器 devid 拉取
3. 确认两边都能在 pull 返回的 `changes` 中找到该账单

### 4.3 用模拟器 devid 推送测试

```go
// 用模拟器 devid 推送以验证服务端分类信任
emuDevID := "88554fd02d3b51af7cdb00f0ab80e879"
resp, _ := client.DoPostCustom("bill", "syncall", params, token, emuDevID)
```

测试发现：用模拟器 devid 推送时服务端保留 cateid，用 CLI devid 推送时服务端有时重置 — 这是因为 CLI 的 `NewBill()` 未设置 `userid` 字段。

## 五、服务端响应验证

### 5.1 Syncall 响应分析

```json
{
  "ec": 200,
  "data": {
    "sync_result": {
      "bill": {
        "new_ids": [1779444285975432],
        "new_count": 0,
        "has_failed": false
      }
    }
  }
}
```

**关键指标**：
- `ec=200`：请求成功，但**不等于账单被存储**
- `new_ids` 非空：服务端确认收到
- `has_failed=false`：无格式错误
- `new_count=0`：正常（此字段含义不明，但 0 不代表失败）

即使 syncall 返回以上全部正常信号，**仍可能**出现：
- Pull 不返回该账单 → 检查 `BillID` 是否为 0
- 分类被重置 → 检查 `userid` 是否为空

### 5.2 Pull 响应中搜索测试账单

```go
for _, c := range pullResult.Changes {
    if c.Remark == "测试备注" {
        fmt.Printf("cateid=%d catename=%s\n", c.CateID, c.CateName)
    }
}
```

服务端在 pull 中返回的 `cateid` 是权威值 — 它是服务端实际存储的分类。

## 六、典型调试流程

### 6.1 账单不出现的排查

```
1. 确认 syncall 返回 ec=200 且 has_failed=false
   ↓
2. 用双方 devid 分别 pull，检查 changes 中是否有该账单
   ↓
3. 如果有 → 检查模拟器 MMKV 的 lasttimes
   如果 lasttimes > 账单 updatetime → 增量同步过滤，需全量同步或等下次增量
   ↓
4. 如果没有 → 检查 syncall payload
   - BillID 是否为 0（服务端丢弃）
   - userid 是否为空
   - 所有必填字段是否存在
```

### 6.2 分类错误的排查

```
1. 拉取服务端 pull → 检查 cateid 是否为推送的值
   ↓
2. 如果服务端 cateid ≠ 推送的 cateid → payload 有问题
   ↓
3. 检查 syncall JSON：
   - "cateid" 字段是否正确（不是 "category" 或 "cate_id"）
   - "userid" 是否为空（空值导致服务器重置所有字段）
   - omitempty 是否意外省略了字段
   ↓
4. 用模拟器 devid 推送同格式 payload 作为对照组
```

### 6.3 加载逆向 skill 参考

当需要查阅 Java 源码或协议细节时：

```bash
# 加载钱迹逆向知识
skill_view("reverse-engineering-android-apis", "references/qianji.md")
```

## 七、测试数据清理

```bash
# 删除所有测试账单
sqlite3 ~/.qianji/qianji.db \
  "SELECT billid FROM user_bill WHERE REMARK IN ('测试1','测试2',...);" \
  | while read id; do ./qianji delete $id; done

# 确认清理
sqlite3 ~/.qianji/qianji.db "SELECT count(*) FROM user_bill;"
```

注意：`qianji delete` 会通过 syncall dellist 同步删除到服务端，模拟器下拉刷新后也会删除对应账单。
