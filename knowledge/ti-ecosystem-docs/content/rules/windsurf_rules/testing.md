# Testing Rule

> **Priority**: P0
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\testing.md`

---

## 🧪 Testing Requirements (Bắt Buộc)

### Khi Nào Phải Test
- ✅ Sau khi implement code mới
- ✅ Sau khi sửa bug
- ✅ Trước khi commit (nếu user yêu cầu)

### Loại Tests
| Language | Test Framework | Command |
|----------|---------------|---------|
| **Go** | built-in testing | `go test ./...` |
| **TypeScript/React** | Vitest/Jest | `npm test` |
| **Python** | pytest | `pytest` |

---

## 📋 Test Checklist

### Unit Tests
- Test các functions/methods chính
- Mock external dependencies
- Cover cả positive và negative cases

### Integration Tests
- Test API endpoints (nếu có)
- Test tương tác giữa các modules
- Verify data flow đúng

### E2E Tests (Khi Có)
- Test user flows chính
- Verify UI behavior (nếu là frontend)

---

## 🚫 Không Được Phép
- ❌ Skip tests khi commit (trừ khi có lý do bất khả kháng)
- ❌ Commit code không compile/run được
- ❌ Ignore test failures mà không fix

---

## 🔗 References
- **Go Testing**: https://go.dev/doc/effective_go#testing
- **Vitest**: https://vitest.dev/
- **TDD Workflow**: `Z:\10_WORKPLACE\Ti\content\skills\tdd-workflow\`
