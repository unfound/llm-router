import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { getEndpoints, getModels, toggleModel, syncModels, createEndpoint, deleteEndpoint, createModel, deleteModel } from '@/api/admin'
import { RefreshCw, Plus, Trash2, Power } from 'lucide-react'

export default function Models() {
  const [endpoints, setEndpoints] = useState<any[]>([])
  const [models, setModels] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [syncing, setSyncing] = useState(false)

  // 新增端点表单
  const [showEndpointForm, setShowEndpointForm] = useState(false)
  const [epForm, setEpForm] = useState({ name: '', api_base: '', api_key: '' })

  // 新增模型表单
  const [showModelForm, setShowModelForm] = useState(false)
  const [mForm, setMForm] = useState({ name: '', endpoint: '', model_id: '', fallback: '', max_retries: 2 })

  const load = () => {
    setLoading(true)
    Promise.all([getEndpoints(), getModels()])
      .then(([ep, m]) => {
        setEndpoints(ep.endpoints || [])
        setModels(m.models || [])
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const handleSync = async () => {
    setSyncing(true)
    await syncModels()
    setTimeout(() => { load(); setSyncing(false) }, 2000)
  }

  const handleToggle = async (id: number) => {
    await toggleModel(id)
    load()
  }

  const handleDeleteEndpoint = async (id: number) => {
    if (!confirm('确定删除该端点？关联的模型也会受影响')) return
    await deleteEndpoint(id)
    load()
  }

  const handleDeleteModel = async (id: number) => {
    if (!confirm('确定删除该模型路由？')) return
    await deleteModel(id)
    load()
  }

  const handleCreateEndpoint = async () => {
    if (!epForm.name || !epForm.api_base || !epForm.api_key) return
    await createEndpoint(epForm)
    setEpForm({ name: '', api_base: '', api_key: '' })
    setShowEndpointForm(false)
    load()
  }

  const handleCreateModel = async () => {
    if (!mForm.name || !mForm.endpoint || !mForm.model_id) return
    await createModel({ ...mForm, is_active: true })
    setMForm({ name: '', endpoint: '', model_id: '', fallback: '', max_retries: 2 })
    setShowModelForm(false)
    load()
  }

  if (loading) return <div className="p-8 text-muted-foreground">加载中...</div>

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">端点 & 模型管理</h1>

      {/* 端点列表 */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>端点 ({endpoints.length})</CardTitle>
          <div className="flex gap-2">
            <Button size="sm" onClick={() => setShowEndpointForm(!showEndpointForm)}>
              <Plus className="h-4 w-4 mr-1" /> 新增
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {showEndpointForm && (
            <div className="mb-4 p-4 border rounded-lg bg-muted/50 space-y-3">
              <div className="grid grid-cols-3 gap-3">
                <input className="col-span-1 border rounded px-3 py-2 text-sm" placeholder="名称 (如 deepseek)" value={epForm.name} onChange={e => setEpForm({ ...epForm, name: e.target.value })} />
                <input className="col-span-1 border rounded px-3 py-2 text-sm" placeholder="API Base URL" value={epForm.api_base} onChange={e => setEpForm({ ...epForm, api_base: e.target.value })} />
                <input className="col-span-1 border rounded px-3 py-2 text-sm" placeholder="API Key" type="password" value={epForm.api_key} onChange={e => setEpForm({ ...epForm, api_key: e.target.value })} />
              </div>
              <div className="flex gap-2">
                <Button size="sm" onClick={handleCreateEndpoint}>创建</Button>
                <Button size="sm" variant="outline" onClick={() => setShowEndpointForm(false)}>取消</Button>
              </div>
            </div>
          )}
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left text-muted-foreground">
                <th className="pb-2 font-medium">名称</th>
                <th className="pb-2 font-medium">API Base</th>
                <th className="pb-2 font-medium">API Key</th>
                <th className="pb-2 font-medium text-right">操作</th>
              </tr>
            </thead>
            <tbody>
              {endpoints.map(ep => (
                <tr key={ep.id} className="border-b last:border-0">
                  <td className="py-3 font-medium">{ep.name}</td>
                  <td className="py-3 text-muted-foreground">{ep.api_base}</td>
                  <td className="py-3 text-muted-foreground">***{ep.api_key.slice(-4)}</td>
                  <td className="py-3 text-right">
                    <Button variant="ghost" size="icon" onClick={() => handleDeleteEndpoint(ep.id)}>
                      <Trash2 className="h-4 w-4 text-red-500" />
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>

      {/* 模型列表 */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>模型路由 ({models.length})</CardTitle>
          <div className="flex gap-2">
            <Button size="sm" variant="outline" onClick={handleSync} disabled={syncing}>
              <RefreshCw className={`h-4 w-4 mr-1 ${syncing ? 'animate-spin' : ''}`} /> 发现模型
            </Button>
            <Button size="sm" onClick={() => setShowModelForm(!showModelForm)}>
              <Plus className="h-4 w-4 mr-1" /> 新增
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {showModelForm && (
            <div className="mb-4 p-4 border rounded-lg bg-muted/50 space-y-3">
              <div className="grid grid-cols-5 gap-3">
                <input className="border rounded px-3 py-2 text-sm" placeholder="别名 (如 cheap)" value={mForm.name} onChange={e => setMForm({ ...mForm, name: e.target.value })} />
                <select className="border rounded px-3 py-2 text-sm" value={mForm.endpoint} onChange={e => setMForm({ ...mForm, endpoint: e.target.value })}>
                  <option value="">选择端点</option>
                  {endpoints.map(ep => <option key={ep.id} value={ep.name}>{ep.name}</option>)}
                </select>
                <input className="border rounded px-3 py-2 text-sm" placeholder="model_id" value={mForm.model_id} onChange={e => setMForm({ ...mForm, model_id: e.target.value })} />
                <input className="border rounded px-3 py-2 text-sm" placeholder="fallback 别名" value={mForm.fallback} onChange={e => setMForm({ ...mForm, fallback: e.target.value })} />
                <input className="border rounded px-3 py-2 text-sm" type="number" placeholder="重试次数" value={mForm.max_retries} onChange={e => setMForm({ ...mForm, max_retries: Number(e.target.value) })} />
              </div>
              <div className="flex gap-2">
                <Button size="sm" onClick={handleCreateModel}>创建</Button>
                <Button size="sm" variant="outline" onClick={() => setShowModelForm(false)}>取消</Button>
              </div>
            </div>
          )}
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left text-muted-foreground">
                <th className="pb-2 font-medium">别名</th>
                <th className="pb-2 font-medium">端点</th>
                <th className="pb-2 font-medium">模型 ID</th>
                <th className="pb-2 font-medium">Fallback</th>
                <th className="pb-2 font-medium">重试</th>
                <th className="pb-2 font-medium">来源</th>
                <th className="pb-2 font-medium">状态</th>
                <th className="pb-2 font-medium text-right">操作</th>
              </tr>
            </thead>
            <tbody>
              {models.map(m => (
                <tr key={m.id} className="border-b last:border-0">
                  <td className="py-3 font-medium">{m.name}</td>
                  <td className="py-3">{m.endpoint_name}</td>
                  <td className="py-3 text-muted-foreground">{m.model_id}</td>
                  <td className="py-3 text-muted-foreground">{m.fallback || '-'}</td>
                  <td className="py-3 text-center">{m.max_retries}</td>
                  <td className="py-3">
                    <Badge variant="outline" className={m.discovered ? 'border-blue-300 text-blue-600' : ''}>
                      {m.discovered ? '发现' : '配置'}
                    </Badge>
                  </td>
                  <td className="py-3">
                    <Badge className={m.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}>
                      {m.is_active ? '启用' : '禁用'}
                    </Badge>
                  </td>
                  <td className="py-3 text-right">
                    <div className="flex gap-1 justify-end">
                      <Button variant="ghost" size="icon" onClick={() => handleToggle(m.id)}>
                        <Power className={`h-4 w-4 ${m.is_active ? 'text-green-500' : 'text-gray-400'}`} />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => handleDeleteModel(m.id)}>
                        <Trash2 className="h-4 w-4 text-red-500" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  )
}
