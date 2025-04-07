# AI 作业批改系统

这是一个基于人工智能的作业批改系统，支持自动分析和评分学生提交的作业。系统可以处理多种科目的作业，包括英语、语文和数学等。

## 主要功能

- 支持上传PDF或图片格式的作业文件
- 自动识别作业内容并进行智能批改
- 支持多种布局方式（单栏、双栏）
- 支持多种科目（通用、数学、英语、语文）
- 异步处理作业，实时显示处理进度
- 显示详细的批改结果，包括每道题的评价

## 技术栈

### 前端
- Next.js/React
- TypeScript
- Ant Design UI库
- Axios用于API通信
- Tailwind CSS样式库

### 后端
- Go语言
- Gin Web框架
- Google Vertex AI (Gemini模型)用于AI分析
- JWT认证

## 项目结构

```
/
├── backend/                # Go后端服务
│   ├── handlers/           # 请求处理器
│   ├── services/           # 业务逻辑
│   ├── utils/              # 工具函数
│   ├── routes/             # API路由
│   └── main.go             # 程序入口
├── src/                    # 前端React代码
│   ├── components/         # 可复用组件
│   └── pages/              # 页面组件
├── public/                 # 静态资源
└── README.md               # 项目说明
```

## 安装与运行

### 前提条件

- Node.js 18+
- Go 1.20+
- Google Cloud账号和项目（用于Vertex AI）

### 后端设置

1. 进入后端目录：
   ```bash
   cd backend
   ```

2. 安装依赖：
   ```bash
   go mod download
   ```

3. 配置环境变量：
   ```bash
   cp .env.example .env
   ```
   
4. 修改`.env`文件，填入必要的配置信息，尤其是Google Cloud项目信息和凭证路径。

5. 启动后端服务：
   ```bash
   ./start-dev.sh
   ```

### 前端设置

1. 安装依赖：
   ```bash
   npm install
   ```

2. 启动开发服务器：
   ```bash
   npm run dev
   ```

3. 访问 http://localhost:3000

## 上传到GitHub前的安全检查

**重要**：在将代码上传到GitHub之前，请确保删除或替换以下敏感信息：

1. **Google Cloud凭证文件**:
   - 删除`backend/zippy-aurora-444204-q2-83e9a4179999.json`
   - 删除`backend/test-credentials.json`
   - 将这些文件添加到`.gitignore`

2. **JWT密钥**:
   - 修改`backend/utils/jwt.go`中的`jwtKey`变量，使用环境变量代替硬编码的密钥

3. **环境变量文件**:
   - 移除`.env`文件（保留`.env.example`作为模板）
   - 确保`.env`和`.env.production`已添加到`.gitignore`

4. **代理和敏感URL**:
   - 检查并移除本地代理设置
   - 移除任何内部URL或API端点

## 使用示例

1. 打开系统首页
2. 选择要上传的作业文件（PDF或图片）
3. 选择作业类型、布局方式和每个学生的页数
4. 点击"上传并开始批改"按钮
5. 系统将开始处理作业并显示进度
6. 批改完成后，系统会显示详细的评分结果和反馈

## 许可证

MIT License 