import { useState, useEffect, FormEvent } from 'react'

interface BookItem {
  id: number
  title: string
  author: string
  isbn?: string
  price?: number
}

export default function BookManager() {
  const [items, setItems] = useState<BookItem[]>([])
  const [editingId, setEditingId] = useState<number | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [formData, setFormData] = useState({ title: '', author: '', isbn: '', price: '' })

  const loadBooks = () => {
    fetch('/api/v1/books')
      .then(r => r.json())
      .then(json => { if (json.code === 0) setItems(json.data ?? []) })
      .catch(console.error)
  }

  useEffect(() => { loadBooks() }, [])

  const resetForm = () => {
    setFormData({ title: '', author: '', isbn: '', price: '' })
    setEditingId(null)
  }

  // 搜索过滤在渲染时进行，不修改 items 状态
  const filteredItems = searchQuery.trim()
    ? items.filter(b =>
        b.title?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        b.author?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        b.isbn?.includes(searchQuery)
      )
    : items

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!formData.title.trim()) return

    const body = {
      title: formData.title.trim(),
      author: formData.author.trim(),
      price: parseFloat(formData.price) || 0,
    }

    if (editingId === -1) {
      fetch('/api/v1/books', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
        .then(r => r.json())
        .then(json => {
          if (json.code === 0) { loadBooks(); resetForm() }
          else alert('保存失败: ' + (json.message ?? '未知错误'))
        })
    } else if (editingId !== null) {
      fetch(`/api/v1/books/${editingId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
        .then(r => r.json())
        .then(json => {
          if (json.code === 0) { loadBooks(); resetForm() }
          else alert('更新失败: ' + (json.message ?? '未知错误'))
        })
    }
  }

  const handleDelete = (id: number) => {
    if (!confirm('确定要删除这本书吗？')) return
    fetch(`/api/v1/books/${id}`, { method: 'DELETE' })
      .then(r => r.json())
      .then(json => {
        if (json.code === 0) loadBooks()
        else alert('删除失败')
      })
  }

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      <h1 className="text-2xl font-bold mb-4 text-gray-800">书籍管理</h1>

      <div className="flex gap-3 mb-4">
        <input
          type="text"
          placeholder="按书名、作者搜索..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
          className="flex-1 p-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button
          onClick={() => { setEditingId(-1); setFormData({ title: '', author: '', isbn: '', price: '' }) }}
          className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
        >
          新增书籍
        </button>
      </div>

      {/* 表格 */}
      <div className="bg-white border shadow-sm rounded-lg overflow-hidden mb-4">
        <table className="w-full">
          <thead className="bg-gray-100 border-b">
            <tr>
              <th className="p-3 text-left text-sm font-medium text-gray-700">ID</th>
              <th className="p-3 text-left text-sm font-medium text-gray-700">书名</th>
              <th className="p-3 text-left text-sm font-medium text-gray-700">作者</th>
              <th className="p-3 text-right text-sm font-medium text-gray-700">价格</th>
              <th className="p-3 text-center text-sm font-medium text-gray-700">操作</th>
            </tr>
          </thead>
          <tbody>
            {filteredItems.length === 0 ? (
              <tr><td colSpan={5} className="p-6 text-center text-gray-400">暂无数据，请新增书籍</td></tr>
            ) : filteredItems.map(item => (
              <tr key={item.id} className="border-t hover:bg-gray-50">
                <td className="p-3 text-sm text-gray-500">{item.id}</td>
                <td className="p-3 font-medium">{item.title}</td>
                <td className="p-3 text-gray-600">{item.author}</td>
                <td className="p-3 text-right text-gray-700">{item.price ?? '-'}</td>
                <td className="p-3 text-center">
                  <button
                    onClick={() => {
                      setEditingId(item.id)
                      setFormData({ title: item.title, author: item.author || '', isbn: item.isbn || '', price: String(item.price ?? '') })
                    }}
                    className="text-blue-600 hover:underline text-sm mr-3"
                  >编辑</button>
                  <button
                    onClick={() => handleDelete(item.id)}
                    className="text-red-600 hover:underline text-sm"
                  >删除</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* 新增/编辑表单 */}
      {editingId !== null && (
        <form onSubmit={handleSubmit} className="bg-white border rounded-lg p-5 shadow-sm max-w-lg">
          <h2 className="font-semibold text-lg mb-4">{editingId === -1 ? '新增书籍' : `编辑书籍 #${editingId}`}</h2>
          <div className="mb-3">
            <label className="block text-sm font-medium mb-1">书名 *</label>
            <input required type="text" value={formData.title}
              onChange={e => setFormData(p => ({ ...p, title: e.target.value }))}
              className="w-full border rounded p-2 focus:outline-none focus:ring-2 focus:ring-blue-400" />
          </div>
          <div className="mb-3">
            <label className="block text-sm font-medium mb-1">作者</label>
            <input type="text" value={formData.author}
              onChange={e => setFormData(p => ({ ...p, author: e.target.value }))}
              className="w-full border rounded p-2 focus:outline-none focus:ring-2 focus:ring-blue-400" />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-1">价格</label>
            <input type="number" step="0.01" value={formData.price}
              onChange={e => setFormData(p => ({ ...p, price: e.target.value }))}
              className="w-full border rounded p-2 focus:outline-none focus:ring-2 focus:ring-blue-400" />
          </div>
          <div className="flex gap-2">
            <button type="submit" className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">保存</button>
            <button type="button" onClick={resetForm} className="bg-gray-400 text-white px-4 py-2 rounded hover:bg-gray-500">取消</button>
          </div>
        </form>
      )}
    </div>
  )
}
