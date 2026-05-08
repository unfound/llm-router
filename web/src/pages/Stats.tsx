import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getStatsTimeSeries, getStatsModels } from '@/api/admin'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, LineChart, Line, CartesianGrid } from 'recharts'

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']

export default function Stats() {
  const [timeSeries, setTimeSeries] = useState<any[]>([])
  const [modelStats, setModelStats] = useState<any[]>([])
  const [hours, setHours] = useState(24)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    Promise.all([getStatsTimeSeries(hours), getStatsModels()])
      .then(([ts, ms]) => {
        setTimeSeries(ts.timeseries || [])
        setModelStats(ms.models || [])
      })
      .finally(() => setLoading(false))
  }, [hours])

  if (loading) return <div className="p-8 text-muted-foreground">加载中...</div>

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">统计</h1>
        <select
          className="border rounded px-3 py-2 text-sm"
          value={hours}
          onChange={e => setHours(Number(e.target.value))}
        >
          <option value={1}>最近 1 小时</option>
          <option value={6}>最近 6 小时</option>
          <option value={24}>最近 24 小时</option>
          <option value={72}>最近 3 天</option>
          <option value={168}>最近 7 天</option>
        </select>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* 请求趋势 */}
        <Card>
          <CardHeader>
            <CardTitle>请求趋势</CardTitle>
          </CardHeader>
          <CardContent>
            {timeSeries.length > 0 ? (
              <ResponsiveContainer width="100%" height={250}>
                <BarChart data={timeSeries}>
                  <XAxis dataKey="hour" tick={{ fontSize: 11 }} tickFormatter={v => v?.slice(11, 16)} />
                  <YAxis tick={{ fontSize: 11 }} />
                  <Tooltip labelFormatter={v => v?.replace('T', ' ')} />
                  <Bar dataKey="success" name="成功" fill="#10b981" />
                  <Bar dataKey="total" name="总请求" fill="#3b82f6" />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-[250px] flex items-center justify-center text-muted-foreground">暂无数据</div>
            )}
          </CardContent>
        </Card>

        {/* Token 趋势 */}
        <Card>
          <CardHeader>
            <CardTitle>Token 消耗趋势</CardTitle>
          </CardHeader>
          <CardContent>
            {timeSeries.length > 0 ? (
              <ResponsiveContainer width="100%" height={250}>
                <LineChart data={timeSeries}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="hour" tick={{ fontSize: 11 }} tickFormatter={v => v?.slice(11, 16)} />
                  <YAxis tick={{ fontSize: 11 }} />
                  <Tooltip labelFormatter={v => v?.replace('T', ' ')} />
                  <Line type="monotone" dataKey="tokens" name="Token" stroke="#f59e0b" strokeWidth={2} />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-[250px] flex items-center justify-center text-muted-foreground">暂无数据</div>
            )}
          </CardContent>
        </Card>

        {/* 模型调用占比 */}
        <Card>
          <CardHeader>
            <CardTitle>模型调用占比</CardTitle>
          </CardHeader>
          <CardContent>
            {modelStats.length > 0 ? (
              <ResponsiveContainer width="100%" height={250}>
                <PieChart>
                  <Pie
                    data={modelStats}
                    dataKey="total"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    outerRadius={80}
                    label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  >
                    {modelStats.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-[250px] flex items-center justify-center text-muted-foreground">暂无数据</div>
            )}
          </CardContent>
        </Card>

        {/* 模型 Token 消耗 */}
        <Card>
          <CardHeader>
            <CardTitle>模型 Token 消耗</CardTitle>
          </CardHeader>
          <CardContent>
            {modelStats.length > 0 ? (
              <ResponsiveContainer width="100%" height={250}>
                <BarChart data={modelStats} layout="vertical">
                  <XAxis type="number" tick={{ fontSize: 11 }} />
                  <YAxis dataKey="name" type="category" tick={{ fontSize: 11 }} width={100} />
                  <Tooltip />
                  <Bar dataKey="total_tokens" name="Token" fill="#8b5cf6" />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-[250px] flex items-center justify-center text-muted-foreground">暂无数据</div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
