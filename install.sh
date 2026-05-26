#!/bin/bash
set -e

REPO="hustuhao/qianji-cli"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"

detect_platform() {
    case "$(uname -s)" in
        Linux*)  echo "linux";;
        Darwin*) echo "darwin";;
        MINGW*|MSYS*|CYGWIN*) echo "windows";;
        *) echo "linux";;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64";;
        aarch64|arm64) echo "arm64";;
        *) echo "amd64";;
    esac
}

echo "检测平台..."
PLATFORM=$(detect_platform)
ARCH=$(detect_arch)
FILENAME="qianji_${PLATFORM}_${ARCH}.tar.gz"
ASSET_URL="https://github.com/$REPO/releases/latest/download/$FILENAME"

echo "平台: $PLATFORM $ARCH"
echo "下载: $ASSET_URL"

TMPDIR=$(mktemp -d)
cd "$TMPDIR"

echo "下载中..."
curl -fsSL "$ASSET_URL" -o "${FILENAME}"

echo "解压中..."
tar -xzf "$FILENAME"
chmod +x qianji_${PLATFORM}_${ARCH}

echo "安装到 $BIN_DIR/qianji ..."
mkdir -p "$BIN_DIR"
mv qianji_${PLATFORM}_${ARCH} "$BIN_DIR/qianji"

cd /
rm -rf "$TMPDIR"

echo "✅ 安装完成！运行: qianji --help"