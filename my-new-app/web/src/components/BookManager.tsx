import { useState, useEffect, FormEvent } from 'react'

interface BookItem {
  id: number
  title: string
  author: string
  price: number
}

export default function BookManager() {
  const [items, setItems] = useState<BookItem[]>([])
  const [editingId, setEditingId] = useState<number | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [formData, setFormData] = useState({ title: '', author: '', price: '' })

  const loadItems = () => {
    fetch('/api/v1/books')
      .then(r => r.json())
      .then(json => { if (json.code === 0) setItems(json.data ?? []) })
      .catch(console.error)
  }

  useEffect(() => { loadItems() }, [])

  const resetForm = () => {
    setFormData({ title: '', author: '', price: '' })
    setEditingId(null)
  }

  const filtered = searchQuery
    ? items.filter(i => i.title?.toLowerCase().includes(searchQuery.toLowerCase()) || i.author?.toLowerCase().includes(searchQuery.toLowerCase()))
    : items

  const handleSave = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const body = { title: formData.title.trim(), author: formData.author.trim(), price: parseFloat(formData.price) || 0 }

    if (editingId === -1) {
      fetch('/api/v1/books', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
        .then(r => r.json())
        .then(json => { if (json.code === 0) { loadItems(); resetForm() } else alert('保存失败: ' + (json.message ?? '未知错误')) })
    } else if (editingId !== null) {
      fetch(`/api/v1/books/${editingId}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
        .then(r => r.json())
        .then(json => { if (json.code === 0) { loadItems(); resetForm() } else alert('更新失败: ' + (json.message ?? '未知错误')) })
    }
  }

  const handleDelete = (id: number) => {
    if (!confirm('确定删除？')) return
    fetch(`/api/v1/books/${id}`, { method: 'DELETE' })
      .then(r => r.json())
      .then(json => { if (json.code === 0) loadItems() })
  }

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      <h1 className="text-2xl font-bold mb-4 text-gray-800">图书管理</h1>

      <div className="flex gap-3 mb-4">
        <input type="text" placeholder="搜索书名或作者..." value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
          className="flex-1 p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-400" />
        <button onClick={() => { setEditingId(-1); setFormData({ title: '', author: '', price: '' }) }}
          className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
          新增图书
        </button>
      </div>

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
            {filtered.length === 0 ? (
              <tr><td colSpan={5} className="p-6 text-center text-gray-400">暂无数据，请新增图书</td></tr>
            ) : filtered.map(item => (
              <tr key={item.id} className="border-t hover:bg-gray-50">
                <td className="p-3 text-sm text-gray-500">{item.id}</td>
                <td className="p-3 font-medium">{item.title}</td>
                <td className="p-3 text-gray-600">{item.author}</td>
                <td className="p-3 text-right text-gray-700">{item.price}</td>
                <td className="p-3 text-center">
                  <button onClick={() => { setEditingId(item.id); setFormData({ title: item.title, author: item.author || '', price: String(item.price ?? '') }) }}
                    className="text-blue-600 hover:underline text-sm mr-3">编辑</button>
                  <button onClick={() => handleDelete(item.id)}
                    className="text-red-600 hover:underline text-sm">删除</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {editingId !== null && (
        <form onSubmit={handleSave} className="bg-white border rounded-lg p-5 shadow-sm max-w-lg">
          <h2 className="font-semibold text-lg mb-4">{editingId === -1 ? '新增图书' : `编辑图书 #${editingId}`}</h2>
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
