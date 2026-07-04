// src/App.tsx - 应用根组件
import { useState, useEffect } from 'react'
import { api } from './api/client'

// App 是应用的根组件，展示项目信息和健康检查状态
function App() {
  const [health, setHealth] = useState<string>('检查中...')

  // 组件挂载时检查后端健康状态
  useEffect(() => {
    api.GET('/health')
      .then(({ data }) => {
        if (data) {
          setHealth('后端服务正常 ✓')
        }
      })
      .catch(() => {
        setHealth('后端服务不可用 ✗')
      })
  }, [])

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="bg-white p-8 rounded-lg shadow-md max-w-md w-full">
        <h1 className="text-2xl font-bold text-gray-800 mb-4">
          {{.ProjectName}}
        </h1>
        <p className="text-gray-600 mb-4">
          基于 Go + {{.BackendTitle}} + {{.ORMTitle}} + {{.DBTitle}} 的全栈应用
        </p>
        <div className="border-t pt-4">
          <p className="text-sm text-gray-500">服务状态: {health}</p>
        </div>
        <div className="mt-6 text-sm text-gray-400">
          <p>前端: React 19 + TypeScript + Vite</p>
        </div>
      </div>
    </div>
  )
}

export default App
