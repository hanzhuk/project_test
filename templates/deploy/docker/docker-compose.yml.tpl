# docker-compose.yml - {{.ProjectName}} 开发环境编排
version: "3.8"

services:
  # 数据库服务
  db:
    image: postgres:15-alpine
    container_name: {{.ProjectName}}-db
    environment:
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-secret}
      POSTGRES_DB: ${DB_NAME:-{{.ProjectName}}_db}
    ports:
      - "${DB_PORT:-5432}:5432"
    volumes:
      - db_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-postgres}"]
      interval: 10s
      timeout: 5s
      retries: 5

  # 后端服务
  server:
    build: .
    container_name: {{.ProjectName}}-server
    ports:
      - "${PORT:-8080}:8080"
    environment:
      DB_HOST: db
      DB_PORT: 5432
      DB_USER: ${DB_USER:-postgres}
      DB_PASSWORD: ${DB_PASSWORD:-secret}
      DB_NAME: ${DB_NAME:-{{.ProjectName}}_db}
      DB_SSL_MODE: disable
      PORT: 8080
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped

volumes:
  db_data:
