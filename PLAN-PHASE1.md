# LLM智能路由系统 — 阶段一详细计划书

> 📌 定位：个人本地使用的轻量级LLM路由管理中心
> 🎯 阶段一目标：路由"跑通"，具备统一接入、转发、记录和可视化的基础能力

---

## 一、项目架构设计

### 1.1 技术选型

| 组件 | 选型 | 理由 |
|------|------|------|
| 后端框架 | Go (Gin/Fiber) | 轻量高性能，适合网关类服务 |
| 数据存储 | SQLite | 零依赖，本地单文件数据库，个人使用绰绰有余 |
| 配置管理 | YAML + 热加载 | 人类可读，支持运行时修改 |
| 前端管理 | React + Vite + shadcn/ui + Tailwind CSS | 现代化组件库，主题定制灵活 |
| 日志存储 | SQLite（结构化日志） | 统一存储，方便查询 |

### 1.2 系统架构图

```
┌─────────────────────────────────────────────────┐
│                 调用方 (Applications)              │
│         curl / Python / Node.js / 浏览器          │
└──────────────────────┬──────────────────────────┘
                       │ OpenAI 兼容 API
                       ▼
┌─────────────────────────────────────────────────┐
│              LLM Router (Go 后端)                 │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐ │
│  │ 请求接收  │→│ 路由决策  │→│  模型转发      │ │
│  │ (认证)   │  │ (别名解析)│  │ (HTTP Proxy)  │ │
│  └──────────┘  └──────────┘  └───────┬───────┘ │
│       │                              │          │
│       ▼                              ▼          │
│  ┌──────────┐               ┌───────────────┐  │
│  │ 日志记录  │               │ 流式SSE透传   │  │
│  └──────────┘               └───────────────┘  │
└──────────────────────┬──────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
   ┌─────────┐   ┌─────────┐   ┌─────────┐
   │DeepSeek │   │  Mimo   │   │ MiniMax │
   │   API   │   │   API   │   │   API   │
   └─────────┘   └─────────┘   └─────────┘
```

---

## 二、数据模型设计

### 2.1 端点表 (endpoints)

```sql
CREATE TABLE endpoints (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,          -- 端点名称，如 "deepseek"
    api_base    TEXT NOT NULL,                 -- API基础URL
    api_key     TEXT NOT NULL,                 -- API密钥
    is_active   BOOLEAN DEFAULT 1,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 2.2 模型表 (models)

```sql
CREATE TABLE models (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,          -- 模型别名，如 "deepseek-flash"
    endpoint_id INTEGER NOT NULL,              -- 关联端点
    model_id    TEXT NOT NULL,                 -- 真实模型ID
    discovered  BOOLEAN DEFAULT 0,            -- 是否自动发现
    is_active   BOOLEAN DEFAULT 1,
    max_retries INTEGER DEFAULT 2,
    fallback    TEXT,                          -- 备用模型别名
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (endpoint_id) REFERENCES endpoints(id)
);
```

### 2.3 设计理念

**端点与模型分离**：端点存储 api_base + api_key，模型通过 endpoint_id 引用端点。启动时自动调用 `/v1/models` 发现可用模型（discovered=true），config.yaml 中的显式配置作为路由规则（别名、降级链）。

### 2.2 请求日志表 (logs)

```sql
CREATE TABLE logs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id      TEXT,                      -- 会话ID
    model_name      TEXT NOT NULL,             -- 实际使用的模型
    alias_name      TEXT,                      -- 调用时使用的别名
    request_tokens  INTEGER,                   -- 请求Token数
    response_tokens INTEGER,                   -- 响应Token数
    total_tokens    INTEGER,                   -- 总Token数
    latency_ms      INTEGER,                   -- 响应耗时(ms)
    status          TEXT NOT NULL,             -- success / failed / retried
    error_message   TEXT,                      -- 失败原因
    request_summary TEXT,                      -- 请求摘要（简短）
    response_summary TEXT,                     -- 响应摘要（简短）
    request_body    TEXT,                      -- 完整请求内容（可选）
    response_body   TEXT,                      -- 完整响应内容（可选）
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_logs_session ON logs(session_id);
CREATE INDEX idx_logs_model ON logs(model_name);
CREATE INDEX idx_logs_created ON logs(created_at);
```

---

## 三、模块一：大模型认证管理

### 3.1 功能清单

| 功能 | 描述 | 优先级 |
|------|------|--------|
| 模型配置CRUD | 增删改查模型配置（别名、厂商、API Key等） | P0 |
| 别名解析 | 将调用方的模型名/别名解析为真实的厂商+模型ID | P0 |
| API Key管理 | 支持查看、更新、删除API Key | P0 |
| 配置热加载 | 修改配置后无需重启服务 | P1 |

### 3.2 数据结构定义

```go
// config/models.go
type ModelConfig struct {
    ID          int64  `json:"id" gorm:"primaryKey"`
    Name        string `json:"name" gorm:"uniqueIndex"`     // 别名
    Provider    string `json:"provider"`                      // 厂商
    ModelID     string `json:"model_id"`                      // 真实模型名
    APIBase     string `json:"api_base"`                      // API基础URL
    APIKey      string `json:"api_key"`                       // 密钥
    IsActive    bool   `json:"is_active"`                     // 是否启用
    MaxRetries  int    `json:"max_retries"`                   // 最大重试次数
    Fallback    string `json:"fallback"`                      // 备用模型
}
```

### 3.3 配置文件格式 (config.yaml)

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  admin_port: 8081

storage:
  db_path: "./data/llm-router.db"
  log_full_content: false

# 端点配置 — 填 api_base 和 api_key，启动时自动发现可用模型
endpoints:
  - name: "deepseek"
    api_base: "https://api.deepseek.com/v1"
    api_key: "sk-xxx"

  - name: "xiaomi"
    api_base: "https://api.mimo.xiaomi.com/v1"
    api_key: "sk-xxx"

  - name: "minimax"
    api_base: "https://api.minimaxi.com/v1"
    api_key: "sk-xxx"

# 模型路由配置 — 引用端点名，指定别名和降级链
models:
  - name: "default"
    endpoint: "minimax"
    model_id: "minimax-2.7"
    is_active: true
    max_retries: 2
    fallback: "deepseek-flash"

  - name: "deepseek-flash"
    endpoint: "deepseek"
    model_id: "deepseek-v4-flash"
    is_active: true
    max_retries: 2
    fallback: "deepseek-pro"

  - name: "deepseek-pro"
    endpoint: "deepseek"
    model_id: "deepseek-v4-pro"
    is_active: true
    max_retries: 1
    fallback: "mimo"

  - name: "mimo"
    endpoint: "xiaomi"
    model_id: "mimo-2.5"
    is_active: true
    max_retries: 1
    fallback: "default"

headers:
  model_override: "X-Msf-Model"
  session_id: "X-Msf-Session-Id"
```

### 3.4 别名解析逻辑

```
收到请求 model: "cheap"
    ↓
查找 models 表 WHERE name = "cheap" AND is_active = 1
    ↓
返回: { provider: "minimax", model_id: "minimax-2.7", api_base: "https://api.minimaxi.com/v1", api_key: "..." }
    ↓
如果找不到 → 返回 404 错误，附带可用模型列表
```

### 3.5 API 端点设计

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /admin/api/models | 获取所有模型列表 |
| POST | /admin/api/models | 创建新模型配置 |
| PUT | /admin/api/models/:id | 更新模型配置 |
| DELETE | /admin/api/models/:id | 删除模型配置 |
| PUT | /admin/api/models/:id/toggle | 启用/禁用模型 |
| POST | /admin/api/models/:id/test | 测试模型连通性 |

---

## 四、模块二：路由分发

### 4.1 功能清单

| 功能 | 描述 | 优先级 |
|------|------|--------|
| OpenAI兼容接口 | `/v1/chat/completions` 完全兼容 | P0 |
| 请求转发 | 根据模型名/别名转发到真实LLM | P0 |
| 流式SSE透传 | `stream: true` 时实时透传后端SSE流 | P0 |
| 接口重试 | 失败时自动重试1-2次 | P0 |
| 降级处理 | 重试失败后切换到备用模型 | P0 |
| 自定义请求头覆盖 | 通过 `X-Msf-Model` 动态切换模型 | P0 |
| Token统计 | 从响应中提取usage字段记录Token消耗 | P1 |

### 4.2 请求处理流程

```
客户端请求 → POST /v1/chat/completions
    ↓
1. 解析请求体，提取 model 字段
    ↓
2. 检查 X-Msf-Model 请求头
   如果存在 → 覆盖 model 字段
    ↓
3. 提取/生成 X-Msf-Session-Id
   如果请求头无此字段 → 生成 UUID 并记录
    ↓
4. 别名解析：将 model 名解析为真实模型配置
    ↓
5. 记录请求开始时间
    ↓
6. 构造转发请求（保持原始请求体不变）
   替换: api_base + /chat/completions
   替换: Authorization: Bearer <api_key>
    ↓
7. 发送请求（支持超时控制，默认30s）
    ↓
8. 成功 → 记录日志，透传响应给客户端
   失败 → 进入重试逻辑
    ↓
9. 重试逻辑：
   a. 检查重试次数 < max_retries
   b. 等待指数退避（1s, 2s）
   c. 重新执行步骤6-8
   d. 如果仍然失败，检查是否有 fallback 模型
   e. 有 fallback → 切换模型，重新执行步骤4-8
   f. 无 fallback → 返回错误响应
    ↓
10. 记录最终结果日志
```

### 4.3 流式SSE透传实现要点

```go
// 关键：不缓冲，逐chunk透传
func streamResponse(w http.ResponseWriter, resp *http.Response) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    scanner := bufio.NewScanner(resp.Body)
    flusher := w.(http.Flusher)
    
    for scanner.Scan() {
        line := scanner.Text()
        fmt.Fprintf(w, "%s\n", line)
        flusher.Flush()
        
        // 记录最后一个chunk的usage信息（如果有的话）
        if strings.HasPrefix(line, "data: ") {
            // 解析并提取token usage
        }
    }
}
```

### 4.4 错误处理与降级

```go
// 降级配置示例
type FallbackChain struct {
    Primary  ModelConfig
    Fallback *ModelConfig   // 第一层降级
    Retry    int            // 当前模型重试次数
}

// 降级策略：
// 1. 主模型重试 max_retries 次（指数退避）
// 2. 如果有 fallback 配置，切换到备用模型
// 3. 备用模型也重试 max_retries 次
// 4. 全部失败返回 502 Bad Gateway + 详细错误信息
```

### 4.5 API 端点设计

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /v1/chat/completions | 主路由端点（OpenAI兼容） |
| GET | /v1/models | 返回所有可用模型列表（OpenAI兼容格式） |
| GET | /health | 健康检查 |

### 4.6 请求头约定

| 请求头 | 说明 | 示例 |
|--------|------|------|
| `X-Msf-Model` | 动态覆盖模型（可选） | `cheap` 或 `mimo` |
| `X-Msf-Session-Id` | 会话ID关联（可选） | `session-abc-123` |
| `Authorization` | 透传给后端（可选，如不填则用配置的key） | `Bearer sk-xxx` |

---

## 五、模块三：通信日志记录

### 5.1 功能清单

| 功能 | 描述 | 优先级 |
|------|------|--------|
| 摘要记录 | 时间、模型、Token、耗时、状态 | P0 |
| 会话ID关联 | 支持手动指定或自动生成会话ID | P0 |
| 完整内容记录 | 可选开关，调试时开启 | P1 |
| 日志查询API | 按时间/模型/会话ID筛选 | P1 |

### 5.2 日志记录流程

```
请求开始
    ↓
记录: request_id, session_id, model_name, request_time, request_summary
    ↓
请求完成/失败
    ↓
记录: latency_ms, status, total_tokens, response_summary, error_message
    ↓
写入 SQLite
```

### 5.3 日志查询 API

| 方法 | 路径 | 参数 | 描述 |
|------|------|------|------|
| GET | /admin/api/logs | page, limit | 分页获取日志 |
| GET | /admin/api/logs | session_id | 按会话筛选 |
| GET | /admin/api/logs | model_name | 按模型筛选 |
| GET | /admin/api/logs | start_time, end_time | 按时间范围 |
| GET | /admin/api/logs | status | 按状态筛选 |
| GET | /admin/api/logs/:id | - | 获取单条完整日志 |
| GET | /admin/api/sessions | - | 获取所有会话列表 |
| GET | /admin/api/sessions/:session_id | - | 获取会话的完整交互链 |

---

## 六、模块四：管理页面

### 6.1 功能清单

| 功能 | 描述 | 优先级 |
|------|------|--------|
| 全局看板 | 今日/本周统计概览 | P0 |
| 模型管理 | 可视化增删改查模型配置 | P0 |
| 日志列表 | 分页、筛选、查看详细日志 | P0 |
| 模型流量统计 | 按模型维度的调用统计图表 | P1 |
| 会话追踪 | 按会话ID查看完整交互链 | P1 |

### 6.2 页面结构

```
管理页面 (React SPA)
├── 看板页 (Dashboard)
│   ├── 今日请求总量 / 成功失败比
│   ├── 平均延迟 / 总Token消耗
│   ├── 预估费用（基于各模型价格）
│   └── 实时请求流（可选）
│
├── 模型管理页 (Models)
│   ├── 模型列表（表格）
│   ├── 新增/编辑模型（对话框）
│   ├── 启用/禁用开关
│   └── 测试连接按钮
│
├── 日志页 (Logs)
│   ├── 筛选栏（时间、模型、会话ID、状态）
│   ├── 日志列表（分页表格）
│   ├── 日志详情（点击查看完整内容）
│   └── 会话视图（按会话ID聚合）
│
└── 统计页 (Stats)
    ├── 模型调用次数饼图
    ├── Token消耗柱状图
    └── 延迟趋势折线图
```

### 6.3 看板数据 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /admin/api/stats/overview | 总览统计（今日/本周） |
| GET | /admin/api/stats/models | 按模型维度统计 |
| GET | /admin/api/stats/timeseries | 时间序列数据（用于图表） |

---

## 七、项目目录结构

```
llm-router/
├── cmd/
│   └── server/
│       └── main.go              # 入口文件
├── internal/
│   ├── config/
│   │   ├── config.go            # 配置加载
│   │   └── models.go            # 数据模型定义
│   ├── handler/
│   │   ├── proxy.go             # /v1/chat/completions 处理
│   │   ├── admin.go             # 管理API处理
│   │   └── health.go            # 健康检查
│   ├── router/
│   │   ├── resolver.go          # 别名解析
│   │   ├── proxy.go             # 请求转发
│   │   └── fallback.go          # 重试与降级
│   ├── storage/
│   │   ├── db.go                # SQLite初始化
│   │   ├── models.go            # 模型CRUD
│   │   └── logs.go              # 日志读写
│   └── middleware/
│       ├── cors.go              # CORS中间件
│       └── logging.go           # 请求日志中间件
├── web/                         # 前端管理页面
│   ├── src/
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx    # 看板页
│   │   │   ├── Models.tsx       # 模型管理
│   │   │   ├── Logs.tsx         # 日志页
│   │   │   └── Stats.tsx        # 统计页
│   │   ├── api/
│   │   │   └── admin.ts         # 管理API封装
│   │   └── App.tsx
│   ├── index.html
│   └── package.json
├── data/                        # 数据目录（git忽略）
├── config.yaml                  # 配置文件（示例）
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## 八、实施步骤（任务拆解）

### 第一步：项目骨架搭建
- [x] 初始化 Go module
- [x] 搭建项目目录结构
- [x] 实现配置加载（YAML）
- [x] 搭建 Gin HTTP 服务器框架

### 第二步：数据层实现
- [x] SQLite 初始化与表创建
- [x] 模型配置 CRUD（models 表）
- [x] 日志写入与查询（logs 表）

### 第三步：核心路由转发
- [x] 别名解析器实现（从数据库查询）
- [x] 普通请求转发（POST → LLM → 响应）
- [x] 流式 SSE 透传（逐 chunk 刷新）
- [ ] 自定义请求头覆盖（X-Msf-Model）— 延后，等日志系统完成后设计
- [ ] 会话ID提取/生成（X-Msf-Session-Id）— 延后，等日志系统完成后设计

### 第四步：重试与降级
- [x] 指数退避重试逻辑（1s, 2s, 4s...）
- [x] 备用模型切换（降级链）
- [x] 错误分类与处理（4xx 不重试，5xx 重试）

### 第五步：日志系统
- [x] 请求/响应摘要记录（自动提取最后一条消息）
- [x] 完整内容记录开关（log_full_content 配置）
- [x] 日志查询 API（分页、按模型/状态/会话筛选）
- [x] 日志详情 API（包含完整请求/响应内容）

### 第六步：管理API
- [x] 端点管理 API（CRUD）
- [x] 模型管理 API（CRUD + 启用/禁用切换）
- [x] 模型自动发现 API（触发端点扫描）
- [x] 统计数据 API（概览 / 模型维度 / 时间序列）
- [x] 健康检查端点
- [x] 会话列表 API

### 第六步附：端点发现机制
- [x] 端点表 + 模型表分离（端点存 key，模型引用端点）
- [x] 启动时自动调用 /v1/models 发现可用模型
- [x] 发现的模型标记 discovered=true，默认不启用

### 第七步：前端管理页面
- [ ] 项目初始化（Vite + React + shadcn/ui + Tailwind CSS）
- [ ] 看板页
- [ ] 模型管理页
- [ ] 日志页
- [ ] 统计页（图表）

### 第八步：集成测试
- [ ] 使用 curl 测试各端点
- [ ] 流式响应测试
- [ ] 重试/降级测试
- [ ] 配置热加载测试

---

## 九、验收标准

| 指标 | 标准 |
|------|------|
| 启动时间 | `go run .` 一键启动，2秒内就绪 |
| 转发延迟 | 非流式请求额外开销 < 50ms |
| 流式透传 | 逐chunk透传，无缓冲，首字节延迟 < 100ms |
| 重试可靠性 | 重试间隔指数退避，降级切换 < 5s |
| 日志完整性 | 每次请求必有记录，字段完整 |
| 管理页面 | 各页面功能可用，数据实时刷新 |
| API兼容性 | 可直接替换 OpenAI 端点，现有应用无需修改 |

---

## 十、风险与注意事项

1. **API Key 明文存储**：阶段一为本地个人使用，明文存储在 SQLite 中。后续阶段需引入加密。
2. **流式透传复杂度**：SSE 透传是本阶段的核心难点，需处理超时、断连、部分响应等边界情况。
3. **Token统计准确性**：不同厂商的 usage 字段格式可能不同，需要逐个适配。
4. **并发安全**：SQLite 的写入是串行的，高并发下可能成为瓶颈。个人使用场景问题不大。
5. **错误信息透传**：降级时需要保留原始错误信息，方便排查问题。

---

*计划书版本：v1.0*
*最后更新：2026-05-08*
