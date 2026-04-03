# PostgreSQL Wire Protocol 完整筆記

> 適用對象：實作 PostgreSQL proxy / sidecar（如 go-pg-router）的開發者

## 0. Related link

- [Frontend/Backend Protocol](https://www.postgresql.org/docs/current/protocol.html)



---

## 1. 整體架構

PostgreSQL 使用自己的**應用層協議（Wire Protocol）**，架構在 TCP 之上。TCP 本身只是傳輸層，所有的格式、路由、認證邏輯都在應用層處理。

```
┌─────────────┐
│  Application │  (你的 app / driver)
├─────────────┤
│  Wire Proto  │  ← 這份文件的主題
├─────────────┤
│     TLS      │  (optional)
├─────────────┤
│     TCP      │
└─────────────┘
```

---

## 2. Protocol 版本

PostgreSQL Wire Protocol 歷史上版本很少，現在只需要關心一個：

| 版本 | 狀態 | 說明 |
|---|---|---|
| 1.0 | 已廢棄 | 極古老 |
| 2.0 | 已廢棄 | PostgreSQL 6.x / 7.x 時代 |
| **3.0** | ✅ **現行標準** | PostgreSQL 7.4（2003）引入，沿用至今（v17） |

### 2.1 Startup Message 握手

連線建立後，Client 送出的第一個訊息是 **Startup Message**，格式如下：

```
┌─────────────────────────────┐
│ Length (4 bytes)            │
│ Protocol Version (4 bytes)  │  ← 高 16 位 = major, 低 16 位 = minor
│   → 0x00030000 = 3.0        │
│ Parameters (key=value pairs)│
│   user=postgres\0           │
│   database=mydb\0           │
│   \0                        │  ← 結尾雙 null
└─────────────────────────────┘
```

### 2.2 特殊 Magic Number

版本欄位也被借用來傳送特殊請求（不是真正的版本號）：

| Magic Number | 十進制 | 用途 |
|---|---|---|
| `0x04D2162E` | 80877102 | **Cancel Request** — 取消正在執行的 query |
| `0x04D2162F` | 80877103 | **SSLRequest** — 要求升級 TLS |
| `0x04D21630` | 80877104 | **GSSENCRequest** — 要求 GSSAPI 加密 |

### 2.3 Proxy 需要處理的 Startup 邏輯

```go
switch protocolVersion {
case 196608:   // 0x00030000 = 3.0 → 正常連線
    handleNormalStartup()
case 80877103: // SSLRequest → 回應是否支援 SSL
    handleSSLRequest()
case 80877102: // CancelRequest → 轉發給對應 backend
    handleCancelRequest()
}
```

---

## 3. 訊息格式

- [Message Formats](https://www.postgresql.org/docs/current/protocol-message-formats.html)

Startup Message 之後，所有訊息（含 Client 和 Server）都使用統一格式：

```
┌──────────┬────────────┬─────────────┐
│ Type (1B)│ Length (4B)│ Payload ... │
└──────────┴────────────┴─────────────┘
```

- **Type**：1 byte ASCII 字元，識別訊息類型
- **Length**：4 bytes big-endian integer，**包含自身 4 bytes**，不含 Type byte
- **Payload**：內容依訊息類型而定

---

## 4. Text 模式 vs Binary 模式

資料傳輸格式有兩種，這是**應用層的選擇**，不是 TCP 層的概念。

| | Text 模式 | Binary 模式 |
|---|---|---|
| **格式** | 所有值轉成字串（人類可讀） | 值以原始二進制表示 |
| **整數 `123`** | `"123"`（3 bytes） | `0x0000007B`（4 bytes） |
| **浮點 `3.14`** | `"3.14"`（4 bytes） | IEEE 754（8 bytes） |
| **日期** | `"2026-03-21"` | Julian day number |
| **效能** | 需要序列化 / 反序列化 | 直接記憶體對應，較快 |
| **可讀性** | 高 | 低 |
| **預設** | Simple Query 預設 | 需明確指定 |

Format code：`0` = text，`1` = binary

---

## 5. Simple Query Protocol

### 5.1 流程

```
Client → Server:  Q (Query)
Server → Client:  T (RowDescription)
                  D (DataRow) × N
                  C (CommandComplete)
                  Z (ReadyForQuery)
```

### 5.2 Q 訊息結構

```
Q | Length (4B) | SQL string\0
```

Payload 就是純 SQL 字串（null-terminated），永遠使用 **text 模式**。

### 5.3 對 Proxy 的影響

- 讀第一個 byte 是 `Q` → Simple Query
- Payload 是純字串 → 直接做字串搜尋判斷 `SELECT` / write
- 一次 roundtrip，無需追蹤狀態

---

## 6. Extended Query Protocol

### 6.1 流程

```
Client → Server:
  P (Parse)    → 送 SQL template（含 $1, $2 佔位符）
  B (Bind)     → 送實際參數值 + format codes
  E (Execute)  → 執行
  S (Sync)     → 結束這個 cycle，觸發 server 回傳結果

Server → Client:
  1 (ParseComplete)
  2 (BindComplete)
  T (RowDescription)
  D (DataRow) × N
  C (CommandComplete)
  Z (ReadyForQuery)
```

### 6.2 Parse 訊息結構（P）

```
P
├── Length (4B)
├── Statement name (null-terminated)  ← "" 代表 unnamed prepared statement
├── SQL query (null-terminated)       ← 這裡才是 SQL template
└── Parameter type OIDs:
    ├── Count (2B)
    └── OID (4B) × count
```

**不能用字串直接搜尋整個 payload**，因為 statement name 和 OID bytes 可能干擾結果，必須正確 parse 二進制結構。

### 6.3 Bind 訊息結構（B）

```
B
├── Length (4B)
├── Portal name (null-terminated)
├── Statement name (null-terminated)
├── Parameter format codes:
│   ├── Count (2B)
│   └── Format code (2B) × count   ← 0=text, 1=binary (input 參數)
├── Parameters:
│   ├── Count (2B)
│   └── Per parameter:
│       ├── Value length (4B)       ← -1 代表 NULL
│       └── Value bytes             ← 格式依上面的 format code
└── Result format codes:
    ├── Count (2B)
    └── Format code (2B) × count   ← 0=text, 1=binary (output 欄位)
```

> **重要**：Bind 訊息裡有**兩組** format code：
> - 第一組：參數的輸入格式（client 送給 server）
> - 第二組：結果的輸出格式（server 回傳給 client）

### 6.4 Format Code 的特殊規則

| Count | 意義 |
|---|---|
| `0` | 所有參數 / 欄位都用 text |
| `1` | 用同一個 format code 套用到所有參數 / 欄位 |
| `N` | 每個參數 / 欄位各自指定 |

---

## 7. Client → Server 訊息類型總覽

| Byte | ASCII | 名稱 | 說明 |
|---|---|---|---|
| `0x51` | `Q` | Query | Simple Query |
| `0x50` | `P` | Parse | Extended：準備 prepared statement |
| `0x42` | `B` | Bind | Extended：綁定參數 |
| `0x45` | `E` | Execute | Extended：執行 |
| `0x44` | `D` | Describe | 詢問回傳欄位結構 |
| `0x53` | `S` | Sync | 結束一個 Extended cycle |
| `0x58` | `X` | Terminate | 關閉連線 |
| `0x43` | `C` | Close | 關閉 prepared statement 或 portal |
| `0x48` | `H` | Flush | 要求 server 立即送出緩衝中的回應 |

---

## 8. Server → Client 訊息類型總覽

| Byte | ASCII | 名稱 | 說明 |
|---|---|---|---|
| `0x54` | `T` | RowDescription | 欄位定義（名稱、OID、format code） |
| `0x44` | `D` | DataRow | 一筆資料列 |
| `0x43` | `C` | CommandComplete | 執行完成（如 `SELECT 3`、`INSERT 0 1`） |
| `0x5A` | `Z` | ReadyForQuery | Server 準備好接下一個 query |
| `0x45` | `E` | ErrorResponse | 錯誤訊息（含 severity、code、message 等欄位） |
| `0x4E` | `N` | NoticeResponse | 警告 / 通知（非致命） |
| `0x31` | `1` | ParseComplete | Parse 成功 |
| `0x32` | `2` | BindComplete | Bind 成功 |
| `0x33` | `3` | CloseComplete | Close 成功 |
| `0x41` | `A` | NotificationResponse | LISTEN/NOTIFY 通知 |

---

## 9. Proxy 實作策略（以 go-pg-router 為例）

### 9.1 路由決策

```
讀第一個 byte
    │
    ├── Q → Simple Query
    │       讀 SQL 字串 → 判斷讀 / 寫 → 路由
    │
    └── P → Extended Query
            解析 Parse 訊息 → 拿 SQL template → 判斷讀 / 寫 → 路由
            之後的 B / E / S → 直接 forward 給已決定的 backend
```

### 9.2 Extended Query 狀態機

```go
type SessionState struct {
    targetBackend *Backend  // 已決定的 backend（primary 或 replica）
    phase         Phase     // Idle / InExtendedQuery
}

// P → 決定 backend，進入 InExtendedQuery 狀態
// B / E / S → forward，維持狀態
// Z (ReadyForQuery) → 重置狀態回 Idle
```

### 9.3 回程資料傳輸

回程（Server → Client）最簡單的做法是**純 forward（io.Copy）**：

```go
// 注意：proxy 已讀過的 bytes 要先補送給 backend
backendConn.Write(alreadyReadBytes)

// 雙向 pipe
go io.Copy(backendConn, clientConn)  // client → backend
go io.Copy(clientConn, backendConn)  // backend → client（回程）
```

### 9.4 回程需不需要解析？

| 需求 | 做法 |
|---|---|
| 只做 routing | 純 `io.Copy`，完全不解析 |
| Logging | 攔截 `C`（CommandComplete） |
| Error handling / retry | 攔截 `E`（ErrorResponse） |
| Cache | 需完整解析 `T` + `D` |

**最小可行**：攔截 `ErrorResponse`，判斷是否因 replica lag 導致，決定是否 retry。其他全部 `io.Copy`。

### 9.5 Binary Payload 的注意事項

如果需要做 query rewriting 或讀取參數值：

- **不能**對整個 Bind payload 做字串操作
- 必須先 parse 出 format code，再依格式解析參數值
- Binary 整數是 big-endian，浮點是 IEEE 754

---

## 10. 整體資料流

```
Client
  │  Q / P / B / E / S
  ▼
Proxy（讀第一個訊息，決定路由）
  │  forward（含已讀 bytes）
  ▼
Backend（primary or replica）
  │  T / D / C / Z
  ▼
Proxy（io.Copy，不解析）
  │
  ▼
Client
```

---

## 11. 參考資料

- [PostgreSQL Protocol Documentation](https://www.postgresql.org/docs/current/protocol.html)
- [PostgreSQL Message Formats](https://www.postgresql.org/docs/current/protocol-message-formats.html)
- [PostgreSQL Error Codes](https://www.postgresql.org/docs/current/errcodes-appendix.html)
