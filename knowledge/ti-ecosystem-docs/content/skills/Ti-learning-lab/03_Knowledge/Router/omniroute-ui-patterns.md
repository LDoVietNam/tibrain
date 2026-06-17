# OmniRoute UI Patterns

> **Nguồn**: OmniRoute repository tại `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\OmniRoute\`
> **Mục đích**: Học và áp dụng các UI patterns tái sử dụng từ OmniRoute vào Ti Router
> **Ngày**: 2026-05-04

---

## Tổng quan

OmniRoute sử dụng bộ UI components tái sử dụng toàn diện trong `src/shared/components/`. Các components này tuân theo các patterns nhất quán cho:
- Biến thể (variants) và kích thước (sizes)
- Accessibility (thuộc tính ARIA)
- Hỗ trợ dark mode
- Trạng thái loading và error
- Quốc tế hóa (i18n)

---

## Component Patterns

### 1. DataTable

**Vị trí**: `src/shared/components/DataTable.tsx`

**Pattern**: Bảng dữ liệu có thể cấu hình với sticky header, click vào row, trạng thái loading/empty

```tsx
<DataTable
  columns={columns}
  data={data}
  onRowClick={(row) => handleRowClick(row)}
  loading={isLoading}
  emptyState={<EmptyState icon="📭" title="Không có dữ liệu" />}
  stickyHeader
/>
```

**Tính năng chính**:
- Sticky header cho bảng dài
- Xử lý click vào row
- Trạng thái loading với skeleton
- Component empty state
- Hỗ trợ sorting và filtering cột
- Thiết kế responsive

**Ví dụ sử dụng** (từ trang Providers):
```tsx
<DataTable
  columns={[
    { key: 'name', label: 'Provider' },
    { key: 'status', label: 'Trạng thái' },
    { key: 'actions', label: '' }
  ]}
  data={providers}
  onRowClick={(provider) => router.push(`/providers/${provider.id}`)}
/>
```

---

### 2. Modal

**Vị trí**: `src/shared/components/Modal.tsx`

**Pattern**: Modal với overlay, phím escape, focus trap, khóa scroll body, nhiều kích thước

```tsx
<Modal
  isOpen={isOpen}
  onClose={handleClose}
  size="md" // sm | md | lg | xl
  title="Tiêu đề Modal"
>
  <Modal.Content>
    Nội dung modal ở đây
  </Modal.Content>
  <Modal.Footer>
    <Button variant="secondary" onClick={handleClose}>
      Hủy
    </Button>
    <Button variant="primary" onClick={handleSave}>
      Lưu
    </Button>
  </Modal.Footer>
</Modal>
```

**Tính năng chính**:
- Biến thể kích thước: sm, md, lg, xl
- Phím escape để đóng
- Focus trap cho accessibility
- Khóa scroll body khi mở
- Overlay với backdrop blur
- Sub-components: Modal.Content, Modal.Footer

**Ví dụ sử dụng** (từ OAuthModal):
```tsx
<Modal
  isOpen={isOpen}
  onClose={onClose}
  size="md"
  title="Kết nối Provider"
>
  <Modal.Content>
    {/* Nội dung OAuth flow */}
  </Modal.Content>
</Modal>
```

---

### 3. Card

**Vị trí**: `src/shared/components/Card.tsx`

**Pattern**: Card với title/subtitle/icon/action, sub-components

```tsx
<Card
  title="Tiêu đề Card"
  subtitle="Subtitle tùy chọn"
  icon="provider"
  action={<Button size="sm">Hành động</Button>}
  padding="md" // sm | md | lg
>
  <Card.Section title="Tiêu đề Section">
    Nội dung section
  </Card.Section>
  <Card.Row label="Nhãn" value="Giá trị" />
  <Card.ListItem
    label="Mục"
    value="Giá trị"
    onHoverAction={<Button size="sm">Sửa</Button>}
  />
</Card>
```

**Tính năng chính**:
- Props: title, subtitle, icon, action
- Biến thể padding
- Hiệu ứng hover trên list items
- Sub-components: Section, Row, ListItem
- ListItem với hover actions

**Ví dụ sử dụng** (từ trang Providers):
```tsx
<Card
  title={provider.name}
  subtitle={provider.type}
  icon={provider.icon}
  action={<Toggle checked={provider.enabled} onChange={toggleProvider} />}
>
  <Card.Row label="Trạng thái" value={provider.status} />
  <Card.ListItem
    label="Kết nối"
    value={provider.connectionCount}
    onHoverAction={<Button size="sm">Quản lý</Button>}
  />
</Card>
```

---

### 4. Button

**Vị trí**: `src/shared/components/Button.tsx`

**Pattern**: Biến thể, kích thước, icon, trạng thái loading

```tsx
<Button
  variant="primary" // primary | secondary | outline | ghost | danger
  size="md" // sm | md | lg
  icon="add"
  iconRight="arrow_forward"
  loading={isLoading}
  disabled={isDisabled}
  onClick={handleClick}
>
  Nội dung Button
</Button>
```

**Tính năng chính**:
- Biến thể: primary (gradient), secondary, outline, ghost, danger
- Kích thước: sm (h-7), md (h-9), lg (h-11)
- Hỗ trợ icon trái và phải
- Trạng thái loading với spinner
- Trạng thái disabled với opacity
- Animation active scale (0.99)
- Material Symbols icons

**Ví dụ sử dụng**:
```tsx
<Button variant="primary" icon="add" onClick={handleAdd}>
  Thêm Provider
</Button>
<Button variant="danger" icon="delete" onClick={handleDelete}>
  Xóa
</Button>
<Button variant="outline" iconRight="external_link" href="https://...">
  Tài liệu
</Button>
```

---

### 5. Input

**Vị trí**: `src/shared/components/Input.tsx`

**Pattern**: Label, error, hint, hỗ trợ icon, accessibility

```tsx
<Input
  label="Nhãn trường"
  type="text"
  placeholder="Nhập giá trị"
  value={value}
  onChange={handleChange}
  error={errorMessage}
  hint="Gợi ý tùy chọn"
  icon="search"
  required
/>
```

**Tính năng chính**:
- Label với chỉ thị required
- Trạng thái error với viền đỏ và thông báo error
- Text hint (hiển thị khi không có error)
- Hỗ trợ icon trái
- Accessibility: aria-required, aria-invalid, aria-describedby
- Fix zoom iOS (text-[16px] sm:text-sm)
- Focus ring với màu primary

**Ví dụ sử dụng**:
```tsx
<Input
  label="API Key"
  type="password"
  placeholder="sk-..."
  value={apiKey}
  onChange={setApiKey}
  error={apiKeyError}
  hint="OpenAI API key của bạn"
  required
/>
```

---

### 6. Select

**Vị trí**: `src/shared/components/Select.tsx`

**Pattern**: Label, error, hint, options, accessibility

```tsx
<Select
  label="Chọn tùy chọn"
  options={[
    { value: 'option1', label: 'Tùy chọn 1' },
    { value: 'option2', label: 'Tùy chọn 2' }
  ]}
  value={selectedValue}
  onChange={handleChange}
  placeholder="Chọn một tùy chọn"
  error={errorMessage}
  hint="Gợi ý tùy chọn"
  required
/>
```

**Tính năng chính**:
- Label với chỉ thị required
- Hỗ trợ error và hint
- Mảng options với value/label
- Icon dropdown tùy chỉnh
- Accessibility: aria-required, aria-invalid, aria-describedby
- Focus ring với màu primary

**Ví dụ sử dụng**:
```tsx
<Select
  label="Loại Provider"
  options={[
    { value: 'openai', label: 'OpenAI' },
    { value: 'anthropic', label: 'Anthropic' }
  ]}
  value={providerType}
  onChange={setProviderType}
  placeholder="Chọn provider"
  required
/>
```

---

### 7. Toggle

**Vị trí**: `src/shared/components/Toggle.tsx`

**Pattern**: Component switch với kích thước, label, description

```tsx
<Toggle
  checked={isEnabled}
  onChange={setIsEnabled}
  label="Tên tính năng"
  description="Mô tả tùy chọn"
  size="md" // sm | md | lg
  disabled={false}
/>
```

**Tính năng chính**:
- Kích thước: sm (w-8 h-4), md (w-11 h-6), lg (w-14 h-7)
- Hỗ trợ label và description
- Trạng thái disabled với opacity
- Role="switch" cho accessibility
- Animation chuyển mượt
- Màu primary khi checked

**Ví dụ sử dụng**:
```tsx
<Toggle
  checked={provider.enabled}
  onChange={(checked) => toggleProvider(provider.id, checked)}
  label="Bật Provider"
  description="Route requests qua provider này"
/>
```

---

### 8. Badge

**Vị trí**: `src/shared/components/Badge.tsx`

**Pattern**: Biến thể, kích thước, dot, icon

```tsx
<Badge
  variant="success" // default | primary | success | warning | error | info
  size="md" // sm | md | lg
  dot
  icon="check"
>
  Nội dung Badge
</Badge>
```

**Tính năng chính**:
- Biến thể với color schemes
- Kích thước: sm (text-[10px]), md (text-xs), lg (text-sm)
- Chỉ thị dot (vòng tròn màu)
- Hỗ trợ icon (Material Symbols)
- Bo góc đầy đủ (hình viên)

**Ví dụ sử dụng**:
```tsx
<Badge variant="success" dot>Đã kết nối</Badge>
<Badge variant="error" icon="error">Lỗi</Badge>
<Badge variant="warning">Sắp hết hạn</Badge>
```

---

### 9. FilterBar

**Vị trí**: `src/shared/components/FilterBar.tsx`

**Pattern**: Search input + filter chips với dropdown

```tsx
<FilterBar
  searchValue={searchQuery}
  onSearchChange={setSearchQuery}
  placeholder="Tìm kiếm..."
  filters={[
    { key: 'status', label: 'Trạng thái', options: ['active', 'inactive'] },
    { key: 'type', label: 'Loại', options: ['type1', 'type2'] }
  ]}
  activeFilters={activeFilters}
  onFilterChange={(key, value) => setFilters({ ...filters, [key]: value })}
>
  <Button size="sm" icon="refresh" onClick={handleRefresh}>
    Làm mới
  </Button>
</FilterBar>
```

**Tính năng chính**:
- Search input với icon
- Filter chips với menu dropdown
- Chỉ thị filter active (highlight)
- Nút clear all
- Điều khiển thêm qua prop children
- Inline styles cho prototyping nhanh

**Ví dụ sử dụng** (từ RequestLoggerV2):
```tsx
<FilterBar
  searchValue={search}
  onSearchChange={setSearch}
  placeholder="Tìm kiếm logs..."
  filters={[
    { key: 'status', label: 'Trạng thái', options: ['ok', 'error'] }
  ]}
  activeFilters={activeFilters}
  onFilterChange={handleFilterChange}
/>
```

---

### 10. OAuthModal

**Vị trí**: `src/shared/components/OAuthModal.tsx`

**Pattern**: OAuth flow đa bước phức tạp (waiting → input → success → error)

```tsx
<OAuthModal
  isOpen={isOpen}
  provider="github"
  providerInfo={{ name: 'GitHub' }}
  onSuccess={handleSuccess}
  onClose={handleClose}
/>
```

**Tính năng chính**:
- Device code flow (GitHub, Qwen, Kiro, v.v.)
- Authorization code flow với popup
- Manual input fallback
- Nhiều phương thức callback (postMessage, BroadcastChannel, localStorage)
- Phát hiện popup và xử lý timeout
- Phát hiện localhost vs remote
- Polling cho token exchange

**Trạng thái Flow**:
- `waiting` - Đợi user authorization
- `input` - Nhập callback URL thủ công
- `success` - Authorization thành công
- `error` - Authorization thất bại

**Ví dụ sử dụng**:
```tsx
<OAuthModal
  isOpen={showOAuth}
  provider="github"
  onSuccess={() => {
    // Refresh connections
    fetchConnections();
    setShowOAuth(false);
  }}
  onClose={() => setShowOAuth(false)}
/>
```

---

### 11. EmptyState

**Vị trí**: `src/shared/components/EmptyState.tsx`

**Pattern**: Empty state với icon, title, description, action button

```tsx
<EmptyState
  icon="📭"
  title="Không có dữ liệu"
  description="Thêm mục đầu tiên của bạn để bắt đầu."
  actionLabel="Thêm mục"
  onAction={handleAdd}
/>
```

**Tính năng chính**:
- Icon emoji lớn với animation bounce
- Title (fallback i18n "nothingHere")
- Text description
- Action button tùy chọn với hover effects
- Layout centered với min-height
- Inline styles cho tính linh hoạt

**Ví dụ sử dụng**:
```tsx
{providers.length === 0 && (
  <EmptyState
    icon="🔌"
    title="Không có Provider"
    description="Kết nối API provider đầu tiên của bạn để bắt đầu route requests."
    actionLabel="Thêm Provider"
    onAction={() => router.push('/providers/add')}
  />
)}
```

---

### 12. Loading

**Vị trí**: `src/shared/components/Loading.tsx`

**Pattern**: Spinner, page loading, skeleton, card skeleton

```tsx
// Spinner
<Spinner size="md" label="Đang tải..." />

// Page loading
<PageLoading message="Đang tải dữ liệu..." />

// Skeleton
<Skeleton className="h-4 w-24" />

// Card skeleton
<CardSkeleton />

// Default
<Loading type="spinner" size="md" />
```

**Tính năng chính**:
- Kích thước spinner: sm, md, lg, xl
- Page loading với overlay full-screen
- Skeleton cho nội dung placeholder
- Card skeleton cho trạng thái loading card
- Accessibility: aria-live, aria-busy, sr-only
- Hỗ trợ motion reduction

**Ví dụ sử dụng**:
```tsx
{isLoading ? (
  <CardSkeleton />
) : (
  <Card title={provider.name}>{content}</Card>
)}

{isPageLoading && <PageLoading message="Đang tải providers..." />}
```

---

### 13. SegmentedControl

**Vị trí**: `src/shared/components/SegmentedControl.tsx`

**Pattern**: Control dạng tab với icons

```tsx
<SegmentedControl
  options={[
    { value: 'tab1', label: 'Tab 1', icon: 'tab' },
    { value: 'tab2', label: 'Tab 2', icon: 'list' }
  ]}
  value={activeTab}
  onChange={setActiveTab}
  size="md" // sm | md | lg
  aria-label="Tùy chọn xem"
/>
```

**Tính năng chính**:
- Segmented control dạng tab
- Hỗ trợ icon cho mỗi option
- Kích thước: sm, md, lg
- Trạng thái active với background
- Accessibility: role="tablist", role="tab", aria-selected
- Transitions mượt

**Ví dụ sử dụng**:
```tsx
<SegmentedControl
  options={[
    { value: 'grid', label: 'Lưới', icon: 'grid_view' },
    { value: 'list', label: 'Danh sách', icon: 'list' }
  ]}
  value={viewMode}
  onChange={setViewMode}
/>
```

---

## Patterns Chung

### 1. Accessibility

Tất cả components tuân theo best practices accessibility:
- Thuộc tính ARIA (aria-label, aria-required, aria-invalid, aria-describedby)
- Semantic HTML (thuộc tính role)
- Điều hướng bàn phím (tabIndex)
- Hỗ trợ screen reader (text sr-only)
- Quản lý focus (focus trap trong modals)

### 2. Dark Mode

Components sử dụng CSS custom properties cho dark mode:
```css
color: var(--text-primary, #e0e0e0);
background: rgba(255,255,255,0.05);
border: 1px solid rgba(255,255,255,0.1);
```

### 3. Quốc tế hóa (i18n)

Components sử dụng `next-intl` cho translations:
```tsx
const t = useTranslations("common");
<EmptyState title={t("nothingHere")} />
```

### 4. Biến thể Kích thước

Patterns kích thước nhất quán trên components:
- `sm`: Nhỏ (compact)
- `md`: Trung bình (mặc định)
- `lg`: Lớn

### 5. Hệ thống Biến thể

Components sử dụng objects variant cho styling:
```tsx
const variants = {
  primary: "bg-primary text-white",
  secondary: "bg-white/10 border",
  ghost: "text-text-muted hover:bg-black/5"
};
```

### 6. Utility Function (cn)

Tất cả components sử dụng utility `cn` cho merging className:
```tsx
import { cn } from "@/shared/utils/cn";
className={cn("base-class", variant, size, className)}
```

### 7. Hệ thống Icon

Sử dụng nhất quán Material Symbols icons:
```tsx
<span className="material-symbols-outlined text-[18px]">
  icon_name
</span>
```

### 8. Xử lý Error

Patterns trạng thái error nhất quán:
```tsx
{error && (
  <p className="text-xs text-red-500 flex items-center gap-1">
    <span className="material-symbols-outlined">error</span>
    {error}
  </p>
)}
```

---

## Áp dụng vào Ti Router

### Patterns Cần Áp dụng

1. **DataTable** - Dùng cho danh sách auth files, providers, logs
2. **Modal** - Dùng cho dialogs (thêm provider, edit config, xem models)
3. **Card** - Dùng cho provider cards, connection cards
4. **Button** - Chuẩn hóa biến thể và kích thước button
5. **Input/Select** - Chuẩn hóa form inputs với error handling
6. **Badge** - Dùng cho chỉ thị trạng thái (đã kết nối, lỗi, warning)
7. **EmptyState** - Dùng cho trạng thái no-data
8. **Loading** - Dùng cho trạng thái loading (skeletons, spinners)
9. **FilterBar** - Dùng cho filtering danh sách (providers, logs)
10. **Toggle** - Dùng cho switches bật/tắt

### Đã Áp dụng

✅ **Modal pattern** - Đã áp dụng vào AuthFileModelsDialog trong Ti Router
- Sử dụng cấu trúc Modal component
- Thêm sub-components Content và Footer
- Triển khai handler onClose

### Các Bước Tiếp theo

1. Tạo folder shared components trong Ti Router
2. Triển khai DataTable cho danh sách auth files
3. Thêm Card components cho hiển thị provider
4. Triển khai FilterBar cho filtering
5. Thêm EmptyState cho scenarios no-data
6. Chuẩn hóa biến thể Button trên toàn app
7. Thêm Badge cho chỉ thị trạng thái
8. Triển khai Toggle cho bật/tắt

---

## Tham khảo

- **Nguồn OmniRoute**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\OmniRoute\src\shared\components\`
- **Ti Router**: `Z:\10_WORKPLACE\Ti\apps\router\`
- **Skyvern UI Patterns**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\Router\skyvern-ui-patterns.md`
