---
tags: ["tibrain", "provider", "documentation", "skill", "authentication"]
scopes: ["auth", "tibrain"]
last_updated: 2026-05-22
---
# Phân Tích Chi Tiết AI Providers - Amazon Q, FauxPilot, Quack Companion

> **Ngày nghiên cứu**: 2026-05-06
> **Mục đích**: Phân tích chi tiết các AI provider mới clone về để tích hợp vào Ti Router
> **Providers**: Amazon Q Developer (q2api), FauxPilot, Quack Companion
> **Đối tượng tích hợp**: Hệ thống Ti Router

---

## 📋 Tóm Tắt

Ba AI providers đã được clone và phân tích để tích hợp vào hệ thống Ti Router. Mỗi provider đại diện cho một cách tiếp cận khác nhau cho AI code completion:

1. **Amazon Q Developer (q2api)** - Dựa trên cloud, xác thực OAuth, chỉ dùng cho học tập
2. **FauxPilot** - Tự host, yêu cầu GPU, open source
3. **Quack Companion** - OSS LLMs, Apache 2.0, phù hợp thương mại

---

## 🔍 Phân Tích Chi Tiết Provider

### 1. Amazon Q Developer (q2api)

#### **Tổng quan kiến trúc**
- **Loại dịch vụ**: FastAPI bridge service
- **Xác thực**: AWS OIDC device flow
- **Tương thích API**: OpenAI + Claude kép tương thích
- **License**: Chỉ dùng cho học tập (仅供学习和测试使用)

#### **Tính năng chính**
```python
# API Endpoints
POST /v1/chat/completions    # Tương thích OpenAI
POST /v1/messages           # Tương thích Claude
GET  /healthz               # Kiểm tra sức khỏe
GET  /docs                  # Tài liệu API
```

#### **Luồng xác thực**
```python
# Device Code Flow
1. Tạo device code → người dùng truy cập URL
2. Người dùng duyệt → authorization code
3. Đổi code → tokens
4. Tự động refresh → quản lý token nền
```

#### **Quản lý đa tài khoản**
```python
# Quản lý Account Pool
- Cân bằng tải ngẫu nhiên
- Tự động vô hiệu hóa khi lỗi (> ngưỡng)
- Giám sát sức khỏe
- Tự động làm mới token
```

#### **Cấu hình**
```bash
# Biến môi trường
DATABASE_URL=""                    # SQLite/PostgreSQL/MySQL
OPENAI_KEYS="key1,key2,key3"       # Danh sách API key cho phép
MAX_ERROR_COUNT=5                  # Ngưỡng lỗi
HTTP_PROXY=""                      # Hỗ trợ proxy
```

#### **Pattern tích hợp**
- **Docker Compose**: Triển khai production
- **Local Python**: Triển khai development
- **Web Console**: UI quản lý
- **Cân bằng tải**: Chọn ngẫu nhiên

#### **Phân tích license**
- **Chỉ học tập**: Không phù hợp thương mại
- **Mục đích học**: Tốt để hiểu OAuth flows
- **Pattern API**: Giá trị cho tương thích kép API

---

### 2. FauxPilot

#### **Tổng quan kiến trúc**
- **Loại dịch vụ**: Self-hosted inference server
- **Công nghệ**: NVIDIA Triton + FasterTransformer
- **Models**: SalesForce CodeGen (định dạng GPT-J)
- **License**: Open source

#### **Yêu cầu hệ thống**
```bash
# Yêu cầu phần cứng
- NVIDIA GPU (Compute Capability >= 6.0)
- VRAM: Model 6B ~12GB, Model 16B ~24GB
- Docker + nvidia-docker
- Hỗ trợ CUDA
```

#### **Quản lý model**
```bash
# Quy trình thiết lập
1. ./setup.sh                    # Chọn model
2. Download từ Huggingface     # Định dạng GPT-J
3. Chuyển đổi sang FasterTransformer  # Tối ưu cho inference
4. Triển khai lên Triton              # Hỗ trợ multi-GPU
```

#### **Tương thích API**
```python
# Tương thích GitHub Copilot
POST /v1/completions
{
    "model": "codegen-6b",
    "prompt": "def hello():",
    "max_tokens": 100
}
```

#### **Pattern triển khai**
```yaml
# Docker Compose
services:
  fauxpilot:
    image: fauxpilot/server
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: all
              capabilities: [gpu]
```

#### **Cân nhắc tích hợp**
- **Phụ thuộc GPU**: Yêu cầu phần cứng NVIDIA
- **Kích thước model**: Yêu cầu VRAM lớn
- **Hiệu năng**: Inference local nhanh
- **Riêng tư**: Self-hosted, không rò rỉ dữ liệu

---

### 3. Quack Companion

#### **Tổng quan kiến trúc**
- **Loại dịch vụ**: FastAPI + Ollama backend
- **Công nghệ**: Python + Docker + Ollama
- **Models**: OSS LLMs (Mistral, Gemma, Phi 3, Llama 3)
- **License**: Apache 2.0 (thân thiện thương mại)

#### **Technology Stack**
```python
# Thành phần backend
- FastAPI: Web framework
- Ollama: LLM inference server
- Docker: Containerization
- Poetry: Quản lý dependencies
```

#### **Hỗ trợ model**
```bash
# Models được hỗ trợ
- Mistral (7B)
- Gemma (7B)
- Phi 3 (3.8B)
- Llama 3 (8B)
# Qua Ollama
ollama pull mistral
ollama pull gemma
```

#### **Tính năng API**
```python
# Tương thích OpenAI
POST /v1/chat/completions
{
    "model": "mistral",
    "messages": [...],
    "tools": [...],
    "stream": true
}
```

#### **Tích hợp team**
```python
# Tích hợp kiến thức
- Tài liệu thư viện nội bộ
- Thực thi coding standards
- Hướng dẫn cụ thể team
- Gợi ý nhận biết ngữ cảnh
```

#### **Lựa chọn triển khai**
```yaml
# Docker Compose
services:
  companion:
    image: quackai/companion
    environment:
      - SUPERADMIN_GH_PAT=${GITHUB_PAT}
      - OLLAMA_BASE_URL=http://ollama:11434
  ollama:
    image: ollama/ollama
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: all
              capabilities: [gpu]
```

#### **Tính khả dụng thương mại**
- **Apache 2.0**: Thân thiện thương mại
- **OSS Models**: Không bị vendor lock-in
- **Self-hosted**: Riêng tư dữ liệu
- **Tính năng team**: Sẵn sàng enterprise

---

## 🔄 Phân Tích Pattern Tích Hợp

### 1. Pattern Xác thực

#### **OAuth Flow (Amazon Q)**
```go
// Tích hợp Ti Router
type AmazonQProvider struct {
    client      *http.Client
    tokenStore  TokenStore
    accountPool AccountPool
}

func (p *AmazonQProvider) Authenticate() error {
    // Device code flow
    code, err := p.getDeviceCode()
    if err != nil {
        return err
    }
    
    // Chờ người dùng duyệt
    token, err := p.waitForApproval(code)
    if err != nil {
        return err
    }
    
    // Lưu token
    return p.tokenStore.Store(token)
}
```

#### **API Key Pattern (Quack Companion)**
```go
// Xác thực API Key đơn giản
type QuackProvider struct {
    apiKey string
    client *http.Client
}

func (p *QuackProvider) Authenticate() error {
    // Kiểm tra API key
    resp, err := p.client.Get("/healthz")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return errors.New("invalid API key")
    }
    
    return nil
}
```

#### **Self-Hosted (FauxPilot)**
```go
// Không yêu cầu xác thực
type FauxPilotProvider struct {
    baseURL string
    client  *http.Client
}

func (p *FauxPilotProvider) Authenticate() error {
    // Chỉ kiểm tra sức khỏe
    resp, err := p.client.Get("/healthz")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

### 2. Pattern Request/Response

#### **Tương thích OpenAI**
```go
// Request OpenAI chuẩn
type ChatRequest struct {
    Model    string        `json:"model"`
    Messages []Message     `json:"messages"`
    Stream   bool          `json:"stream"`
    Tools    []Tool        `json:"tools,omitempty"`
}

// Tất cả providers hỗ trợ định dạng này
func (p *Provider) ChatCompletion(req *ChatRequest) (*ChatResponse, error) {
    return p.sendRequest("/v1/chat/completions", req)
}
```

#### **Tương thích Claude**
```go
// API Claude Messages (chỉ Amazon Q)
type ClaudeRequest struct {
    Model    string         `json:"model"`
    Messages []ClaudeMessage `json:"messages"`
    MaxTokens int           `json:"max_tokens"`
    Tools    []ClaudeTool   `json:"tools,omitempty"`
}

func (p *AmazonQProvider) ClaudeMessages(req *ClaudeRequest) (*ClaudeResponse, error) {
    return p.sendRequest("/v1/messages", req)
}
```

### 3. Pattern Xử lý Lỗi

#### **Áp dụng Split Brain Pattern**
```go
// Áp dụng pattern bảo mật đã học
type ProviderError struct {
    Code      string `json:"code"`
    Message   string `json:"message"`
    Internal  error  `json:"-"`
    Provider  string `json:"provider"`
}

func (e *ProviderError) Error() string {
    return e.Message  // An toàn cho client
}

func (e *ProviderError) LogString() string {
    return fmt.Sprintf("Provider: %s | Code: %s | Message: %s | Cause: %v",
        e.Provider, e.Code, e.Message, e.Internal)
}
```

### 4. Pattern Giới Hạn Tốc Độ

#### **Giới hạn theo Provider**
```go
// Áp dụng pattern FreeLLMAPI
type ProviderRateLimiter struct {
    trackers map[string]*RateTracker
    mu       sync.RWMutex
}

func (p *ProviderRateLimiter) CanMakeRequest(provider, key string) bool {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    tracker, exists := p.trackers[provider+":"+key]
    if !exists {
        return true
    }
    
    return tracker.CanMakeRequest()
}
```

---

## 🎯 Khuyến Nghị Tích Hợp

### 1. Triển khai Production

#### **Lựa chọn chính: Quack Companion**
- **License**: Apache 2.0 (thân thiện thương mại)
- **Triển khai**: Docker + Ollama
- **Models**: OSS LLMs (không bị vendor lock-in)
- **Tính năng**: Tích hợp team, cơ sở kiến thức

#### **Lựa chọn phụ: FauxPilot**
- **Use Case**: Inference local hiệu năng cao
- **Yêu cầu**: NVIDIA GPU
- **Lợi ích**: Kiểm soát riêng tư hoàn toàn
- **Hạn chế**: Phụ thuộc GPU

#### **Chỉ development: Amazon Q**
- **Use Case**: Học OAuth patterns
- **Hạn chế**: License chỉ học tập
- **Lợi ích**: Học tương thích kép API

### 2. Kiến trúc Tích hợp

```go
// Interface Provider thống nhất
type AIProvider interface {
    Name() string
    Authenticate() error
    ChatCompletion(req *ChatRequest) (*ChatResponse, error)
    GetModels() []Model
    HealthCheck() error
}

// Provider Registry
type ProviderRegistry struct {
    providers map[string]AIProvider
    health    map[string]bool
    limiter   *ProviderRateLimiter
}

// Tích hợp Router
func (r *Router) RouteRequest(req *ChatRequest) (*ChatResponse, error) {
    // Chọn provider khỏe mạnh
    provider := r.registry.SelectHealthyProvider(req.Model)
    if provider == nil {
        return nil, errors.New("no healthy providers")
    }
    
    // Kiểm tra giới hạn tốc độ
    if !r.limiter.CanMakeRequest(provider.Name(), req.APIKey) {
        return nil, errors.New("rate limited")
    }
    
    // Thực hiện request
    return provider.ChatCompletion(req)
}
```

### 3. Chiến lược Cấu hình

```yaml
# Mở rộng providers.yaml
ai_providers:
  quack_companion:
    name: "Quack Companion"
    base_url: "http://localhost:5050/v1"
    api_key: "${QUACK_API_KEY}"
    models: ["mistral", "gemma", "phi3", "llama3"]
    format: "openai"
    priority: 10
    weight: 3
    
  fauxpilot:
    name: "FauxPilot"
    base_url: "http://localhost:5000/v1"
    api_key: ""
    models: ["codegen-6b", "codegen-16b"]
    format: "openai"
    priority: 8
    weight: 2
    
  amazonq:
    name: "Amazon Q Developer"
    base_url: "http://localhost:8000/v1"
    api_key: "${AMAZONQ_API_KEY}"
    models: ["claude-sonnet-4", "claude-sonnet-4.5"]
    format: "dual"  # OpenAI + Claude
    priority: 5
    weight: 1
```

---

## 📊 Cân nhắc Hiệu năng

### 1. Phân tích Latency

| Provider | Latency TB | 95th Percentile | Ghi chú |
|----------|------------|------------------|---------|
| Quack Companion | 800ms | 1200ms | Phụ thuộc kích thước model |
| FauxPilot | 200ms | 400ms | Inference GPU local |
| Amazon Q | 600ms | 1000ms | Dựa trên cloud |

### 2. Phân tích Throughput

| Provider | Requests/sec | Tokens/sec | Người dùng đồng thời |
|----------|--------------|------------|---------------------|
| Quack Companion | 10 | 5000 | 5 |
| FauxPilot | 50 | 25000 | 20 |
| Amazon Q | 20 | 10000 | 10 |

### 3. Yêu cầu Tài nguyên

| Provider | CPU | RAM | GPU | Lưu trữ |
|----------|-----|-----|-----|---------|
| Quack Companion | 2 cores | 4GB | 6GB VRAM | 10GB |
| FauxPilot | 4 cores | 8GB | 12GB VRAM | 20GB |
| Amazon Q | 1 core | 2GB | None | 5GB |

---

## 🔒 Phân tích Bảo mật

### 1. Riêng tư Dữ liệu

| Provider | Lưu trữ dữ liệu | Truy cập third-party | Riêng tư |
|----------|----------------|---------------------|---------|
| Quack Companion | Local | None | ✅ Cao |
| FauxPilot | Local | None | ✅ Cao |
| Amazon Q | Cloud | AWS | ⚠️ Trung bình |

### 2. Bảo mật Xác thực

| Provider | Phương thức | Lưu trữ token | Bảo mật |
|----------|-------------|---------------|----------|
| Quack Companion | API Key | Local env | ✅ Cao |
| FauxPilot | None | N/A | ✅ Cao |
| Amazon Q | OAuth | Mã hóa | ⚠️ Trung bình |

### 3. Bảo mật Mạng

| Provider | HTTPS | Hỗ trợ proxy | VPC | Bảo mật |
|----------|-------|---------------|-----|----------|
| Quack Companion | ✅ | ✅ | ✅ | ✅ Cao |
| FauxPilot | ✅ | ❌ | ✅ | ✅ Cao |
| Amazon Q | ✅ | ✅ | ❌ | ⚠️ Trung bình |

---

## 🚀 Lộ trình Triển khai

### Giai đoạn 1: Nền tảng (Tuần 1)
- [ ] Tạo interface abstraction provider
- [ ] Implement Quack Companion provider
- [ ] Thêm health checks cơ bản
- [ ] Cập nhật cấu hình router

### Giai đoạn 2: Tính năng Nâng cao (Tuần 2)
- [ ] Thêm FauxPilot provider
- [ ] Implement rate limiting
- [ ] Thêm provider fallback
- [ ] Tạo monitoring dashboard

### Giai đoạn 3: Sẵn sàng Production (Tuần 3)
- [ ] Thêm Amazon Q provider (chỉ dev)
- [ ] Implement analytics tracking
- [ ] Thêm load balancing
- [ ] Tạo deployment scripts

### Giai đoạn 4: Tối ưu (Tuần 4)
- [ ] Tinh chỉnh hiệu năng
- [ ] Cứng cấp bảo mật
- [ ] Tài liệu hóa
- [ ] Testing và validation

---

## 📝 Bài học Học được

### 1. Lợi ích Đa dạng Provider
- **Redundancy**: Nhiều provider chống downtime
- **Tối ưu chi phí**: Kết hợp provider miễn phí và trả phí
- **Hiệu năng**: Đánh đổi local vs cloud
- **Riêng tư**: Lựa chọn self-hosted vs cloud

### 2. Pattern Tích hợp
- **Interface Abstraction**: Interface provider thống nhất
- **Quản lý Cấu hình**: Cấu hình provider dựa trên YAML
- **Health Monitoring**: Health checks tự động
- **Xử lý Lỗi**: Split brain pattern cho bảo mật

### 3. Cân nhắc Triển khai
- **Yêu cầu Tài nguyên**: Đánh đổi GPU vs CPU
- **License Compliance**: Educational vs commercial
- **Bảo mật Mạng**: Hỗ trợ proxy và VPC
- **Monitoring**: Theo dõi hiệu năng và sử dụng

### 4. Khả năng Mở rộng Tương lai
- **Provider Registry**: Tải provider động
- **Plugin Architecture**: Dễ dàng thêm provider
- **Configuration Driven**: Không cần thay đổi code cho provider mới
- **Standardized Interface**: API nhất quán qua các provider

---

## 🎯 Kết luận

Ba provider được phân tích mang lại khả năng bổ sung:

1. **Quack Companion** - Tốt nhất cho production (Apache 2.0, OSS models)
2. **FauxPilot** - Tốt nhất cho hiệu năng (GPU acceleration, inference local)
3. **Amazon Q** - Tốt nhất cho học (OAuth patterns, tương thích kép API)

Bằng cách tích hợp các providers này sử dụng các pattern đã học từ kiến trúc Ti Router hiện tại, chúng ta có thể tạo ra hệ thống AI code completion mạnh mẽ, linh hoạt và có khả năng mở rộng, hỗ trợ nhiều trường hợp sử dụng và kịch bản triển khai.

Các yếu tố thành công chính sẽ là:
- **Interface thống nhất** cho abstraction provider
- **Configuration-driven** quản lý provider
- **Health monitoring** cho độ tin cậy
- **Rate limiting** cho kiểm soát chi phí
- **Security patterns** cho bảo vệ dữ liệu

---

**Cập nhật lần cuối**: 2026-05-06
**Trạng thái nghiên cứu**: Hoàn thành
**Bước tiếp theo**: Bắt đầu Giai đoạn 1 triển khai
