# AGENTS

## 项目概览
- `castserver-go` 是一个基于 Go 的网络音频广播与扬声器管理服务。
- 核心功能包括设备发现、音频解码、推送、接收、控制及 Web 后台管理。
- 程序入口为 `main.go`，模块注册与启动由 `module.go` 定义。

## 构建与运行
- 推荐命令：`make`
- `make` 执行：`go fmt .`、`go vet .`、生成主程序，并构建前端静态资源。
- 也可以使用：`go build -o castserver .`
- 运行检查：`castserver --help`
- 默认 Web 管理地址：`http://localhost:4415`

## 主要目录
- `common/`：通用配置、日志、数据库、本地音频、工具函数等。
- `common/audio`: 音频流结构定义
- `common/bus`: 消息总线
- `common/config`: 配置文件结构定义与存储
- `common/database`: 数据库支持
- `common/dsp`: dsp基础库
- `common/element`: 音频工作流基本单位的定义
- `common/i18n`: 多国语言支持
- `common/jsonpack`: 流式json压缩格式支持
- `common/log` : 日志
- `common/pipeline`: 音频工作流结构定义
- `common/playlist`: 音频播放列表数据结构定义
- `common/protocol`: 音频网络传输协议数据结构定义
- `common/sounds`: 内嵌音效文件
- `common/speaker`: 扬声器设备数据结构
- `common/stream`: 音频流数据结构
- `common/utils`: 基础工具包
- `decoder/`：音频解码与 FFmpeg 集成，以及本地播放与文件流支持。
- `decoder/ffmpeg`: cgo与ffmpeg的中间层
- `decoder/localspeaker`: 通过本地声卡播放声音
- `detector/`：网络/设备发现与探测逻辑。
- `player/`: 音频播放器
- `pusher/`：音频数据推送、队列与播放控制。
- `control/`：控制消息、模块管理、音量与样本。
- `receiver/`：接收端协议实现，包括 DLNA/AirPlay。
- `web/`：HTTP API、WebSocket、Web 前端入口。
- `mutexer/`：服务端唯一实例检测。

## 代码结构与约定
- `module.go` 中定义 `Module` 接口：`Init(ctx)`、`Start()`、`DeInit()`。
- 程序通过 `initModules`/`startModules`/`deinitModules` 顺序管理模块生命周期。
- `main.go` 负责解析命令行参数、初始化日志、读取配置、创建带信号处理的 rootCtx。
- 配置与版本信息通过 `common/config` 提供。

## 特殊依赖与环境要求
- 本项目依赖 native 库：`libavcodec`、`libavformat`、`libavutil`、`libswresample`。
- 本地播放使用 ALSA，需要 `libasound2-dev`。
- 需要 `pkg-config` 和 `CGO_ENABLED=1` 才能正确编译。
- 交叉编译目标 ARM64/ARM 的过程参见 `README.md`。

## 前端
- 前端目录：`web/front`
- 构建命令：`cd web/front && npm i && npm run build`
- `make` 会触发前端构建。

## 文档链接
- 项目说明与编译指南：[`README.md`](README.md)
- 前端说明：[`web/front/README.MD`](web/front/README.MD)

## 使用建议
- 先阅读 `main.go` 和 `module.go`，理解模块初始化与运行流程。
- 修改代码后运行 `go fmt .` 并确保 `go vet .` 通过。
- 遇到音频或平台问题时，优先检查 `common/config`、`decoder/ffmpeg` 和 `decoder/localspeaker` 相关代码。

## 说明
- 当前仓库未包含 `.github/copilot-instructions.md`，已创建本文件帮助 AI 代理快速理解项目结构。
