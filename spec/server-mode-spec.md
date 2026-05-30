# ngapost2md Server Mode Spec

## 1. Overview

为 ngapost2md 添加 Server 运行模式，以 HTTP API 暴露核心下载/更新功能，并提供可选的 Web 前端界面。前端通过 HTTP API 与后端交互，实时下载进度通过 WebSocket 推送。

## 2. 启动方式

新增子命令 `serve`：

```
ngapost2md serve [OPTIONS]
```

| Option | Short | Default | Description |
|---|---|---|---|
| `--host` | | `0.0.0.0` | 绑定 IP 地址 |
| `--port` | `-p` | `8080` | 监听端口 |
| `--password` | | 从 config.ini `[server]` section 读取 | Basic Auth 密码 |
| `--no-ui` | | `false` | 禁用前端，仅提供 API |

- 若 `--password` 未指定，从 `config.ini` 的 `[server]` section 中 `password` 读取
- 若两者均无密码，Server 模式拒绝启动并提示设置密码
- CLI 模式（原有 positional arg 方式）不受影响，`serve` 和原 CLI 为两条互斥路径

## 3. 认证

### 3.1 Session Cookie（Web 前端）

Web 前端通过登录页面进行认证，服务端维持 session：

- 访问受保护页面时，若未认证，自动重定向到 `/login.html` 登录页
- 用户提交用户名（`admin`）和密码 → `POST /api/login` 验证 → 服务端创建 session → 通过 `Set-Cookie` 写入 `ngapost2md_session` cookie
- Session 有效期为 **72 小时**，服务端定期清理过期 session
- Cookie 属性：`HttpOnly`、`SameSite=Lax`、`Path=/`
- 登出通过 `POST /api/logout`，服务端删除 session 并清除 cookie
- 前端页面通过 `/api/login` 状态码检测登录有效性，未认证时自动跳转登录页

### 3.2 Basic Auth（API 客户端兼容）

同时保留 Basic Auth 作为 API 客户端的认证方式：

- 用户名固定为 `admin`，密码由 `--password` 或 config.ini `[server].password` 决定
- 认证优先级：session cookie > Basic Auth
- 未认证 API 请求返回 401，未认证页面请求重定向到登录页

### 3.3 公开路由

以下路由不要求认证：
- `POST /api/login`
- `POST /api/logout`
- `/login.html`（登录页面）

### 3.4 WebSocket 认证

WebSocket 连接支持两种认证方式：
- **Session cookie**：浏览器客户端自动携带 cookie（推荐）
- **URL token**：`?token=base64(admin:password)`（兼容旧客户端）

## 4. HTTP API

### 4.0 认证

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/login` | 登录，body: `{"username": "admin", "password": "xxx"}`，成功返回 200 并设置 session cookie |
| `POST` | `/api/logout` | 登出，清除 session cookie |

### 4.1 下载/更新任务

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/download` | 开始下载新帖子，body: `{"tid": 123456, "authorId": 0}` |
| `POST` | `/api/update` | 增量更新已有帖子，body: `{"tid": 123456}` |
| `GET` | `/api/tasks` | 获取当前任务队列（第一项为正在执行的任务，后续为排队任务） |
| `DELETE` | `/api/tasks/{tid}` | 取消队列中等待的任务（正在执行中的任务无法取消） |

#### 任务状态模型

```json
{
  "tid": 123456,
  "type": "download",      // "download" | "update"
  "status": "downloading", // "pending" | "downloading" | "processing" | "generating_markdown" | "completed" | "failed" | "cancelled"
  "currentPage": 5,
  "totalPage": 10,
  "currentFloor": 120,
  "totalFloor": 200,
  "stage": "downloading",  // "downloading" | "processing_content" | "downloading_assets" | "generating_markdown"
  "error": "",
  "startTime": "2026-05-17T10:00:00Z",
  "endTime": ""
}
```

### 4.2 已下载帖子

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/posts` | 获取所有已下载帖子列表 |
| `GET` | `/api/posts/{tid}/download` | 将帖子文件夹打包为 zip 下载（排除 `.` 开头的文件） |

扫描工作目录下所有符合 `{tid}` 或 `{tid}-{title}` 或 `{tid}({authorId})` 格式的子目录，读取其 `process.ini` 获取元数据。

#### 下载 zip

- 响应 Content-Type 为 `application/zip`
- 文件名格式：`{folderName}.zip`
- 排除所有以 `.` 开头的文件和目录（如 `.DS_Store`、`.gitkeep` 等）

```json
[
  {
    "tid": 123456,
    "authorId": 0,
    "title": "帖子标题",
    "folderName": "123456-帖子标题",
    "maxPage": 10,
    "maxFloor": 200,
    "floorCount": 180,
    "hasMarkdown": true
  }
]
```

### 4.3 定时任务

定时任务存储在独立文件 `schedules.json`（与 config.ini 同目录）。

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/schedules` | 获取所有定时任务 |
| `POST` | `/api/schedules` | 创建定时任务 |
| `PUT` | `/api/schedules/{id}` | 更新定时任务 |
| `DELETE` | `/api/schedules/{id}` | 删除定时任务 |

#### schedules.json 格式

```json
{
  "tasks": [
    {
      "id": "1",
      "tid": 123456,
      "authorId": 0,
      "cron": "0 9 * * 1-5",
      "enabled": true,
      "lastRunTime": "2026-05-17T09:00:00+08:00",
      "nextRunTime": "2026-05-18T09:00:00+08:00",
      "lastRunStatus": "completed",
      "createdTime": "2026-05-15T10:00:00+08:00"
    }
  ]
}
```

- `id` 为自增整数字符串，由 Server 自动分配
- `cron` 采用标准 5 字段 cron 表达式（分 时 日 月 周），时区为系统本地时区
- `lastRunStatus` 为 `"completed"` | `"failed"` | `""`
- `nextRunTime` 由 Server 根据 cron 表达式自动计算

### 4.4 配置

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/config` | 获取当前 config.ini 全部内容（返回 key-value 结构） |
| `PUT` | `/api/config` | 更新 config.ini（接收部分 key-value，合并写入） |

- Server 模式与 CLI 模式共享同一个 `config.ini`
- GET 返回所有 section 及 key-value（不含 `[server]` 中的 `password`），同时返回 `_comments` 字段包含每个配置项的注释说明
- PUT 仅更新传入的 key，未传入的保持原值
- 更新后立即生效（重新加载配置），但对于正在运行的任务，使用任务启动时的配置
- `[server]` section 中的 `password` 不可通过 API 修改（安全考虑），只能通过命令行 `--password` 或手动编辑 config.ini

#### GET /api/config 响应格式

```json
{
  "network": {
    "base_url": "https://bbs.nga.cn",
    "ua": "Mozilla/5.0 ...",
    "thread": "1"
  },
  "post": {
    "get_ip_location": "False"
  },
  "_comments": {
    "network": {
      "base_url": "软件访问的 NGA 域名。默认值为 https://bbs.nga.cn。",
      "ua": "浏览器 User-Agent，填写你常用浏览器的 UA 即可。",
      "thread": "网络并发数，理论上提高并发数可以增加下载速度。"
    },
    "post": {
      "get_ip_location": "是否查询用户基于 IP 的地理位置？若启用则会导致网络请求量最高增加 20 倍。"
    }
  }
}
```

- `_comments` 仅在配置项有注释时出现，注释文本来源于 `config.ini` 中以 `;` 开头的注释行

### 4.5 WebSocket

| Path | Description |
|---|---|
| `/ws` | WebSocket 连接，推送任务进度更新 |

WebSocket 认证方式（优先级从高到低）：
- Session cookie：浏览器客户端自动携带
- URL query 参数：`?token=base64(admin:password)`

#### Server 推送消息格式

```json
{
  "type": "progress",
  "tid": 123456,
  "taskType": "download",
  "status": "downloading",
  "currentPage": 5,
  "totalPage": 10,
  "currentFloor": 120,
  "totalFloor": 200,
  "stage": "downloading"
}
```

```json
{
  "type": "task_complete",
  "tid": 123456,
  "taskType": "download",
  "status": "completed",
  "message": "下载完成，共 10 页 200 楼"
}
```

```json
{
  "type": "task_failed",
  "tid": 123456,
  "taskType": "download",
  "status": "failed",
  "error": "网络请求失败: timeout"
}
```

## 5. 前端页面

前端为纯 HTML/CSS/JS，样式使用 Pico CSS（通过 CDN 或嵌入），文件通过 `go:embed` 嵌入 Go 二进制。

**架构原则**：前端为薄客户端（thin client），队列相关处理完全由后端 `TaskManager` 负责。前端仅通过 HTTP API 与后端交互，不在客户端维护队列状态：

- 新建任务：`POST /api/download` 或 `POST /api/update`
- 删除队列内任务：`DELETE /api/tasks/{tid}`
- 获取当前任务队列：`GET /api/tasks`（后端返回的数组第一项为当前执行中的任务，后续为排队任务）
- 实时进度通过 WebSocket 推送（`/ws`），前端收到消息后刷新 API 数据并更新进度展示

### 5.1 页面一：下载帖子

- 输入框：tid 或 NGA URL
- 可选输入框：authorId
- 按钮：开始下载，调用 `POST /api/download` 将任务加入后端队列
- 任务队列表格：
  - 列：tid、类型、状态、入队时间、操作
  - 执行中的任务行高亮显示，排队中的任务显示"移除"按钮
  - 数据来源于 `GET /api/tasks`，由后端 TaskManager 返回权威队列状态
- 下载启动后，显示实时进度区域：
  - 阶段指示器：下载 → 内容处理 → 生成 Markdown
  - 页码进度条：当前页/总页
  - 楼层进度：当前楼/总楼
  - 状态文字：downloading / processing / completed / failed
- 进度通过 WebSocket 实时推送，前端收到 WebSocket 消息后刷新队列数据并更新进度展示

### 5.2 页面二：帖子列表

- 表格展示所有已下载帖子：
  - 列：tid、标题、楼层数、最大页数、文件夹名
- 每行有"增量更新"按钮，点击后跳转到下载页面的进度展示模式（或就地展示进度）
- 每行有"下载"按钮，点击后触发 `GET /api/posts/{tid}/download` 下载 zip 包
- 支持刷新列表

### 5.3 页面三：定时任务

- 表格展示所有定时任务：
  - 列：tid、帖子标题、cron 表达式、下次执行时间、上次执行时间、上次执行状态、启用状态
- 新增任务：选择已下载的帖子（下拉列表）+ 输入 cron 表达式
- 编辑任务：修改 cron 表达式或启用/禁用
- 删除任务：确认后删除
- cron 表达式输入旁边提供常用模板快捷按钮（每天9点、每周一9点、每小时等）

### 5.4 页面四：配置

- 分组展示 config.ini 的所有 section 和 key-value
- 每个 key 旁边显示 `?` 图标，鼠标悬停展示对应的注释说明（tooltip）
- 每个 key 有对应的输入框，类型根据值推断（bool → checkbox，数字 → number input，字符串 → text input）
- 保存按钮：提交修改
- `[server]` section 的 `password` 不显示（安全考虑）
- 修改后提示"配置已保存并生效"

### 5.5 导航

顶部导航栏，四个页面通过 tab 或链接切换。单页应用风格（各页面独立 HTML 文件，通过导航跳转）。导航栏包含"登出"链接，点击后调用 `/api/logout` 清除 session 并跳转登录页。

### 5.6 登录页

独立的登录页面 `/login.html`：
- 用户名（默认 `admin`）和密码输入框
- 提交后 `POST /api/login`，成功后跳转首页
- 登录失败显示错误提示
- 未认证用户访问任何受保护页面时自动重定向到此页

## 6. 技术架构

### 6.1 Go 侧变更

1. **main.go**: 添加 `serve` 子命令分支，解析 serve 相关参数
2. **server/server.go**: 新文件，HTTP 服务器、路由、认证中间件（session cookie + Basic Auth）
3. **server/session.go**: 新文件，session 管理（创建、验证、过期清理，72h TTL）
4. **server/api.go**: 新文件，API handler 实现（含 login/logout）
5. **server/ws.go**: 新文件，WebSocket handler 及进度推送
6. **server/task.go**: 新文件，任务队列管理（FIFO，单任务执行）
7. **server/schedule.go**: 新文件，定时任务管理（CRUD + goroutine cron 定时器）
8. **server/frontend.go**: 新文件，静态文件 serve（go:embed）
9. **server/frontend/**: 新目录，存放前端 HTML/CSS/JS 文件（含 login.html）

### 6.2 新增依赖

| Dependency | Purpose |
|---|---|
| `github.com/gorilla/websocket` | WebSocket 支持 |
| `github.com/robfig/cron/v3` | Cron 表达式解析与定时执行 |

### 6.3 进度回调机制

```go
type ProgressCallback func(stage string, currentPage int, totalPage int, currentFloor int, totalFloor int)

// Tiezi.Download 增加 callback 参数
func (t *Tiezi) Download(callback ProgressCallback) error
```

- CLI 模式调用时传入 nil（不影响原有行为）
- Server 模式调用时传入回调，回调将进度信息推送到 WebSocket 连接管理器

### 6.4 定时任务执行器

- 使用 `robfig/cron/v3` 库解析 cron 表达式
- Server 启动时加载 `schedules.json`，为所有 `enabled=true` 的任务添加 cron job
- cron job 触发时，调用 `Tiezi.Download()`（增量更新）
- 任务执行结果更新到 `schedules.json`（lastRunTime、lastRunStatus）
- 新增/修改/删除/启禁 任务时，动态更新 cron scheduler

### 6.5 前端嵌入

```go
//go:embed frontend/*
var frontendFS embed.FS
```

- 当 `--no-ui` 未指定时，所有非 `/api` 和 `/ws` 路径 serve 前端静态文件
- 默认路由 `/` → `frontend/index.html`（下载页面）

## 7. config.ini 新增 section

```ini
[server]
host = 0.0.0.0
password = your_password_here
port = 8080
```

- `host`: 绑定 IP 地址，默认 `0.0.0.0`（所有网络接口），可被 `--host` 覆盖
- `password`: Basic Auth 密码，通过 API 不可修改
- `port`: 默认端口，可被 `--port` 覆盖

## 8. 错误处理

- 同一 tid 不能同时存在于队列中或正在执行（`POST /api/download`、`POST /api/update` 返回 409 Conflict）
- 正在执行中的任务不可取消（`DELETE /api/tasks/{tid}` 返回 404）
- tid 格式不合法（返回 400 Bad Request）
- 定时任务 cron 表达式不合法（返回 400 Bad Request）
- Server 启动时无密码（拒绝启动，日志提示配置密码）

## 9. 与 CLI 模式的兼容

- Server 模式和 CLI 模式共享 config.ini 和工作目录
- Server 正在下载某个 tid 时，CLI 不应同时操作该 tid（但无法强制阻止，仅做提示）
- CLI 模式代码不做任何改动，仅新增 serve 路径