# OmniRoute UI Patterns - Kế Hoạch Áp Dụng Cho Ti Router

> **Mục đích**: Kế hoạch áp dụng OmniRoute UI patterns vào Ti Router
> **Ngày tạo**: 2026-05-04
> **Trạng thái**: Sẵn sàng thực thi
> **Ưu tiên**: P1 (Cao)

---

## Tổng quan

Áp dụng 13 UI component patterns từ OmniRoute để cải thiện Ti Router UI:
- Tăng tính nhất quán (consistency)
- Cải thiện accessibility
- Chuẩn hóa dark mode
- Tối ưu user experience

---

## Phase 1: Nền tảng (Shared Components)

### 1.1 Tạo Cấu Trúc Shared Components

**Vị trí**: `apps/router/web/src/components/shared/`

**Tasks**:
- [ ] Tạo cấu trúc folder:
  ```
  components/
  └── shared/
      ├── Button.tsx
      ├── Input.tsx
      ├── Select.tsx
      ├── Toggle.tsx
      ├── Badge.tsx
      ├── Modal/
      │   ├── Modal.tsx
      │   ├── ModalContent.tsx
      │   └── ModalFooter.tsx
      ├── Card/
      │   ├── Card.tsx
      │   ├── CardSection.tsx
      │   ├── CardRow.tsx
      │   └── CardListItem.tsx
      ├── DataTable.tsx
      ├── FilterBar.tsx
      ├── EmptyState.tsx
      ├── Loading.tsx
      └── SegmentedControl.tsx
  ```

- [ ] Tạo `index.ts` để export tất cả components
- [ ] Thêm utility function `cn()` cho merging className
- [ ] Setup CSS variables cho hỗ trợ dark mode

**Thời gian ước tính**: 2-3 giờ

---

## Phase 2: Core Components (Theo Ưu tiên)

### 2.1 Button Component

**Nguồn**: OmniRoute `src/shared/components/Button.tsx`

**Implementation**:
```tsx
// Button.tsx
const variants = {
  primary: "bg-gradient-to-b from-primary to-primary-hover text-white",
  secondary: "bg-white/10 border border-white/10",
  outline: "border border-white/15",
  ghost: "text-text-muted hover:bg-black/5",
  danger: "bg-red-500 text-white",
};

const sizes = {
  sm: "h-7 px-3 text-xs",
  md: "h-9 px-4 text-sm",
  lg: "h-11 px-6 text-sm",
};
```

**Áp dụng vào các button hiện có**:
- [ ] `AuthFilesPage.tsx` - Button xem models
- [ ] Tất cả action buttons trên toàn app

**Thời gian ước tính**: 1 giờ

---

### 2.2 Input & Select Components

**Nguồn**: OmniRoute `src/shared/components/Input.tsx`, `Select.tsx`

**Implementation**:
- [ ] Input với label, error, hint, hỗ trợ icon
- [ ] Select với options, error, hint support
- [ ] Thuộc tính accessibility (aria-required, aria-invalid, aria-describedby)

**Áp dụng vào các forms hiện có**:
- [ ] Auth file forms
- [ ] Provider configuration forms
- [ ] Settings forms

**Thời gian ước tính**: 2 giờ

---

### 2.3 Toggle Component

**Nguồn**: OmniRoute `src/shared/components/Toggle.tsx`

**Implementation**:
- [ ] Switch với kích thước (sm, md, lg)
- [ ] Hỗ trợ label và description
- [ ] Trạng thái disabled

**Áp dụng vào**:
- [ ] Provider enable/disable toggles
- [ ] Feature toggles trong settings

**Thời gian ước tính**: 1 giờ

---

### 2.4 Badge Component

**Nguồn**: OmniRoute `src/shared/components/Badge.tsx`

**Implementation**:
- [ ] Biến thể: default, primary, success, warning, error, info
- [ ] Kích thước: sm, md, lg
- [ ] Chỉ thị dot
- [ ] Hỗ trợ icon

**Áp dụng vào**:
- [ ] Connection status badges (đã kết nối, lỗi, warning)
- [ ] Provider type badges
- [ ] Health status badges

**Thời gian ước tính**: 1 giờ

---

### 2.5 Modal Component

**Nguồn**: OmniRoute `src/shared/components/Modal.tsx`

**Implementation**:
- [ ] Kích thước: sm, md, lg, xl
- [ ] Phím escape để đóng
- [ ] Focus trap
- [ ] Khóa scroll body
- [ ] Sub-components: ModalContent, ModalFooter

**Áp dụng vào**:
- [ ] ✅ AuthFileModelsDialog (đã implement)
- [ ] Provider configuration modal
- [ ] Settings modal
- [ ] Confirmation dialogs

**Thời gian ước tính**: 2 giờ

---

## Phase 3: Advanced Components

### 3.1 Card Component

**Nguồn**: OmniRoute `src/shared/components/Card.tsx`

**Implementation**:
- [ ] Card với title, subtitle, icon, action
- [ ] Sub-components: CardSection, CardRow, CardListItem
- [ ] Biến thể padding
- [ ] Hiệu ứng hover trên list items

**Áp dụng vào**:
- [ ] Provider cards
- [ ] Connection cards
- [ ] Status cards

**Thời gian ước tính**: 2 giờ

---

### 3.2 DataTable Component

**Nguồn**: OmniRoute `src/shared/components/DataTable.tsx`

**Implementation**:
- [ ] Sticky header
- [ ] Row click handler
- [ ] Trạng thái loading với skeleton
- [ ] Hỗ trợ empty state
- [ ] Sorting và filtering cột

**Áp dụng vào**:
- [ ] Auth files list (thay thế bảng hiện tại)
- [ ] Providers list
- [ ] Logs display

**Thời gian ước tính**: 3 giờ

---

### 3.3 FilterBar Component

**Nguồn**: OmniRoute `src/shared/components/FilterBar.tsx`

**Implementation**:
- [ ] Search input với icon
- [ ] Filter chips với dropdown
- [ ] Chỉ thị filter active
- [ ] Nút clear all
- [ ] Điều khiển thêm qua children

**Áp dụng vào**:
- [ ] Providers page filter
- [ ] Auth files filter
- [ ] Logs filter

**Thời gian ước tính**: 2 giờ

---

### 3.4 EmptyState Component

**Nguồn**: OmniRoute `src/shared/components/EmptyState.tsx`

**Implementation**:
- [ ] Icon với animation bounce
- [ ] Title và description
- [ ] Action button tùy chọn
- [ ] Layout centered

**Áp dụng vào**:
- [ ] Trạng thái không có providers
- [ ] Trạng thái không có auth files
- [ ] Các trạng thái no-data khác

**Thời gian ước tính**: 1 giờ

---

### 3.5 Loading Component

**Nguồn**: OmniRoute `src/shared/components/Loading.tsx`

**Implementation**:
- [ ] Spinner (sm, md, lg, xl)
- [ ] Page loading với overlay
- [ ] Skeleton cho placeholders
- [ ] Card skeleton

**Áp dụng vào**:
- [ ] Initial page load
- [ ] Async operations
- [ ] Card loading states
- [ ] Table loading states

**Thời gian ước tính**: 1 giờ

---

### 3.6 SegmentedControl Component

**Nguồn**: OmniRoute `src/shared/components/SegmentedControl.tsx`

**Implementation**:
- [ ] Control dạng tab
- [ ] Hỗ trợ icon cho mỗi option
- [ ] Kích thước: sm, md, lg
- [ ] Trạng thái active với background

**Áp dụng vào**:
- [ ] View mode switcher (grid/list)
- [ ] Tab navigation

**Thời gian ước tính**: 1 giờ

---

## Phase 4: Integration & Migration

### 4.1 Migrate Existing Components

**Tasks**:
- [ ] Thay thế tất cả buttons hiện có với Button component
- [ ] Thay thế tất cả inputs hiện có với Input component
- [ ] Thay thế tất cả selects hiện có với Select component
- [ ] Thay thế tất cả modals hiện có với Modal component
- [ ] Thêm badges vào chỉ thị trạng thái
- [ ] Thêm EmptyState vào các scenarios no-data
- [ ] Thêm Loading states vào async operations

**Thời gian ước tính**: 4-5 giờ

---

### 4.2 Cập Nhật Styling

**Tasks**:
- [ ] Thêm CSS variables cho dark mode
- [ ] Cập nhật Tailwind config nếu cần
- [ ] Đảm bảo Material Symbols icons có sẵn
- [ ] Test dark mode trên tất cả components

**Thời gian ước tính**: 2 giờ

---

### 4.3 Accessibility Audit

**Tasks**:
- [ ] Verify thuộc tính ARIA trên tất cả components
- [ ] Test điều hướng bàn phím
- [ ] Test tương thích screen reader
- [ ] Verify quản lý focus trong modals

**Thời gian ước tính**: 2 giờ

---

## Phase 5: Testing & Documentation

### 5.1 Component Testing

**Tasks**:
- [ ] Test tất cả components trong light mode
- [ ] Test tất cả components trong dark mode
- [ ] Test thiết kế responsive
- [ ] Test accessibility (bàn phím, screen reader)
- [ ] Test trạng thái error
- [ ] Test trạng thái loading

**Thời gian ước tính**: 3 giờ

---

### 5.2 Documentation

**Tasks**:
- [ ] Cập nhật documentation components
- [ ] Thêm usage examples cho mỗi component
- [ ] Tạo storybook hoặc component preview page
- [ ] Cập nhật AGENTS.md với patterns mới

**Thời gian ước tính**: 2 giờ

---

## Tóm Tắt Timeline

| Phase | Mô tả | Thời gian ước tính |
|-------|-------------|----------------|
| Phase 1 | Nền tảng (Cấu trúc Shared Components) | 2-3 giờ |
| Phase 2 | Core Components (Button, Input, Toggle, Badge, Modal) | 7 giờ |
| Phase 3 | Advanced Components (Card, DataTable, FilterBar, EmptyState, Loading, SegmentedControl) | 10 giờ |
| Phase 4 | Integration & Migration | 8-9 giờ |
| Phase 5 | Testing & Documentation | 5 giờ |
| **Tổng** | **Tất cả Phases** | **32-34 giờ (~4 ngày)** |

---

## Thứ Tự Ưu tiên

### Tuần 1 (Nền tảng + Core)
1. ✅ Phase 1: Tạo cấu trúc shared components
2. ✅ Phase 2.1: Button component
3. ✅ Phase 2.2: Input & Select components
4. ✅ Phase 2.3: Toggle component
5. ✅ Phase 2.4: Badge component
6. ✅ Phase 2.5: Modal component (đã làm một phần)

### Tuần 2 (Advanced)
7. ✅ Phase 3.1: Card component
8. ✅ Phase 3.2: DataTable component
9. ✅ Phase 3.3: FilterBar component
10. ✅ Phase 3.4: EmptyState component
11. ✅ Phase 3.5: Loading component
12. ✅ Phase 3.6: SegmentedControl component

### Tuần 3 (Integration + Testing)
13. ✅ Phase 4: Integration & Migration
14. ✅ Phase 5: Testing & Documentation

---

## Dependencies

**Bắt buộc**:
- ✅ OmniRoute UI patterns đã document (`omniroute-ui-patterns.md`)
- ✅ Material Symbols icons (Google Fonts)
- ✅ Tailwind CSS v3+ (đã có trong Ti Router)
- ✅ TypeScript (đã có trong Ti Router)

**Tùy chọn**:
- Storybook cho component preview (khuyến nghị)
- Testing library (Jest + React Testing Library)

---

## Tiêu Chí Thành Công

- [ ] Tất cả 13 components đã implement
- [ ] UI hiện tại đã migrate sang components mới
- [ ] Dark mode hoạt động trên tất cả components
- [ ] Accessibility đã verify (ARIA, bàn phím, screen reader)
- [ ] Thiết kế responsive đã test
- [ ] Documentation hoàn tất
- [ ] Không có breaking changes vào chức năng hiện có

---

## Ghi chú

- **OAuthModal** từ OmniRoute khá phức tạp và specific cho OAuth flows - có thể không cần implement cho Ti Router ban đầu
- Bắt đầu với core components (Button, Input, Toggle, Badge, Modal) vì chúng có tác động cao nhất
- Sử dụng progressive migration - thay thế components từng phần
- Test kỹ lưỡng sau khi implement mỗi component

---

## Tham Khảo

- **OmniRoute UI Patterns**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\Router\omniroute-ui-patterns.md`
- **Skyvern UI Patterns**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\Router\skyvern-ui-patterns.md`
- **Ti Router**: `Z:\10_WORKPLACE\Ti\apps\router\`
