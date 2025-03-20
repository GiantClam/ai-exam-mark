# 作业批改系统

一个基于 Next.js 和 Go 的作业批改系统，支持手写文字识别和自动评分。

## 功能特点

- 支持上传作业图片
- 自动识别手写文字
- 支持单栏和双栏布局的试卷
- 实时预览和反馈
- 响应式设计

## 技术栈

### 前端
- Next.js 14
- React 18
- TypeScript
- Tailwind CSS
- shadcn/ui

### 后端
- Go
- Gin 框架
- Google Cloud Vision API
- Gemini API

## 本地开发

### 前置要求

- Node.js 18+
- Go 1.20+
- npm 或 yarn
- Google Cloud 账号和凭证

### 安装步骤

1. 克隆仓库
```bash
git clone https://github.com/yourusername/homework-marking.git
cd homework-marking
```

2. 安装前端依赖
```bash
npm install
```

3. 安装后端依赖
```bash
cd backend
go mod download
cd ..
```

4. 配置环境变量
```bash
# 复制环境变量示例文件
cp .env.example .env.local
cp backend/.env.example backend/.env

# 编辑环境变量文件，填入必要的配置
```

5. 配置 Google Cloud 凭证
- 将你的 Google Cloud 服务账号密钥文件放在 `backend` 目录下
- 更新 `backend/.env` 中的 `GOOGLE_APPLICATION_CREDENTIALS` 路径

6. 启动开发服务器
```bash
# 启动后端服务
cd backend
./run_server.sh

# 新开一个终端，启动前端服务
npm run dev
```

访问 http://localhost:3000 查看应用

## 部署

### 前端部署 (Vercel)

1. 在 Vercel 上创建新项目
2. 导入 GitHub 仓库
3. 配置环境变量
4. 部署

### 后端部署

1. 准备服务器环境
2. 配置环境变量
3. 编译并运行后端服务

## 项目结构

```
homework-marking/
├── frontend/           # Next.js 前端应用
│   ├── src/           # 源代码
│   ├── public/        # 静态资源
│   └── package.json   # 前端依赖配置
├── backend/           # Go 后端服务
│   ├── cmd/          # 主程序入口
│   ├── handlers/     # 请求处理器
│   ├── models/       # 数据模型
│   └── services/     # 业务逻辑
└── README.md         # 项目文档
```

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

MIT License - 详见 LICENSE 文件 