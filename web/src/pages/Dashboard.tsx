import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getStatsOverview, getStatsModels } from '@/api/admin'
import { Activity, CheckCircle, XCircle, Clock, Coins } from 'lucide-react'

export default function Dashboard() {
  const [overview, setOverview] = useState<any>(null)
  const [modelStats, setModelStats] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([getStatsOverview(), getStatsModels()])
      .then(([ov, ms]) => {
        setOverview(ov)
        setModelStats(ms.models || [])
      })
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-muted-foreground">加载中...</div>

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">看板</h1>

      {/* 概览卡片 */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
        <StatCard title="总请求" value={overview?.total_requests ?? 0} icon={<Activity className="h-4 w-4" />} />
        <StatCard title="成功" value={overview?.success_count ?? 0} icon={<CheckCircle className="h-4 w-4 text-green-500" />} />
        <StatCard title="失败" value={overview?.fail_count ?? 0} icon={<XCircle className="h-4 w-4 text-red-500" />} />
        <StatCard title="平均延迟" value={`${Math.round(overview?.avg_latency_ms ?? 0)}ms`} icon={<Clock className="h-4 w-4" />} />
        <StatCard title="总 Token" value={overview?.total_tokens ?? 0} icon={<Coins className="h-4 w-4" />} />
      </div>

      {/* 成功率 */}
      <Card>
        <CardHeader>
          <CardTitle>成功率</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <div className="text-4xl font-bold">{(overview?.success_rate ?? 0).toFixed(1)}%</div>
            <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
              <div
                className="h-full bg-green-500 transition-all"
                style={{ width: `${overview?.success_rate ?? 0}%` }}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 模型统计表 */}
      <Card>
        <CardHeader>
          <CardTitle>模型调用统计</CardTitle>
        </CardHeader>
        <CardContent>
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left text-muted-foreground">
                <th className="pb-2 font-medium">模型</th>
                <th className="pb-2 font-medium text-right">总调用</th>
                <th className="pb-2 font-medium text-right">成功</th>
                <th className="pb-2 font-medium text-right">失败</th>
                <th className="pb-2 font-medium text-right">平均延迟</th>
                <th className="pb-2 font-medium text-right">Token</th>
              </tr>
            </thead>
            <tbody>
              {modelStats.map((m, i) => (
                <tr key={i} className="border-b last:border-0">
                  <td className="py-3 font-medium">{m.name}</td>
                  <td className="py-3 text-right">{m.total}</td>
                  <td className="py-3 text-right text-green-600">{m.success}</td>
                  <td className="py-3 text-right text-red-600">{m.failed}</td>
                  <td className="py-3 text-right">{Math.round(m.avg_latency)}ms</td>
                  <td className="py-3 text-right">{m.total_tokens}</td>
                </tr>
              ))}
              {modelStats.length === 0 && (
                <tr><td colSpan={6} className="py-8 text-center text-muted-foreground">暂无数据</td></tr>
              )}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  )
}

function StatCard({ title, value, icon }: { title: string; value: any; icon: React.ReactNode }) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
      </CardContent>
    </Card>
  )
}
