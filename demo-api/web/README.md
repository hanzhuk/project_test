# demo-api Web

demo-api 前端应用，使用 React 19 + TypeScript + Vite 构建。

## 技术栈

- **框架**: React 19
- **语言**: TypeScript
- **构建工具**: Vite 6
- **样式**: TailwindCSS
- **API 客户端**: openapi-fetch（类型安全）

## 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

开发服务器默认启动在 `http://localhost:3000`，API 请求自动代理到 `http://localhost:8080`。

### 3. 构建生产版本

```bash
npm run build
```

构建产物输出到 `dist/` 目录。

### 4. 预览生产版本

```bash
npm run preview
```

## 项目结构

```
src/
├── api/           # API 客户端与类型
│   ├── client.ts  # API 客户端封装
│   └── api-types.ts # API 类型定义
├── components/    # 通用组件
├── pages/         # 页面组件
├── stores/        # 状态管理
├── types/         # 类型定义
├── App.tsx        # 根组件
├── main.tsx       # 应用入口
└── index.css      # 全局样式
```

## 环境变量

在项目根目录创建 `.env` 文件：

```
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## Docker 部署

```bash
docker build -t demo-api-web:latest .
docker run -p 80:80 demo-api-web:latest
```
