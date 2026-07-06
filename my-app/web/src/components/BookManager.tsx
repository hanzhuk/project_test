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
  const [searchTerm, setSearchTerm] = useState('')
  
  // 表单状态
  const [formData, setFormData] = useState({
    title: '',
    author: '',
    price: ''
  })

  // 错误与加载状态
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchBooks = () => {
    setLoading(true)
    fetch('/api/v1/books')
      .then(r => r.json())
      .then(json => {
        if (json.code === 0) {
          setItems(json.data || [])
          setError(null)
        } else {
          setError(json.message || '获取书籍列表失败')
        }
      })
      .catch(err => {
        console.error(err)
        setError('网络错误，无法连接服务器')
      })
      .finally(() => {
        setLoading(false)
      })
  }

  useEffect(() => {
    fetchBooks()
  }, [])

  const resetForm = () => {
    setFormData({ title: '', author: '', price: '' })
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    
    if (!formData.title.trim() || !formData.author.trim() || !formData.price) {
      alert('请填写所有必填字段')
      return
    }
    
    setLoading(true)
    let response: Response | null = null
    
    try {
      if (editingId === -1) {
        // 新增：POST /api/v1/books
        response = await fetch('/api/v1/books', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            title: formData.title.trim(),
            author: formData.author.trim(),
            price: Number(formData.price)
          })
        })
      } else if (editingId !== null && editingId > 0) {
        // 编辑：PUT /api/v1/books/{id}
        response = await fetch(`/api/v1/books/${editingId}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            title: formData.title.trim(),
            author: formData.author.trim(),
            price: Number(formData.price)
          })
        })
      } else {
        setLoading(false)
        return
      }

      const json = await response.json().catch(() => null)

      if (response.status === 200 || response.status === 201) {
        resetForm()
        setEditingId(null)
        fetchBooks()
      } else {
        alert('操作失败: ' + (json?.message || '未知错误'))
      }
    } catch (err: any) {
      console.error(err)
      alert('网络错误：' + err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('确定删除这本书吗？')) return
    
    setLoading(true)
    try {
      const res = await fetch(`/api/v1/books/${id}`, { method: 'DELETE' })
      const json = await res.json().catch(() => null)
      
      if (res.status === 200 || res.status === 204) {
        fetchBooks()
      } else {
        alert('删除失败: ' + (json?.message || '未知错误'))
      }
    } catch (err: any) {
      console.error(err)
      alert('网络错误：' + err.message)
    } finally {
      setLoading(false)
    }
  }

  const filteredItems = items.filter(item => 
    (item.title && item.title.toLowerCase().includes(searchTerm.toLowerCase())) ||
    (item.author && item.author.toLowerCase().includes(searchTerm.toLowerCase()))
  )

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <h2 className="text-2xl font-bold mb-6 text-gray-800">图书管理模块</h2>

      {/* 搜索与新增栏 */}
      <div className="mb-6 flex flex-col md:flex-row gap-4 items-center justify-between">
        <input
          type="text"
          placeholder="按书名/作者搜索..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full md:w-80 px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button 
          onClick={() => { setEditingId(-1); resetForm() }} 
          className="w-full md:w-auto bg-green-600 text-white px-5 py-2.5 rounded-lg hover:bg-green-700 transition-colors shadow-sm font-medium"
        >
          新增图书
        </button>
      </div>

      {error && (
        <div className="mb-4 p-4 text-red-700 bg-red-50 border border-red-200 rounded-lg">
          {error}
        </div>
      )}

      {/* 表格 */}
      <div className="bg-white shadow-md rounded-lg overflow-hidden border">
        <table className="w-full border-collapse">
          <thead className="bg-gray-100 border-b">
            <tr>
              <th className="p-4 text-left font-semibold text-gray-700">ID</th>
              <th className="p-4 text-left font-semibold text-gray-700">书名</th>
              <th className="p-4 text-left font-semibold text-gray-700">作者</th>
              <th className="p-4 text-right font-semibold text-gray-700">价格</th>
              <th className="p-4 text-center font-semibold text-gray-700">操作</th>
            </tr>
          </thead>
          <tbody>
            {filteredItems.map(item => (
              <tr key={item.id} className="hover:bg-gray-50 transition-colors border-b last:border-none">
                <td className="p-4 text-left text-gray-600">{item.id}</td>
                <td className="p-4 text-left font-medium text-gray-800">{item.title}</td>
                <td className="p-4 text-left text-gray-600">{item.author}</td>
                <td className="p-4 text-right text-gray-800">${Number(item.price).toFixed(2)}</td>
                <td className="p-4 text-center">
                  <div className="flex gap-2 justify-center">
                    <button 
                      onClick={() => { 
                        setEditingId(item.id)
                        setFormData({
                          title: item.title,
                          author: item.author,
                          price: String(item.price)
                        })
                      }} 
                      className="text-blue-600 hover:text-blue-800 font-medium transition-colors"
                    >
                      编辑
                    </button>
                    <button 
                      onClick={() => handleDelete(item.id)} 
                      className="text-red-600 hover:text-red-800 font-medium transition-colors"
                    >
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {filteredItems.length === 0 && (
          <div className="text-center text-gray-500 py-10">
            {loading ? '加载中...' : '暂无图书数据'}
          </div>
        )}
      </div>

      {/* 编辑/新增对话框或表单 */}
      {editingId !== null && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50 animate-fade-in">
          <form 
            onSubmit={handleSubmit} 
            className="bg-white p-6 rounded-xl shadow-xl border w-full max-w-md"
          >
            <h3 className="text-xl font-bold mb-4 text-gray-800">
              {editingId === -1 ? '新增图书' : `编辑图书 #${editingId}`}
            </h3>
            
            <div className="mb-4">
              <label className="block text-sm font-semibold mb-1.5 text-gray-700">书名 *</label>
              <input 
                type="text" 
                value={formData.title} 
                onChange={(e) => setFormData({...formData, title: e.target.value})} 
                placeholder="请输入书名" 
                required 
                className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div className="mb-4">
              <label className="block text-sm font-semibold mb-1.5 text-gray-700">作者 *</label>
              <input 
                type="text" 
                value={formData.author} 
                onChange={(e) => setFormData({...formData, author: e.target.value})} 
                placeholder="请输入作者" 
                required 
                className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div className="mb-6">
              <label className="block text-sm font-semibold mb-1.5 text-gray-700">价格 *</label>
              <input 
                type="number" 
                step="0.01" 
                value={formData.price} 
                onChange={(e) => setFormData({...formData, price: e.target.value})} 
                placeholder="请输入价格" 
                required 
                className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div className="flex justify-end gap-3">
              <button 
                type="button" 
                onClick={() => setEditingId(null)} 
                className="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium"
              >
                取消
              </button>
              <button 
                type="submit" 
                disabled={loading}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:bg-blue-300"
              >
                {loading ? '保存中...' : '保存'}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  )
}