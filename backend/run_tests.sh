#!/bin/bash

# 设置日志输出颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}正在运行作业批改系统测试...${NC}"

# 运行模拟模式测试
echo -e "${BLUE}测试模拟模式功能:${NC}"
cd "$(dirname "$0")" # 确保在backend目录中
export UseMockMode=true # 设置环境变量启用模拟模式
go test -v ./services -run TestMockMode

# 检查测试结果
if [ $? -eq 0 ]; then
    echo -e "${GREEN}所有测试通过!${NC}"
    exit 0
else
    echo -e "${RED}测试失败!${NC}"
    exit 1
fi 