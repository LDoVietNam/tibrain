---
tags: ["tibrain", "caching", "documentation", "router", "provider"]
scopes: ["resilience", "tibrain"]
last_updated: 2026-05-22
---
# Provider Registry với Caching trong Go - Kinh Nghiệm từ GitHub

## Mục tiêu
Triển khai Provider Registry có cơ chế caching hiệu quả cho Ti Router, dựa trên các best practices từ các dự án open source lớn.

## Các mẫu thiết kế từ GitHub

### 1. HashiCorp Terraform - MemoizeSource
- **Cách hoạt động**: Bọc một Source khác, cache tất cả kết quả (kể cả lỗi)
- **Đặc điểm**:
  - Cache tồn tại suốt vòng đời của object
  - Dùng cho network requests để tái sử dụng kết quả `AvailableVersions`
  - Không có TTL (mặc định tồn tại đến khi restart)

### 2. Operator Framework - Cache với RWMutex
- **Cấu trúc**:
  ```go
  type Cache struct {
      m          sync.RWMutex
      snapshots   map[SourceKey]*snapshotHeader
      sem         chan struct{} // Giới hạn concurrent updates
  }
  ```
- **Đặc điểm**:
  - Dùng `sync.RWMutex` để tối ưu đọc/ghi
  - Có semaphore giới hạn số lượng concurrent snapshot updates
  - Phân biệt theo namespace/source priority

### 3. Boring Registry - Pull-through Cache
- Hỗ trợ multiple storage backends (S3, GCS, Azure Blob)
- Có cơ chế network mirror và pull-through cho providers
- Dùng `patrickmn/go-cache` cho in-memory caching (có TTL và cleanup)

## Áp dụng cho Ti Router Registry

### Hiện trạng (Z:\Ti\router\layers\provider\registry.go)
- ✅ Đã có in-memory cache với `sync.RWMutex`
- ✅ Đã có TTL (mặc định 5 phút cho provider cache, 30 giây cho snapshots)
- ✅ Double-checked locking trong `GetOrCreate` để tránh race condition
- ✅ Cache invalidation khi đăng ký provider mới (`snapshotsCache = nil`)
- ❌ Thiếu: Persistent cache (lưu xuống disk để giữ qua restart)
- ❌ Thiếu: Cache metrics (theo dõi hit/miss rate)
- ❌ Thiếu: Max cache size limit (tránh memory leak)

### Cải tiến đề xuất
1. **Thêm persistent cache**: Sử dụng `bbolt` hoặc `sqlite` để lưu cache provider đã tạo
2. **Thêm cache metrics**: Theo dõi cache hit, miss, eviction count
3. **Thêm max cache size**: Giới hạn số lượng provider tối đa trong cache
4. **Tối ưu snapshot cache**: Dùng goroutine để update snapshot async thay vì block

## Tài liệu tham khảo
- [HashiCorp Terraform MemoizeSource](https://pkg.go.dev/github.com/kubegems/opentofu/pkg/getproviders#MemoizeSource)
- [Operator Framework Cache](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/pkg/controller/registry/resolver/cache/cache.go)
- [Boring Registry](https://github.com/boring-registry/boring-registry)
- [go-cache library](https://pkg.go.dev/github.com/patrickmn/go-cache)
