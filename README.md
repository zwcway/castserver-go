# Castspeaker
Castspeaker 是一个应用于基于以太网/WIFI局域网，C/S 架构，在带宽允许内支持更多数量扬声器客户端的数字扬声器管理系统，使用时钟同步保证所有扬声器客户端高质量的同步播放。Castspeaker可以自动发现局域网内的扬声器客户端，自动匹配采样率和位宽。支持设备分组，可以为每个分组单独指定播放源，通过在移动设备或者电脑端播放到指定的设备分组。可应用在智能家居中全屋音响，支持使用 homeassitant 作为管理端。

![overview](https://raw.githubusercontent.com/zwcway/castserver-go/main/doc/web.png)

# 编译
## 依赖
- `golang` / `nodejs` / `make` / `pkg-config`
- `libavcodec` / `libavformat` / `libavutil` / `libswresample`
- linux 下本地播放使用 alsa， 还需依赖 `libasound2-dev`

## 编译
$ make

## 运行
castserver --help

## web 后台
http://localhost:4415
