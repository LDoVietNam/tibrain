# Hệ đa Agent LangGraph (triển khai trong thư mục backend)

## 1. Triển khai tại Z:\SnJ\backend
Cấu trúc LangGraph và CliProxy đặt trực tiếp trong ackend để tiện tích hợp các dịch vụ nội bộ (n8n, MailHub, ProxyHub, Telegram) với API server hiện có. Tất cả tài liệu mô tả agent, workflow và prompt đều viết bằng tiếng Việt.

## 2. Mẫu prompt S.P.R.I.N.T cho từng agent
- **S (System Role):** Mô tả vai trò rõ ràng, ví dụ  Bạn là Kỹ sư Router backend điều phối toàn bộ hệ LangGraph.
- **P (Purpose):** Cho biết task cụ thể (ví dụ quản lý template MailHub, deploy workflow n8n, forward message Telegram).
- **R (Resources/Tools):** Liệt kê node, API, CliProxy, token tương ứng.
- **I (Instructions):** Ghi rõ bước thực hiện (nhận input, validate, gọi tool, update state, chuyển agent).
- **N (Negative Constraints):** Những lệnh không được chạy, hành động cấm (không deploy khi chưa review, không forward message nhạy cảm).
- **T (Transfer Logic):** Điều kiện chuyển luồng agent khác (ví dụ n8n hoàn thành chuyển Executor, CliProxy lỗi thì chuyển Error Handler).

## 3. Các agent chính trong backend
- **Manager Router:** Nhận yêu cầu từ API, xác định dịch vụ (n8n/MailHub/ProxyHub/Telegram) rồi cập nhật state.next_agent. Không viết code trực tiếp.
- **n8n Agent:** Soạn workflow JSON, đánh dấu state.n8n_ready, chuyển sang Executor khi đã được review.
- **MailHub Agent:** Kiểm tra template, gửi request kiểm thử qua CliProxy, cập nhật state.mailhub.last_sync, chuyển lỗi sang Error Handler nếu cần.
- **ProxyHub Agent:** Kiểm tra pool proxy, chạy script proxy-checker qua CliProxy, rotate credential khi thiết lập, giữ fallback.
- **Telegram Agent:** Theo dõi watchlist, forward summary, tạo task cho Manager khi cần hành động, tránh duplicate via metadata.
- **Executor (CliProxy):** Chỉ chạy lệnh an toàn, capture stdout/stderr/exit code, log vào state.executed_commands, đưa lỗi cho Error Handler.

## 4. Quản lý trạng thái Graph
State nằm trong ackend/langgraph/state.py hoặc tương tự nên ghi nhận: messages, 
ext_agent, history, lags (n8n_ready, proxy_checked), executed_commands, errors. Tránh chạy lại lệnh (ghi 
pm install sau khi chạy). Node interrupt_before yêu cầu xác nhận con người mỗi khi lệnh rủi ro cao (xóa DB). Khi CliProxy trả lỗi thì đưa state.next_agent = ErrorHandler.

## 5. Giám sát CliProxy và Telegram log
- Executor luôn kiểm tra exit_code; nếu khác 0 chuyển sang node xử lý lỗi.
- Ghi log từng lệnh kèm timestamp, đẩy thông báo vào Telegram Agent nếu là cảnh báo.
- Telegram Agent có thể nhận alert từ Logs và forward tới nhóm Ops.

## 6. Tiếp theo
Bạn có thể:
1. Tạo file prompt riêng cho mỗi agent trong ackend/prompts/ để tái sử dụng.
2. Viết config LangGraph (StateGraph + nodes) trong ackend/langgraph/graph.py theo pseudo-code đã mô tả trước đó.
3. Liên kết Transaction state với API server (vd CLIProxyAPIPlus/src/internal/api/server.go) để nhận sự kiện và khởi chạy LangGraph node cần thiết.

Muốn mình giúp viết file prompt cụ thể hoặc cấu trúc graph code mẫu nào không?
