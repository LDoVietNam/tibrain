# Router Agent Rules

Rules cho làm việc với Router Agent trong Ti Router.

## Rule 1: Luôn đọc documentation trước khi code

Khi làm việc với Router Agent, luôn đọc:
1. `packages/ticrew/members/router-agent/README.md` - Overview
2. `packages/ticrew/members/router-agent/docs/ROUTER_AGENT.md` - Chi tiết
3. `packages/ticrew/members/router-agent/master-plan.md` - Plans

## Rule 2: Hiểu kiến trúc modular

Router Agent có 17 runtime files, mỗi file có chức năng riêng:
- router_agent.go - Runtime implementation chính
- router_bridge.go - Bridge layer với EventHook system
- router_integration.go - Integration layer
- router_controller.go - Controller layer
- router_circuit_breaker.go - Circuit Breaker
- router_strategies.go - 13 Routing Strategies
- router_tier_routing.go - Per-Model Tier Routing
- router_trivial_probe.go - Trivial Probe Optimization
- router_session_pool.go - OAuth Session Pool
- router_reaction_engine.go - Reaction Engine
- router_latency_tracker.go - Latency Tracker
- router_rate_limit.go - Rate Limiting
- router_analytics.go - Aggregate Analytics
- router_optimizer.go - Self-Optimization
- router_scoring.go - Scoring & Learning
- router_dashboard.go - HTTP API Dashboard

Không xóa hoặc gộp các files này trừ khi có lý do rõ ràng.

## Rule 3: Không sửa definition/contract

File `packages/ticrew/members/router-agent/agent.go` là definition/contract của Router Agent.
Chỉ sửa khi cần thay đổi cấu trúc chính của Router Agent.

## Rule 4: Giữ EventHook system

router_bridge.go có EventHook system độc nhất. Không xóa trừ khi đã tích hợp vào router_integration.go.

## Rule 5: Test sau khi sửa

Sau khi sửa bất kỳ file trong runtime/, luôn:
- Run tests
- Check lint errors
- Verify functionality

## Rule 6: Update documentation

Sau khi thay đổi:
- Cập nhật README.md nếu cần
- Cập nhật docs/ROUTER_AGENT.md nếu cần
- Cập nhật plan files nếu cần

## Rule 7: Sử dụng workflow

Sử dụng workflow `router-agent-workflow.md` khi làm việc với Router Agent.
