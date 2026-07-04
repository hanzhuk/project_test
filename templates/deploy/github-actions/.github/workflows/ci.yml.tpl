# .github/workflows/ci.yml - {{.ProjectName}} CI/CD 配置
name: CI

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  # 后端测试与构建
  backend:
    name: Backend CI
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: secret
          POSTGRES_DB: test_db
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Download dependencies
        run: go mod download

      - name: Run go vet
        run: go vet ./...

      - name: Run tests
        env:
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: postgres
          DB_PASSWORD: secret
          DB_NAME: test_db
          DB_SSL_MODE: disable
        run: go test ./... -v

      - name: Build
        run: go build -ldflags "-s -w" -o bin/server ./cmd/server

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: server-binary
          path: bin/server

  # Docker 镜像构建
  docker:
    name: Docker Build
    runs-on: ubuntu-latest
    needs: backend
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Build Docker image
        run: docker build -t {{.ProjectName}}:latest .
