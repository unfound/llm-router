# LLM Router

轻量级LLM智能路由系统 — 统一管理多模型API Key，灵活调度，自动省钱。

## 🎯 核心功能

- **统一接入点**：所有应用只需对接路由一个地址
- **模型别名**：为模型设置别名（如 `cheap`），方便调用
- **流式透传**：完整透传后端LLM的SSE流，无缓冲
- **自动重试与降级**：失败时自动重试或切换备用模型
- **完整日志**：记录所有交互，支持会话关联和可视化

## 🚀 快速开始

```bash
# 克隆项目
git clone https://github.com/unfound/llm-router.git
cd llm-router

# 复制配置文件
cp config.example.yaml config.yaml
# 编辑 config.yaml，添加你的 API Key

# 启动
go run .
```

## 📖 配置示例

```yaml
models:
  - name: "cheap"
    provider: "openai"
    model_id: "gpt-4o-mini"
    api_base: "https://api.openai.com/v1"
    api_key: "sk-xxx"
    fallback: "default"
  
  - name: "default"
    provider: "openai"
    model_id: "gpt-4o"
    api_base: "https://api.openai.com/v1"
    api_key: "sk-xxx"
```

## 🔌 使用方式

直接替换 OpenAI 端点即可：

```bash
# 原来
curl https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer sk-xxx" \
  -d '{"model": "gpt-4o", "messages": [{"role": "user", "content": "hello"}]}'

# 改用路由
curl http://localhost:8080/v1/chat/completions \
  -d '{"model": "cheap", "messages": [{"role": "user", "content": "hello"}]}'
```

## 📊 管理页面

访问 `http://localhost:8081` 管理模型配置、查看日志和统计。

## 📋 开发计划

详见 [PLAN-PHASE1.md](PLAN-PHASE1.md)

## 📄 License

MIT
