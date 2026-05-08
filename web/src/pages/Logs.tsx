import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { getLogs, getLogDetail } from '@/api/admin'
import { ChevronLeft, ChevronRight, Eye } from 'lucide-react'

export default function Logs() {
  const [logs, setLogs] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [filters, setFilters] = useState({ model_name: '', status: '' })
  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<any>(null)

  const limit = 20

  const load = () => {
    setLoading(true)
    getLogs({ ...filters, limit, offset: page * limit })
      .then(data => { setLogs(data.logs || []); setTotal(data.total || 0) })
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [page, filters])

  const showDetail = async (id: number) => {
    const d = await getLogDetail(id)
    setDetail(d)
  }

  const pages = Math.ceil(total / limit)

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">日志</h1>

      {/* 筛选栏 */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex gap-4 items-end">
            <div className="flex-1">
              <label className="text-sm text-muted-foreground mb-1 block">模型</label>
              <input
                className="w-full border rounded px-3 py-2 text-sm"
                placeholder="筛选模型..."
                value={filters.model_name}
                onChange={e => { setFilters({ ...filters, model_name: e.target.value }); setPage(0) }}
              />
            </div>
            <div className="w-32">
              <label className="text-sm text-muted-foreground mb-1 block">状态</label>
              <select
                className="w-full border rounded px-3 py-2 text-sm"
                value={filters.status}
                onChange={e => { setFilters({ ...filters, status: e.target.value }); setPage(0) }}
              >
                <option value="">全部</option>
                <option value="success">成功</option>
                <option value="failed">失败</option>
              </select>
            </div>
            <Button variant="outline" onClick={() => { setFilters({ model_name: '', status: '' }); setPage(0) }}>
              重置
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* 日志列表 */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>日志列表 ({total} 条)</CardTitle>
          <div className="flex gap-2 items-center text-sm text-muted-foreground">
            <Button variant="outline" size="sm" disabled={page === 0} onClick={() => setPage(p => p - 1)}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span>{page + 1} / {pages || 1}</span>
            <Button variant="outline" size="sm" disabled={page >= pages - 1} onClick={() => setPage(p => p + 1)}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="py-8 text-center text-muted-foreground">加载中...</div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-2 font-medium">ID</th>
                  <th className="pb-2 font-medium">时间</th>
                  <th className="pb-2 font-medium">模型</th>
                  <th className="pb-2 font-medium">别名</th>
                  <th className="pb-2 font-medium">状态</th>
                  <th className="pb-2 font-medium text-right">延迟</th>
                  <th className="pb-2 font-medium text-right">Token</th>
                  <th className="pb-2 font-medium">摘要</th>
                  <th className="pb-2 font-medium text-right">详情</th>
                </tr>
              </thead>
              <tbody>
                {logs.map(log => (
                  <tr key={log.id} className="border-b last:border-0 hover:bg-muted/50">
                    <td className="py-3 text-muted-foreground">{log.id}</td>
                    <td className="py-3 text-muted-foreground whitespace-nowrap">{log.created_at?.replace('T', ' ').slice(0, 19)}</td>
                    <td className="py-3">{log.model_name}</td>
                    <td className="py-3 text-muted-foreground">{log.alias_name}</td>
                    <td className="py-3">
                      <Badge className={log.status === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}>
                        {log.status === 'success' ? '成功' : '失败'}
                      </Badge>
                    </td>
                    <td className="py-3 text-right">{log.latency_ms}ms</td>
                    <td className="py-3 text-right">{log.total_tokens}</td>
                    <td className="py-3 text-muted-foreground max-w-[200px] truncate">{log.request_summary || '-'}</td>
                    <td className="py-3 text-right">
                      <Button variant="ghost" size="icon" onClick={() => showDetail(log.id)}>
                        <Eye className="h-4 w-4" />
                      </Button>
                    </td>
                  </tr>
                ))}
                {logs.length === 0 && (
                  <tr><td colSpan={9} className="py-8 text-center text-muted-foreground">暂无日志</td></tr>
                )}
              </tbody>
            </table>
          )}
        </CardContent>
      </Card>

      {/* 日志详情弹窗 */}
      {detail && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setDetail(null)}>
          <div className="bg-background rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[80vh] overflow-auto" onClick={e => e.stopPropagation()}>
            <div className="p-6 border-b flex justify-between items-center">
              <h2 className="text-lg font-semibold">日志详情 #{detail.id}</h2>
              <Button variant="ghost" size="sm" onClick={() => setDetail(null)}>关闭</Button>
            </div>
            <div className="p-6 space-y-4 text-sm">
              <div className="grid grid-cols-2 gap-4">
                <div><span className="text-muted-foreground">模型：</span>{detail.model_name}</div>
                <div><span className="text-muted-foreground">别名：</span>{detail.alias_name}</div>
                <div><span className="text-muted-foreground">状态：</span>
                  <Badge className={detail.status === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}>
                    {detail.status}
                  </Badge>
                </div>
                <div><span className="text-muted-foreground">延迟：</span>{detail.latency_ms}ms</div>
                <div><span className="text-muted-foreground">Token：</span>{detail.total_tokens}</div>
                <div><span className="text-muted-foreground">时间：</span>{detail.created_at}</div>
              </div>
              {detail.error_message && (
                <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700">{detail.error_message}</div>
              )}
              {detail.request_summary && (
                <div>
                  <div className="text-muted-foreground mb-1">请求摘要：</div>
                  <pre className="p-3 bg-muted rounded text-xs overflow-auto">{detail.request_summary}</pre>
                </div>
              )}
              {detail.request_body && (
                <div>
                  <div className="text-muted-foreground mb-1">完整请求：</div>
                  <pre className="p-3 bg-muted rounded text-xs overflow-auto max-h-48">{detail.request_body}</pre>
                </div>
              )}
              {detail.response_body && (
                <div>
                  <div className="text-muted-foreground mb-1">完整响应：</div>
                  <pre className="p-3 bg-muted rounded text-xs overflow-auto max-h-48">{detail.response_body}</pre>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
