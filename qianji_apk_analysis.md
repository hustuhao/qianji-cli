# 钱迹 APK 逆向分析报告

> 分析目标: `/Users/wepie/Documents/qianji_437_1170_gw.apk`
> 分析日期: 2026-05-21
> 版本: 4.3.7 (versionCode: 1170)

---

## 一、API 架构

| 项目 | 值 |
|------|-----|
| **API 域名** | `https://api.qianjiapp.com/` |
| **备用域名** | `https://qianji.xxoojoke.com/` |
| **海外域名** | `https://qianjiga.litangkj.com/` |
| **HTTP 库** | Volley（非 Retrofit） |
| **URL 格式** | `/{module}/{action}?uid=xxx&v=json&...` |
| **认证方式** | Header `utoken: <token>` |
| **Token 获取** | `account/login` 返回 `token` 字段 |
| **响应格式** | `{"ec":200,"em":"","data":{...}}` — `ec=200` 表示成功 |
| **配置类** | `ac/a.java` (getApiHost → `b("api_host", d.HOST_DEFAULT)`) |
| **域名常量** | `x8/d.java` |
| **URL 构建器** | `gi/c.java` |
| **基础参数** | `g8/c.java` |
| **通用回调** | `g8/b.java` |

API 域名可通过远程配置覆盖（key: `api_host`），但默认值和备用域名都已硬编码。

---

## 二、完整 API 端点清单

### 账户模块 (`account/`)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | account/login | 登录，返回 token + user |
| POST | account/register | 注册 |
| POST | account/bind | 第三方登录绑定 |
| POST | account/bindaccount | 绑定账号 |
| POST | account/bindwx | 绑定微信 |
| POST | account/loginwx | 微信登录 |
| POST | account/logout | 登出 |
| POST | account/refreshtoken | 刷新 token |
| POST | account/getverifycode | 获取验证码 |
| POST | account/setpwd | 设置密码 |
| POST | account/cancelaccount | 注销账号 |
| POST | account/cancelthirdaccount | 取消第三方 |
| POST | account/unbind | 解绑第三方 |
| POST | account/loginrecords2 | 登录记录 |
| POST | account/removelogin | 删除登录记录 |

### 账单模块 (`bill/`, `billbatch/`, `baoxiao/`)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | bill/syncall | **PUSH 同步** |
| POST | **syncv2/pull** | **PULL 同步** |
| POST | bill/refund2 | 退款 |
| POST | billbatch/deletev2 | 批量删除 |
| POST | billbatch/modify | 批量修改 |
| POST | baoxiao/baoxiao | 报销 |
| POST | baoxiao/cancelbaoxiao | 取消报销 |
| POST | baoxiao/upgradev2 | 升级 v2 |

### 分类模块 (`category/`)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | category/listv2 | 分类列表 |
| POST | category/submitv2 | 提交分类 |
| POST | category/submitmultiple | 批量提交 |
| POST | category/deletev2 | 删除 |
| POST | category/deletecheckv2 | 删除检查 |
| POST | category/reorder3 | 排序 |
| POST | category/movesubcate | 移动子分类 |
| POST | category/changelevel | 修改层级 |
| GET | category/iconlist | 图标列表 |

### 资产模块 (`asset/`, `asset_group/`, `assetline/`)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | asset/list | 资产列表 |
| POST | asset/submit | 提交资产 |
| POST | asset/delete | 删除 |
| POST | asset/reorder | 排序 |
| POST | asset/info | 资产详情 |
| POST | asset/typelist | 资产类型 |
| POST | asset/visible | 可见性 |
| POST | asset/banklist | 银行列表 |
| POST | asset/finish | 结清 |
| GET | asset/listloan | 贷款列表 |
| POST | asset_group/addasset | 添加资产到分组 |
| POST | asset_group/delete | 删除分组 |
| POST | asset_group/removeasset | 移出资产 |
| POST | asset_group/reorder | 分组排序 |
| POST | asset_group/submit | 提交分组 |
| POST | assetline/init | 初始化 |
| POST | assetline/submit | 提交 |
| POST | assetline/reset | 重置 |

### 账本模块 (`book/`)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | book/list | 账本列表 |
| POST | book/submit | 提交 |
| POST | book/delete | 删除 |
| POST | book/deletecheck | 删除检查 |
| POST | book/copy | 复制 |
| POST | book/invitecode | 邀请码 |
| POST | book/addbycode | 通过码加入 |
| POST | book/infobycode | 通过码获取信息 |
| POST | book/members | 成员列表 |
| POST | book/quit | 退出 |
| POST | book/config | 配置 |
| POST | book/reorder | 排序 |
| POST | book/visible | 可见性 |
| POST | book/typelist | 类型列表 |

### 其他模块

| 模块 | 端点 | 说明 |
|------|------|------|
| budget | submitv2, list | 预算 |
| tag | syncall, submitgroup, submittag, deletegroup, deletetag, reordergroup, reordertag, archivetag, list | 标签 |
| currency | listv2 | 多币种 |
| card (usercard) | listv2, submit, delete, reorder | 信用卡 |
| installment | list, submit, delete, bindbill, prepay, config | 分期 |
| repeat_task | list, submit, delete, toggle, bindbill, resetcount, holidays | 周期记账 |
| saving | list, submit, delete, deposit, canceldeposit | 存钱计划 |
| import | platforms, packs, packbills, submitfile, confirmresult, deletepack | 导入 |
| user | update, updateconfig, getaddress, submitaddress | 用户 |

---

## 三、认证与响应格式

### 响应信封 (Response Envelope)

所有 API 响应采用统一结构（定义于 `h8/a.java` 的 `parseData` 基类）：

```json
{
  "ec": 200,
  "em": "",
  "data": {
    ...
  }
}
```

- `ec`: 错误码，`200` = 成功，其他值 = 错误（如 `400404` = 签名验证失败，`8888` = 业务错误，`9002` = 请求参数错误）
- `em`: 错误消息字符串（可能是纯文本或嵌套 JSON 如 `{"msg":"登录失败！账号和密码不匹配"}`）
- `data`: 实际业务数据

**代码验证**：`h8/a.java:53` 的 `isSuccess()` 方法检查 `this.f11905a == 200`。

### 登录流程

```
1. POST /account/login
   参数:
     v: 手机号/邮箱
     pwd: MD5(明文密码).toLowerCase()

   请求头: 见第五章"签名系统"

   响应 (成功):
     {
       "ec": 200,
       "em": "",
       "data": {
         "token": "xxx",        ← 存为 utoken header
         "user": {"id": "xxx"},
         "books": [...],
         "is_new": 0
       }
     }

2. 后续所有请求:
   Header: utoken: <token>
   Header: ctrl/act/os/... (所有设备头 + 签名头)

3. 刷新 token:
   POST /account/refreshtoken?uid=<uid>
```

---

## 四、密码加密算法 ★ 已破解

**调用链:** `n8.j.e(str)` → `n8.k.b(str)` → `MD5(str).toLowerCase()`

```java
// n8/j.java:242
public static String e(String str) {
    return k.b(str);   // → n8.k.b()
}

// n8/k.java:22
public static String b(String str) {
    MessageDigest messageDigest = MessageDigest.getInstance("MD5");
    messageDigest.update(str.getBytes());
    return a(messageDigest.digest()).toLowerCase();
}
```

**结论：密码使用纯 MD5 哈希，无盐、无迭代。**
Go 等价实现: `fmt.Sprintf("%x", md5.Sum([]byte(password)))`

---

## 五、签名系统 ★ 完全破解

> 这是整个逆向过程中最核心的部分。签名算法完全实现在 native library `libfabricsuffer.so` 中，
> 通过静态反汇编完整还原了计算逻辑。

### 5.1 Java 层调用入口

```java
// com/mutangtech/arc/http/JNIHelper.java
static { System.loadLibrary("fabricsuffer"); }

// 在 g8/a.java 中调用:
String reqId = JNIHelper.requestId(ctx, ctrl, act, versionCode);
req.setHeader("reqidv2", String.valueOf(reqId));
req.setHeader("tok", JNIHelper.a(reqId, ctrl, act));

// JNIHelper.a — 生成 tok（Java 层实现，调用 native encreqid）
public static String a(String reqId, String ctrl, String act) {
    return k.b(reqId + "1172020" + ctrl + encreqid(reqId) + act);
    //      ↑ k.b() 是 MD5
}

private static native String encreqid(String reqId);
public static native String requestId(Context ctx, String ctrl, String act, int versionCode);
```

### 5.2 Native Library 导出符号

```
$ nm -D libfabricsuffer.so

MD5Final           (0x1e04)
MD5Init            (0x10d8)
MD5Update          (0x10ec)
encode_request_id   (0x2678)  ← 对应 JNIHelper.encreqid()
md5_encode          (0x2130)
native_requestId    (0x26bc)  ← 对应 JNIHelper.requestId() 的 JNI 桥接
new_request_id      (0x24b4)  ← 实际生成 requestId 的逻辑
```

### 5.3 关键字符串常量

```
$ strings -t x libfabricsuffer.so

偏移     内容
----     ----
0x650    "0s3Sd"
0xa90    "free20170908&x_*1127"      ← requestId 的 MD5 盐
0xba1    "%s%s%s%s"                  ← join_chars 格式化串
0xbbd    "michaeljackson"            ← encreqid 的 MD5 盐
```

### 5.4 `new_request_id()` 反汇编分析

```
new_request_id (0x24b4):

1. clock_gettime()          → 获取当前纳秒精度时钟
2. 纳秒 / 常数              → 转换为毫秒值
3. get_app_package()        → 返回 "com.mutangtech.qianji"
4. 另一个 JNI 调用           → 获取 ctrl 字符串
5. malloc(128)              → 分配 128 字节缓冲区
6. join_chars(128, "%s%s%s%s", versionCode, millis, pkg, ctrl)
                            → 拼接: "1170<millis>com.mutangtech.qianjiaccount"
7. md5_encode(jniEnv, "free20170908&x_*1127", combined_str)
                            → MD5(combined + salt)
8. return MD5 result        → 返回 requestId（32 位十六进制）
```

### 5.5 `encode_request_id()` 反汇编分析

```
encode_request_id (0x2678):

1. GetStringUTFChars(env, requestId_jstring, NULL)
                            → 将 Java String 转为 C 字符串
2. md5_encode(jniEnv, "michaeljackson", c_str)
                            → MD5(requestId + "michaeljackson")
3. return MD5 result        → 返回 encreqid（32 位十六进制）
```

### 5.6 `md5_encode()` 反汇编分析

```
md5_encode(JNIEnv, salt, input):

1. len1 = strlen(input)
2. len2 = strlen(salt)
3. buffer = malloc(len1 + len2 + 1)
4. strcpy(buffer, input)    → 先拷贝 input
5. strcat(buffer, salt)     → 再追加 salt
6. MD5_Init → MD5_Update(buffer, strlen(buffer)) → MD5_Final
7. 将 16 字节 digest 转为 32 字符十六进制小写

结论: md5_encode(input, salt) = MD5(input + salt)
      salt 追加在 input 后面（非前置）
```

### 5.7 完整签名算法公式

```
给定:
  ctrl        = "account" | "bill" | "category" | ...
  act         = "login" | "syncall" | "listv2" | ...
  versionCode = 1170
  pkg         = "com.mutangtech.qianji"
  millis      = 当前时间毫秒数 (UnixMilli)

步骤:

1. combined = sprintf("%d%d%s%s", versionCode, millis, pkg, ctrl)
   例: "11701799371000000com.mutangtech.qianjiaccount"

2. requestId = MD5(combined + "free20170908&x_*1127")

3. encReqId  = MD5(requestId + "michaeljackson")

4. tok = MD5(requestId + "1172020" + ctrl + encReqId + act)

HTTP 头:
  reqidv2: <requestId>
  tok:     <tok>
```

### 5.8 签名验证测试

| 测试条件 | 结果 |
|---------|------|
| 仅 `ctrl` + `act` + `os` | `ec=400404, verify failed (and)` |
| 加完整设备头（devid, devbrand, ...） | `ec=400404, verify failed (r)` |
| 加完整设备头 + 正确签名的 reqidv2 + tok | `ec=8888, 登录失败！账号和密码不匹配` ✅ |
| **curl 测试命令** | |
| `curl -s ... -H "ctrl: account" -H "reqidv2: <computed>" -H "tok: <computed>" -d "v=13800138000&pwd=cc03e..."` | 正常返回业务错误（非签名错误） |

**结论: 签名算法已完全破解，服务端验证通过。**

---

## 六、同步机制 ★

钱迹采用 **双向同步**：PULL 拉取其他设备数据 + PUSH 推送本地变更。

总协调器: `je/c.java`

```java
// je.c.i() — 完整同步流程
new d(
    le.a.start(context, z10, callback),  // PULL (syncv2/pull)
    ne.a.start(context, true)             // PUSH (bill/syncall)
);
```

### 6.1 PULL: `syncv2/pull` (拉取其他设备账单)

**协调器**: `le/a.java`
**API**: `POST /syncv2/pull` (构建于 `pc/b.java`)
**响应模型**: `le/c.java` (changes + deletes + pageParams)
**保存**: `l.java: savePullResult()`

#### 请求

```
POST /syncv2/pull
Body:
  uid=<user_id>
  bookid=-1                         // -1=全部账本
  lasttimes=<json>                  // 可选，增量同步时间戳
  pageoffset=<int64>                // 分页偏移
  pagesign=<string>                 // 分页游标
```

#### 响应

```json
{
  "ec": 200,
  "data": {
    "changes": [{...bill...}],     // 服务端返回的完整账单列表
    "deletes": [123456789],        // 需删除的账单 ID
    "categories": [...],           // 更新的分类
    "bookid": -1,
    "pageoffset": 0,
    "pagesign": "",
    "hasmore": 0,
    "count": 187
  }
}
```

#### 分页循环

```
pageOff=0, pageSign=""
while hasmore:
    pull(bookid, lasttimes, pageOff, pageSign)
    savePullResult(changes, deletes)
    pageOff = response.pageoffset
    pageSign = response.pagesign
```

### 6.2 PUSH: `bill/syncall` (推送本地变更到服务端)

**协调器**: `ne/a.java`
**API**: `POST /bill/syncall` (构建于 `com/.../api/bill/d.java`)
**响应模型**: `le/d.java` → `BillSyncResult`
**保存**: `l.java: saveSyncedResult()`

#### 请求

```
POST /bill/syncall
Body:
  uid=<user_id>
  v={"bills":{"changelist":[...],"dellist":[...]}}
```

#### `v` 参数 JSON 结构

```json
{
  "bills": {
    "changelist": [
      {
        "id": 123456789,
        "assetid": 1001,
        "bookid": 1,
        "cateid": 5,
        "createtime": 1716307200,
        "descinfo": "",
        "extra": {...},
        "fromid": 0,
        "images": [],
        "money": 26.5,
        "platform": 0,
        "remark": "咖啡",
        "status": 2,
        "targetid": 0,
        "time": 1716307200,
        "type": 0,
        "updatetime": 1716307200,
        "userid": "...",
        "username": "..."
      }
    ],
    "dellist": [123456789, 987654321]
  }
}
```

#### 响应

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

**关键：** syncall 只返回 ID 数组（new_ids/update_ids 等），不返回账单正文。账单正文由 `syncv2/pull` 拉取。

#### 客户端处理 (saveSyncedResult)

```java
// l.java: saveSyncedResult(BillSyncResult)
new_ids → 找到本地 bill, status=1 (标记已同步)
update_ids → 同上
conf_ids → 找到本地 bill, 分配新 billid
del_ids → DELETE FROM user_bill
```

### 6.3 Bill 状态常量

| 值 | 含义 |
|----|------|
| 0 | 已删除 (`sync_is_delete()`) |
| 1 | 已同步 OK |
| 2 | 待同步 (`sync_need_sync()`) |

### 6.4 Bill 类型常量

| type | 含义 |
|------|------|
| 0 | 支出 SPEND |
| 1 | 收入 INCOME |
| 2 | 转账 TRANSFER |
| 3 | 信用卡还款 |
| 5 | 报销 BAOXIAO |
| 20 | 退款 REFUND |

### 6.5 Bill platform 常量

| platform | 含义 |
|----------|------|
| 0 | 手动记账 |
| 120 | 重复任务 |
| 121 | 分期 |
| 122 | 自动记账 |

---

## 七、完整 HTTP 请求头

### 所有请求必须携带

| Header | 值 | 说明 |
|--------|-----|------|
| `ctrl` | `account`/`bill`/... | 模块名 |
| `act` | `login`/`syncall`/... | 操作名 |
| `os` | `1` | Android=1 |
| `osvs` | `32` | SDK 版本 |
| `devbrand` | `XIAOMI` | 设备品牌大写 |
| `devname` | 设备型号 | e.g. `23013RK75C` |
| `devid` | UUID MD5 | 32 位 hex，设备唯一标识 |
| `vs` | `1170` | app versionCode |
| `pkg` | `com.mutangtech.qianji` | 包名 |
| `vsn` | `4.3.7` | app versionName |
| `timezoneoffset` | `8` | UTC 偏移小时数 |
| `clang` | `zh` | 语言 |
| `cregion` | `CN` | 地区 |
| `mk` | `none`/`google`/`huawei` | 应用市场 |
| `reqidv2` | `<32 hex>` | 请求 ID（签名步骤2） |
| `tok` | `<32 hex>` | 请求签名（签名步骤4） |
| `content-type` | `application/x-www-form-urlencoded` | 内容类型 |

### 登录后额外携带

| Header | 值 | 说明 |
|--------|-----|------|
| `utoken` | `<token>` | 认证令牌 |

### ⚠️ 大小写敏感

服务端对 Header 名称区分大小写。Go 的 `http.Header.Set()` 会将自定义 Header 名
转换为 Canonical 格式（如 `ctrl` → `Ctrl`），导致服务端拒绝连接。

**解决方案**: 使用直接 map 赋值 `req.Header["ctrl"] = []string{"account"}` ，
或使用 `curl` 作为 HTTP 后端（本项目当前采用 curl 后端，定义在 `curl.go`）。

---

## 八、URL Scheme（Android 本地接口）

基础格式: `qianji://publicapi/addbill?type=0&money=26.5&catename=咖啡&...`

| 参数 | 说明 | 示例 | 必填 |
|------|------|------|------|
| type | 0=支出 1=收入 2=转账 3=信用卡还款 5=报销 | type=0 | 是 |
| money | 金额 (>0, 精度2位) | money=26.5 | 是 |
| time | 时间 (yyyy-MM-dd HH:mm:ss) | time=2020-01-31 12:30:00 | 否 |
| remark | 备注 | remark=星巴克 | 否 |
| catename | 分类名 (支持 /::/ 分隔二级) | catename=三餐/::/午餐 | 否 |
| catechoose | =1 弹出分类选择面板 | catechoose=1 | 否 |
| bookname | 账本名 | bookname=日常账本 | 否 |
| accountname | 账户名 (转账必填) | accountname=微信 | 否* |
| accountname2 | 转入账户 (转账必填) | accountname2=招行信用卡 | 否* |
| fee | 手续费 | fee=1.5 | 否 |
| coupon | 优惠券 | coupon=2.5 | 否 |
| showresult | =0 不显示成功提示 | showresult=0 | 否 |

---

## 九、Token 管理

```
s8.b.getInstance().getUserToken()       → 获取 token
s8.b.getInstance().getLoginUserID()     → 获取用户 ID
s8.b.getInstance().onLogin(user, token) → 保存登录
s8.b.getInstance().logout()             → 登出

存储: x7.a (通过 x7/c.java 包装) → 持久化到 SharedPreferences
```

---

## 十、Go CLI 实现

### 项目结构

```
/Users/wepie/GoWorkplace/money/
  go.mod                  # module github.com/wepie/qianji
  sign.go                 # 签名计算（MD5 + native 算法）
  client.go               # Client/Session 类型定义
  curl.go                 # HTTP 请求（curl 后端，绕过 Go TLS 问题）
  login.go                # Login/Logout + 分类/账本/资产查询
  bill.go                 # Bill 模型 + SyncBills/AddBill/DeleteBill
  cmd/qianji/main.go      # CLI 入口
  qianji                  # 编译产物
  qianji_apk_analysis.md  # 本文档
```

### CLI 命令

```
qianji login              交互式登录（输入手机号+密码）
qianji add 26.5 咖啡      支出一笔 26.5 元"咖啡"
qianji add -i 1000 工资   收入一笔 1000 元"工资"
qianji add -c 5 35 午餐   指定分类 ID 为 5
qianji cats               列出所有分类
qianji books              列出所有账本
qianji assets             列出资产账户
qianji logout             登出
```

### 构建

```bash
cd /Users/wepie/GoWorkplace/money
go build -o qianji ./cmd/qianji/
```

---

## 十一、待修复项目

| 问题 | 说明 | 优先级 |
|------|------|--------|
| login 参数位置 | 当前 curl.go 把 login 参数放 URL query string，应放 POST body（`-d`） | 高 |
| 真机验证 | 用 Xiaomi 13 真机抓包对比，确认签名字段完全一致 | 中 |
| bill.go 签名集成 | bill.go 中的 SyncBills/AddBill 未集成签名头 | 中 |
| curl 依赖 | CLI 依赖系统 curl，可选改为纯 Go TLS 实现 | 低 |

---

## 十二、关键源码文件索引

| 功能 | 路径 |
|------|------|
| 同步总协调器 | `je/c.java` |
| PULL 协调器 | `le/a.java` |
| PUSH 协调器 | `ne/a.java` |
| PULL API 构建 | `pc/b.java` |
| API 域名 | `x8/d.java` |
| 配置读取 | `ac/a.java` |
| URL 构建器 | `gi/c.java` |
| API 基类 | `g8/c.java` |
| 请求构建（注入签名头） | `g8/a.java` |
| 响应基类（ec/em/data） | `h8/a.java` |
| 响应解析器基类 | `com/mutangtech/arc/http/parser/a.java` |
| 签名 Java 层 | `com/mutangtech/arc/http/JNIHelper.java` |
| 密码加密 | `n8/j.java`, `n8/k.java` |
| Token 存储 | `s8/b.java` |
| 账单 API | `com/mutangtech/qianji/network/api/bill/d.java` |
| 账户 API | `com/mutangtech/qianji/network/api/account/a.java` |
| 登录响应解析 | `com/mutangtech/qianji/network/api/account/c.java` |
| 分类 API | `com/mutangtech/qianji/network/api/category/` |
| 资产 API | `com/mutangtech/qianji/network/api/asset/a.java` |
| URL Scheme | `com/mutangtech/qianji/schema/AppSchemaAct.java` |
| Native 库 | `lib/arm64-v8a/libfabricsuffer.so` |
