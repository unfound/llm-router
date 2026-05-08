package router

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/unfound/llm-router/internal/storage"
)

// ProxyRequest 代理请求
type ProxyRequest struct {
	Model     *storage.ModelWithEndpoint
	Body      []byte
	IsStream  bool
	StartTime time.Time
}

// ProxyResponse 代理响应
type ProxyResponse struct {
	StatusCode     int
	Body           []byte
	ResponseTokens int
	TotalTokens    int
	LatencyMs      int64
	Error          error
}

// Forward 转发请求到目标 LLM
func Forward(w http.ResponseWriter, req *ProxyRequest) *ProxyResponse {
	// 构造目标 URL（从端点信息获取 api_base）
	base := strings.TrimRight(req.Model.APIBase, "/")
	targetURL := base + "/chat/completions"
	if !strings.HasSuffix(base, "/v1") {
		targetURL = base + "/v1/chat/completions"
	}

	// 创建转发请求
	httpReq, err := http.NewRequest("POST", targetURL, strings.NewReader(string(req.Body)))
	if err != nil {
		return &ProxyResponse{Error: err}
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	if req.Model.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.Model.APIKey)
	}

	// 发送请求
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &ProxyResponse{Error: err}
	}
	defer resp.Body.Close()

	latency := time.Since(req.StartTime).Milliseconds()

	// 流式响应
	if req.IsStream {
		return streamResponse(w, resp, latency)
	}

	// 普通响应
	return normalResponse(w, resp, latency)
}

// normalResponse 处理普通（非流式）响应
func normalResponse(w http.ResponseWriter, resp *http.Response, latency int64) *ProxyResponse {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ProxyResponse{StatusCode: resp.StatusCode, Error: err}
	}

	// 解析 usage 信息
	var result struct {
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	json.Unmarshal(body, &result)

	responseTokens := 0
	totalTokens := 0
	if result.Usage != nil {
		responseTokens = result.Usage.CompletionTokens
		totalTokens = result.Usage.TotalTokens
	}

	// 透传响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	return &ProxyResponse{
		StatusCode:     resp.StatusCode,
		Body:           body,
		ResponseTokens: responseTokens,
		TotalTokens:    totalTokens,
		LatencyMs:      latency,
	}
}

// streamResponse 处理流式 SSE 响应
func streamResponse(w http.ResponseWriter, resp *http.Response, latency int64) *ProxyResponse {
	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return &ProxyResponse{Error: fmt.Errorf("不支持流式响应")}
	}

	scanner := bufio.NewScanner(resp.Body)
	// 增大 buffer 以支持大 chunk
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var lastUsage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	}

	for scanner.Scan() {
		line := scanner.Text()

		// 写入并刷新
		fmt.Fprintf(w, "%s\n", line)
		flusher.Flush()

		// 尝试从 data: [DONE] 前的最后一个 chunk 提取 usage
		if strings.HasPrefix(line, "data: ") && line != "data: [DONE]" {
			var chunk struct {
				Usage *struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
					TotalTokens      int `json:"total_tokens"`
				} `json:"usage"`
			}
			if json.Unmarshal([]byte(line[6:]), &chunk) == nil && chunk.Usage != nil {
				lastUsage = chunk.Usage
			}
		}
	}

	responseTokens := 0
	totalTokens := 0
	if lastUsage != nil {
		responseTokens = lastUsage.CompletionTokens
		totalTokens = lastUsage.TotalTokens
	}

	return &ProxyResponse{
		StatusCode:     resp.StatusCode,
		ResponseTokens: responseTokens,
		TotalTokens:    totalTokens,
		LatencyMs:      latency,
	}
}
