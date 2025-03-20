# AI 作业批改系统后端服务

这是一个基于 Google Vertex AI 的智能作业批改系统后端服务，支持英语、语文、数学等科目的作业批改。

## 功能特点

- 支持多种作业类型（英语、语文、数学）
- 使用 Google Vertex AI 进行智能分析
- 提供详细的作业评价和反馈
- 支持图片格式的作业上传
- 提供模拟模式用于测试

## 技术栈

- Go 语言
- Google Vertex AI (Gemini 模型)
- PostgreSQL 数据库
- Gin Web 框架

## 环境要求

- Go 1.20 或更高版本
- PostgreSQL 数据库
- Google Cloud 账号和项目
- 有效的 Google Cloud 凭证文件

## 快速开始

1. 克隆仓库：
```bash
git clone https://github.com/GiantClam/ai-exam-mark.git
cd ai-exam-mark/backend
```

2. 安装依赖：
```bash
go mod download
```

3. 配置环境变量：
```bash
cp .env.example .env
```
然后编辑 `.env` 文件，填入必要的配置信息。

4. 启动开发服务器：
```bash
./start-dev.sh
```

## 环境变量配置

主要配置项说明：

```env
# 服务配置
PORT=8080
ENV=development
LOG_LEVEL=debug

# Google Cloud 配置
GOOGLE_CLOUD_PROJECT=your-project-id
GOOGLE_CLOUD_LOCATION=your-location
GOOGLE_APPLICATION_CREDENTIALS=path/to/your/credentials.json

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=homework_marking

# CORS 配置
CORS_ORIGIN=http://localhost:3000

# 文件上传配置
MAX_FILE_SIZE=10485760  # 10MB
UPLOAD_DIR=uploads
```

## API 接口

### 作业批改接口

- URL: `/api/homework/analyze`
- 方法: POST
- 参数:
  - file: 作业图片文件
  - type: 作业类型 (english/chinese/math)
- 返回: JSON 格式的批改结果

### 模拟模式

当无法访问 Google Cloud 服务时，系统会自动切换到模拟模式，返回预设的批改结果。

## 开发指南

### 目录结构

```
backend/
├── cmd/            # 主程序入口
├── config/         # 配置管理
├── controllers/    # 控制器
├── models/         # 数据模型
├── services/       # 业务逻辑
├── utils/          # 工具函数
└── main.go         # 程序入口
```

### 构建和部署

1. 开发环境构建：
```bash
./start-dev.sh
```

2. 生产环境构建：
```bash
./start-prod.sh
```

## 注意事项

1. 确保 Google Cloud 凭证文件正确配置
2. 生产环境部署前请修改所有默认密码
3. 建议在生产环境使用 HTTPS
4. 定期备份数据库

## 许可证

MIT License 