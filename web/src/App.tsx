import { BrowserRouter, Routes, Route, NavLink } from 'react-router-dom'
import Dashboard from '@/pages/Dashboard'
import Models from '@/pages/Models'
import Logs from '@/pages/Logs'
import Stats from '@/pages/Stats'
import { LayoutDashboard, Cpu, FileText, BarChart3 } from 'lucide-react'

const navItems = [
  { to: '/', label: '看板', icon: LayoutDashboard },
  { to: '/models', label: '端点 & 模型', icon: Cpu },
  { to: '/logs', label: '日志', icon: FileText },
  { to: '/stats', label: '统计', icon: BarChart3 },
]

export default function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-background flex">
        {/* 侧边栏 */}
        <aside className="w-56 border-r bg-card p-4 flex flex-col gap-1">
          <div className="mb-6 px-3">
            <h1 className="text-lg font-bold">LLM Router</h1>
            <p className="text-xs text-muted-foreground">管理面板</p>
          </div>
          {navItems.map(item => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors ${
                  isActive
                    ? 'bg-primary text-primary-foreground font-medium'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                }`
              }
            >
              <item.icon className="h-4 w-4" />
              {item.label}
            </NavLink>
          ))}
        </aside>

        {/* 主内容 */}
        <main className="flex-1 p-8 overflow-auto">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/models" element={<Models />} />
            <Route path="/logs" element={<Logs />} />
            <Route path="/stats" element={<Stats />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}
