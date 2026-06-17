# Rules for Task Classification and Documentation Management

> **Phiên bản**: 1.0.0
> **Cập nhật lần cuối**: 2026-05-06
> **Mục đích**: Tránh lỗi phân loại sai dựa trên đường dẫn file thay vì chức năng thực tế

---

## Rule 1: Hiểu rõ kiến trúc trước khi phân loại

**SAI**: Phân loại dựa trên đường dẫn file
**ĐÚNG**: Phân loại dựa trên chức năng và mối quan hệ giữa các component

**Kiến trúc Router Agent**:
```
Router Agent (AI Agent trong Ti Crew)
    ↓ Quản lý
Ti Router (LLM Gateway - apps/core/router/)
    ↓ Routing đến
LLM Providers (28+ providers)
```

**Mối quan hệ**:
- Router Agent là "bộ não" quản lý Ti Router
- Ti Router là nền tảng nơi Router Agent hoạt động
- Các plan về Ti Router đều liên quan vì chúng tạo nền tảng cho Router Agent

---

## Rule 2: Đọc README và hiểu context trước khi quyết định

**Bước bắt buộc**:
1. Đọc README.md của component chính
2. Hiểu role, position, relationship
3. Kiểm tra xem các plan đề cập đến cái gì
4. Xác định mối quan hệ thực tế

**Checklist**:
- [ ] Đọc README.md của Router Agent
- [ ] Hiểu role: "AI Agent Của Toàn Bộ Ti Router"
- [ ] Hiểu position: Embedded routing helper trong Ti Router
- [ ] Đọc các plan gốc để xem đề cập đến cái gì
- [ ] Xác định mối quan hệ: Router Agent quản lý Ti Router

---

## Rule 3: Không phân loại dựa trên đường dẫn file

**SAI**:
- "Plan đề cập đến apps/router/ nên không liên quan đến packages/ticrew/members/router-agent/"
- "Plan ở Ti-learning-lab nên không liên quan"

**ĐÚNG**:
- "Plan đề cập đến apps/router/ nhưng apps/router/ là nơi Router Agent hoạt động → CÓ LIÊN QUAN"
- "Plan ở Ti-learning-lab nhưng nói về learning system cho Ti → CÓ LIÊN QUAN"

**Phân loại dựa trên**:
- Chức năng thực tế
- Mục tiêu của plan
- Mối quan hệ giữa các component
- Impact đến Router Agent

---

## Rule 4: Xác định 3 loại mối quan hệ

### 1. Trực tiếp (Direct)
- Plan nói trực tiếp về Router Agent
- Ví dụ: Build Plan, Documentation từ router-agent-docs

### 2. Gián tiếp (Indirect)  
- Plan nói về nền tảng nơi Router Agent hoạt động
- Ví dụ: Router Completion Plan, Priority Roadmap, AI_AGENT_ROUTER_PLAN.md

### 3. Không liên quan (Unrelated)
- Plan nói về hệ thống hoàn toàn khác
- Ví dụ: Plan về UI khác, CLI khác

**Quy tắc**: Trực tiếp và Gián tiếp đều CÓ LIÊN QUAN

---

## Rule 5: Kiểm tra sâu trước khi kết luận

**Bước kiểm tra**:
1. Đọc file plan gốc (full content)
2. Kiểm tra cấu trúc thư mục thực tế
3. Đọc README của các component liên quan
4. Hiểu kiến trúc tổng thể
5. Xác định mối quan hệ

**Không được**:
- Chỉ đọc tiêu đề hoặc vài dòng đầu
- Chỉ kiểm tra đường dẫn file
- Kết luận dựa trên giả định

---

## Rule 6: Khi nghi ngờ, hỏi người dùng

**Khi không chắc chắn**:
- "Tôi thấy plan này đề cập đến X, nhưng không chắc có liên quan đến Router Agent không?"
- "Mối quan hệ giữa X và Y là gì?"
- "Bạn có thể giải thích rõ hơn về kiến trúc không?"

**Không được**:
- Tự kết luận dựa trên thông tin không đầy đủ
- Giả định mối quan hệ
- Phân loại mà không hiểu rõ

---

## Rule 7: Cập nhật kiến thức khi có thông tin mới

**Khi người dùng cung cấp thông tin mới**:
- Cập nhật hiểu biết về kiến trúc
- Điều chỉnh phân loại nếu cần
- Không cứng nhắc với phân loại cũ

**Ví dụ**:
- Người dùng nói: "Router Agent là thành viên đầu tiên của crew"
- Hành động: Cập nhật hiểu biết, tất cả plan về Ti Router đều liên quan

---

## Rule 8: Ghi lại lý do phân loại

**Khi phân loại task/plan**:
- Ghi rõ lý do tại sao CÓ LIÊN QUAN hoặc KHÔNG LIÊN QUAN
- Dựa trên chức năng, không đường dẫn
- Dựa trên mối quan hệ thực tế

**Ví dụ**:
```markdown
- AI_AGENT_ROUTER_PLAN.md: CÓ LIÊN QUAN vì nói về việc biến Ti Router thành nơi Router Agent có thể hoạt động
- Router Completion Plan: CÓ LIÊN QUAN vì hoàn thiện nền tảng nơi Router Agent hoạt động
```

---

## Rule 9: Review lại phân loại sau khi hoàn thành

**Bước review**:
1. Đọc lại tất cả phân loại
2. Kiểm tra có vi phạm rules nào không
3. Điều chỉnh nếu cần
4. Xác nhận với người dùng nếu nghi ngờ

---

## Rule 10: Học từ sai lầm

**Khi mắc lỗi**:
- Ghi lại lỗi sai
- Phân tích nguyên nhân
- Cập nhật rules để tránh lặp lại
- Áp dụng rules mới vào các phân loại sau

**Ví dụ lỗi này**:
- Lỗi: Phân loại dựa trên đường dẫn file
- Nguyên nhân: Không hiểu kiến trúc và mối quan hệ
- Bài học: Luôn đọc README và hiểu context trước khi phân loại
- Rule mới: Rule 1, Rule 2, Rule 3

---

**Last Updated**: 2026-05-06
**Trigger**: Lỗi phân loại dựa trên đường dẫn file thay vì chức năng thực tế
