# Dockerfile - {{.ProjectName}} 前端多阶段构建
# 阶段一：构建
FROM node:20-alpine AS builder

WORKDIR /app

# 复制 package.json 并安装依赖
COPY package.json package-lock.json* ./
RUN npm install

# 复制源代码并构建
COPY . .
RUN npm run build

# 阶段二：运行（Nginx 静态服务）
FROM nginx:alpine

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 Nginx 配置（可选）
# COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
