# ngapost2md 项目说明

## 项目概述

ngapost2md 是一个用 Go 语言编写的工具，用于将 NGA 论坛的帖子转换为 Markdown 格式。它能够爬取帖子内容，包括回复人、时间、内容和图片，并将其保存为本地 Markdown 文件。该项目是 2023 年由 Go 语言重写的版本，旧版 Python 版本已不再维护。

主要功能包括：
- 爬取 NGA 论坛帖子内容
- 将内容转换为 Markdown 格式
- 保存帖子中的图片到本地
- 支持只下载特定用户 ID 的发言层
- 支持获取用户 IP 位置
- 支持本地表情图片资源
- 支持文件夹和文件名使用帖子标题
- 支持切分生成的 Markdown 文件

## 项目结构

```
ngapost2md/
├── go.mod          # Go 模块定义文件
├── go.sum          # Go 模块校验和文件
├── main.go         # 程序入口文件
├── config/         # 配置相关代码
│   └── config.go
├── nga/            # 核心功能代码
│   ├── nga.go      # 主要的帖子处理逻辑
│   └── utils.go    # 工具函数
├── assets/         # 资源文件
│   └── config.ini  # 默认配置文件模板
├── README.md       # 项目说明文档
├── LICENSE         # 许可证文件
└── ngapost2md      # 编译后的可执行文件（示例）
```

## 构建和运行

### 构建

要构建此项目，需要 Go 1.24 或更高版本。

```bash
# 克隆项目
git clone https://github.com/ludoux/ngapost2md.git
cd ngapost2md

# 下载依赖
go mod download

# 构建可执行文件
go build -o ngapost2md main.go
```

### 运行

在运行程序之前，需要先配置 `config.ini` 文件，确保其中的 `ngaPassportUid`、`ngaPassportCid` 和 `ua` 配置项已正确设置。

```bash
# Linux
./ngapost2md <tid>

# Windows
.\ngapost2md.exe <tid>
```

其中 `<tid>` 是 NGA 帖子的 ID。

### 参数说明

- `tid`: 待下载的帖子 ID
- `--authorid <aid>`: 只下载指定用户 ID 的发言层
- `-v, --version`: 显示版本信息并退出
- `-h, --help`: 显示帮助信息并退出
- `-u, --update`: 检查最新版本
- `--gen-config-file`: 生成默认配置文件 `config.ini` 并退出

### 配置文件

程序需要一个 `config.ini` 文件与可执行文件在同一目录下。配置文件包含网络设置、帖子处理选项等。首次运行前需要修改配置文件中的 `ua`、`ngaPassportUid` 和 `ngaPassportCid` 等项。

## 开发约定

- 代码使用 Go 语言编写，遵循 Go 语言的编码规范
- 配置文件使用 INI 格式
- 项目使用 Go Modules 进行依赖管理
- 代码结构清晰，功能模块分离（`config`、`nga` 等包）
- 使用 `go-flags` 库处理命令行参数
- 使用 `req/v3` 库进行 HTTP 请求
- 使用 `jsonparser` 库解析 JSON 数据
- 使用 `ants/v2` 库进行并发控制
- 使用 `ini.v1` 库处理 INI 配置文件
- 使用 `cast` 库进行类型转换
- 代码中包含详细的注释说明
- 遵循 Go 语言的错误处理模式
- 修改完后只需要测试编译即可，不需要运行测试
- 你只需修改*.go文件，不要修改或删除任何其他文件