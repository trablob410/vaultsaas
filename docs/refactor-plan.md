# Valt — Kế hoạch refactor (soạn 2026-08-31)

Nguồn: rà soát code trực tiếp ngày 2026-08-31 + thảo luận chiến lược
(xem [decisions.md](decisions.md) cho lý do từng hướng). Mục tiêu refactor:
biến Valt từ "vault có SaaS layer dở dang" thành **MCP credential gate tự host,
thuần Apache-2.0**, với ba key features: dynamic lease, revoke cascade, audit 4W1H.

## A. Kết quả rà code (chân dung hiện trạng)

### Dynamic lease — `internal/dynsecret/` (688 dòng)
- Kiến trúc đúng hướng: `Provider` interface (Create/Revoke/Renew), 1 impl
  `postgres` tạo role tạm `valt_xxxx` (`CREATE ROLE ... VALID UNTIL`, password
  random, CONNECTION LIMIT 5).
- **Đứt gãy chính:** workflow approval KHÔNG gọi dynsecret — lease tạo qua
  endpoint riêng. Luồng "xin quyền → duyệt → nhận dynamic lease" chưa tồn tại.
- Expiry worker 60s chỉ `UPDATE revoked_at` trong DB, không DROP role ở backend
  (role tự chết theo VALID UNTIL nhưng để lại rác vĩnh viễn trong pg).
- TODO trong code: lease credential mã hóa bằng master key, chưa per-project key.
- Chỉ 1 provider; thiếu provider "derived API key" (use case cấp AI per nhân viên).

### Revoke — chỉ theo đơn vị lẻ
- Có `RevokeCredential(requestID)`, `RevokeLease(leaseID)`.
- Không có: `RemoveMember` org (chỉ có add/list), revoke-all theo user, hook IdP.

### Audit — schema tốt, hash-chain có 2 bug thật
- `Entry` đã có `ip_address`, `user_agent`, `region_code`, `metadata`;
  `LogFromRequest` tự điền IP (XFF/X-Real-IP) + UA. Nhưng chỉ 2 call site dùng IP.
- **Bug 1:** `ComputeHash` chỉ hash `prev|user|action|resource|event_type|status`
  — không bao gồm IP, UA, metadata, event_time → sửa các trường đó trong DB
  KHÔNG phá vỡ chuỗi. Chain chưa đúng nghĩa "chống sửa".
- **Bug 2:** `lastHash` chỉ in-memory — không load lại từ DB khi restart, không
  mutex khi ghi đồng thời → chuỗi fork, `VerifyChain` fail.
- Chưa có endpoint verify-chain, export CSV/PDF.

### Khác
- README claim "zero-knowledge" nhưng Phase 10 cho server decrypt rồi trả
  plaintext qua API → sửa claim hoặc làm client-side decrypt (chọn sửa claim trước).
- Client: `valt-cli` đã có thiết kế đúng (`setup`, `mcp install --ide claude`,
  `run` inject env, `request/status`). Release workflow chỉ build CLI
  (GoReleaser, tag `cli/v*`), chưa build MCP server Rust.

## B. Giữ – Cắt – Sửa

**GIỮ:** vault core (envelope encryption), workflow approval + policy engine,
agent identity, audit (sau khi fix), dynsecret, org/workspace/project, MCP server,
valt-cli, docker-compose + scripts VPS, `plans/` (tham khảo).

**CẮT:** landing pricing section, dashboard `/settings/upgrade`, SaaS super-admin
pages, billing package — theo Quyết định 1 (không SaaS). Cắt bằng cách ẩn/xóa route,
giữ migration DB nguyên (không phá dữ liệu người dùng cũ).

**SỬA:** hash-chain (bug 1+2), README (bỏ claim zero-knowledge sai, viết lại theo
định vị credential gate), điền where-context ở mọi call site audit.

## C. Roadmap 90 ngày (thứ tự theo dependency)

| Tuần | Việc | Đầu ra đo được |
|---|---|---|
| 1–2 | Fix hash-chain: hash toàn bộ trường (kèm migration re-hash hoặc cắt chuỗi), lastHash đọc từ DB trong transaction + row lock | test tái tạo bug 1, 2 → pass |
| 3–4 | Where-context mọi call site + `GET /audit/verify` + export CSV | verify endpoint chạy được trên dữ liệu thật |
| 5–8 | Provider "derived API key" + nối workflow approval → dynsecret lease | e2e: agent xin → duyệt → nhận lease TTL ngắn |
| 9–10 | Revoke cascade (`POST /users/{id}/revoke-all`), `RemoveMember` org, sweeper DROP role | 1 lệnh làm chết mọi credential của 1 user |
| 11–12 | Slack approval (từ notify → action), release pipeline MCP server + README mới, tag v1.0 | release có checksum 3 nền tảng |

## D. Việc nhà đã xong / còn treo

- [x] 2026-08-31: dọn binary khỏi repo (server.exe, valt-cli.exe ~36MB,
  repomix-output.xml, tsbuildinfo) + rule .gitignore (commit 2796380).
- [ ] Lịch sử git vẫn chứa binary cũ — cần `git filter-repo` + force push khi
  chuyển repo (Quyết định 6), làm một lần cho gọn.
- [ ] Transfer repo → `ldphuong-vn`/org khi bắt đầu đợt phát triển chính.
