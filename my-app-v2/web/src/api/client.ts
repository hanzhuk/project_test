// src/api/client.ts - API 客户端封装
// 使用 openapi-fetch 调用后端 API，类型安全
import createClient from 'openapi-fetch'
import type { paths } from './api-types'

// 创建 API 客户端，baseUrl 从环境变量读取
const client = createClient<paths>({
  baseUrl: import.meta.env.VITE_API_BASE_URL || '/api/v1',
})

// 统一导出 API 客户端
export const api = client

// 通用请求封装，带错误处理
export async function request<T>(promise: Promise<{ data?: T; error?: unknown }>): Promise<T> {
  const { data, error } = await promise
  if (error) {
    throw error
  }
  return data as T
}
