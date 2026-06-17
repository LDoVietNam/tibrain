# Agent Team trong Claude Code - Video Notes

## Thông Tin Video
- **Title:** Làm chủ 90% kỹ năng về Agent Team trong Claude Code chỉ trong 22 phút
- **Creator:** Long Phan (@pnhlong88)
- **URL:** https://www.youtube.com/watch?v=fWUgYTD3Jvs
- **Duration:** ~22 phút
- **Model:** Claude 4.6 Opus
- **Ngày xem:** 2026-04-05

## 📝 Ghi Chú Từ Video

### 1. Phân Biệt Chế Độ
| Chế Độ | Đặc Điểm | Hạn Chế |
|--------|----------|---------|
| **Mặc định** | 1 phiên duy nhất | Context limit → ảo giác, quên ngữ cảnh |
| **Subagent** | Agent phụ thực hiện task, báo cáo tóm tắt | Không có session riêng |
| **Agent Team** | Team Lead + Members, mỗi member có session riêng | Phức tạp hơn |

### 2. Vai Trò Trong Team
- **Đội trưởng (Captain):**
  - Nhận yêu cầu từ user
  - Lập kế hoạch tổng thể
  - Chia nhỏ task
  - Cung cấp context cho từng member

- **Members:**
  - Mỗi member có session riêng biệt
  - Tập trung vào khu vực cụ thể của codebase
  - Không bị phân tâm bởi các phần khác

### 3. Cơ Chế Quan Trọng
- **Race Condition Check:** Khóa task khi agent bắt đầu xử lý → tránh 2 agents làm cùng 1 task
- **Memory Files:** Dùng `.md` files để chia sẻ context giữa agents (vì không share history)
- **Skill Creation:** Biến quy trình thành Skill để tái sử dụng (`/research`)

### 4. Công Cụ Hỗ Trợ
- **TMUX:** Chia terminal thành nhiều panes → theo dõi từng agent độc lập
- **Config:** `experimental.agent.team` trong `settings.json`
- **Cost Optimization:** Team Lead dùng Opus 4.6, Members dùng Sonnet (rẻ hơn)

### 5. Lifecycle
- **Start:** Captain tạo team, phân task
- **Execute:** Members làm việc song song với session riêng
- **Complete:** Captain dọn dẹp, đóng sessions
- **Debug mode:** Giữ agents ở chế độ chờ để maintain context

## 🎯 Key Takeaways - Áp Dụng Vào Ticlaw

| # | Kỹ năng từ video | Áp dụng vào Ticlaw | Priority |
|---|------------------|-------------------|----------|
| 1 | Team Lead + Members pattern | Captain Agent → Worker Agents | P0 |
| 2 | Session riêng per agent | Mỗi agent có context window riêng | P0 |
| 3 | Race Condition Check | Task locking system | P0 |
| 4 | Memory Files (.md) | Shared context files giữa agents | P1 |
| 5 | Skill Creation | Biến workflow thành reusable skills | P1 |
| 6 | Cost Optimization | Lead dùng model xịn, Workers dùng model rẻ | P2 |
| 7 | TMUX-like monitoring | Dashboard theo dõi từng agent | P2 |

## 📋 Action Items
- [x] Xem video và tóm tắt nội dung chính
- [x] Extract patterns có thể áp dụng
- [x] Viết implementation plan cho Ticlaw
- [ ] Tạo analysis report → `01_Analysis/`
- [ ] Bắt đầu implement Phase 1 (Agent Registry)
