# Castspeaker
Castspeaker 是一个应用于基于以太网/WIFI局域网，C/S 架构，在带宽允许内支持更多数量扬声器客户端的数字扬声器管理系统，使用时钟同步保证所有扬声器客户端高质量的同步播放。Castspeaker可以自动发现局域网内的扬声器客户端，自动匹配采样率和位宽。支持设备分组，可以为每个分组单独指定播放源，通过在移动设备或者电脑端播放到指定的设备分组。可应用在智能家居中全屋音响，支持使用 homeassitant 作为管理端。

![overview](https://raw.githubusercontent.com/zwcway/castserver-go/main/doc/web.png)

# 编译
## 依赖
- `golang` / `nodejs` / `make` / `pkg-config`
- `libavcodec` / `libavformat` / `libavutil` / `libswresample`
- linux 下本地播放使用 alsa， 还需依赖 `libasound2-dev`

## 编译
```bash
make
```

## 运行
```bash
castserver --help
```

# web 后台
http://localhost:4415

# 交叉编译（香橙派）
## 安装工具链 gcc-arm-9.2
- 香橙派 linux 系统的库版本 `GLIBC_2.30`
- 下载并解压 [gcc](https://mirrors.tuna.tsinghua.edu.cn/armbian-releases/_toolchain/gcc-arm-9.2-2019.12-x86_64-aarch64-none-linux-gnu.tar.xz)
```bash
ln -s `realpath gcc-arm-9.2-2019.12-x86_64-aarch64-none-linux-gnu` /aarch64
export PATH=/aarch64/bin:$PATH
export PKG_CONFIG_LIBDIR=/aarch64/lib/pkgconfig 
export CC=aarch64-none-linux-gnu-gcc
```

## 编译 ffmpeg-4.4.4 (libavcodec libavformat libswresample libavutil)
```bash
git clone --depth=1 https://git.ffmpeg.org/ffmpeg.git -b n4.4.4
cd ffmpeg
./configure --prefix=/aarch64 --arch=arm64 --enable-cross-compile --target-os=linux --cross-prefix=aarch64-none-linux-gnu- --disable-all --enable-gpl --enable-shared --enable-network --enable-autodetect --enable-avcodec --enable-avformat  --enable-avutil  --enable-swresample --enable-asm --enable-decoder=*
make -j12
make install

# 验证是否安装成功
pkg-config --exists libavcodec libavformat libswresample libavutil && echo true || echo false
```

## 编译 alsa-1.2.4
- 下载并解压后进目录 [alsa](https://codeload.github.com/alsa-project/alsa-lib/zip/refs/tags/v1.2.4)
```bash
./gitcompile --prefix=/aarch64 --host=aarch64-none-linux-gnu
make install

# 验证是否安装成功
pkg-config --exists alsa && echo true || echo false
```

## 编译 castserver-go
```bash
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 GOARM=7 go build -ldflags '-s -w' -o castserver ./
```

# 安装进香橙派并运行
```bash
scp castserver orangepi:/bin/
ssh orangepi castserver -i wlan0
```
