// src/api/api-types.ts - API 类型定义占位
// 运行 `npm run gen-api` 从后端 OpenAPI 文档生成类型
// 此文件为占位，实际类型由 openapi-typescript 生成

export interface paths {
  '/health': {
    get: {
      responses: {
        200: {
          content: {
            'application/json': {
              status: string
            }
          }
        }
      }
    }
  }
  '/api/v1/ping': {
    get: {
      responses: {
        200: {
          content: {
            'application/json': {
              code: number
              message: string
              data?: unknown
            }
          }
        }
      }
    }
  }
}
