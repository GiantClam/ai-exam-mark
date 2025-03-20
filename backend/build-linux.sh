#!/bin/bash

# 设置错误时退出
set -e

# 默认参数
BUILD_MODE="dev"
OUTPUT_NAME="ai_exam_mark"
BUILD_TAGS=""
RUN_TESTS=false

# 显示帮助信息
show_help() {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  -m, --mode     Build mode (dev/prod) [default: dev]"
    echo "  -o, --output   Output binary name [default: ai_exam_mark]"
    echo "  -t, --test     Run tests before building"
    echo "  -h, --help     Show this help message"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--mode)
            BUILD_MODE="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_NAME="$2"
            shift 2
            ;;
        -t|--test)
            RUN_TESTS=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# 验证构建模式
if [[ "$BUILD_MODE" != "dev" && "$BUILD_MODE" != "prod" ]]; then
    echo "Error: Build mode must be either 'dev' or 'prod'"
    exit 1
fi

# 设置环境变量
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# 根据构建模式设置编译参数
if [[ "$BUILD_MODE" == "prod" ]]; then
    BUILD_TAGS="-tags=prod"
    echo "Building for production..."
    # 生产环境优化
    BUILD_FLAGS="-ldflags=-s -ldflags=-w -trimpath"
else
    echo "Building for development..."
    # 开发环境包含调试信息
    BUILD_FLAGS=""
fi

# 清理旧的构建文件
if [ -f "$OUTPUT_NAME" ]; then
    echo "Cleaning old build..."
    rm "$OUTPUT_NAME"
fi

# 下载依赖
echo "Downloading dependencies..."
go mod download

# 运行测试（如果启用）
if [ "$RUN_TESTS" = true ]; then
    echo "Running tests..."
    go test ./... $BUILD_TAGS
fi

# 开始构建
echo "Building binary..."
go build $BUILD_FLAGS $BUILD_TAGS -o "$OUTPUT_NAME" .

# 检查构建结果
if [ -f "$OUTPUT_NAME" ]; then
    echo "Build successful! Binary: $OUTPUT_NAME"
    # 显示文件信息
    ls -lh "$OUTPUT_NAME"
else
    echo "Build failed!"
    exit 1
fi 