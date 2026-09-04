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
| ~~1–2~~ ✅ | ~~Fix hash-chain~~ (DONE 2026-09-04, xem §E) | test tái tạo bug 1, 2 → pass |
| 3–4 | Where-context mọi call site + `GET /audit/verify` + export CSV | verify endpoint chạy được trên dữ liệu thật |
| 5–8 | Provider "derived API key" + nối workflow approval → dynsecret lease | e2e: agent xin → duyệt → nhận lease TTL ngắn |
| 9–10 | Revoke cascade (`POST /users/{id}/revoke-all`), `RemoveMember` org, sweeper DROP role | 1 lệnh làm chết mọi credential của 1 user |
| 11–12 | Slack approval (từ notify → action), release pipeline MCP server + README mới, tag v1.0 | release có checksum 3 nền tảng |

## D. Việc nhà đã xong / còn treo

- [x] 2026-08-31: dọn binary khỏi repo (server.exe, valt-cli.exe ~36MB,
  repomix-output.xml, tsbuildinfo) + rule .gitignore (commit 2796380).
- [x] 2026-09-03: transfer repo → `ldphuong-vn/vaultsaas`; redirect URL cũ hoạt động.
- [x] 2026-09-04: `git filter-repo` xóa binary khỏi lịch sử + force push
  (master + feat/custom-policy). Repo pack: ~50MB → 1.47MB.
- [ ] Known pre-existing (không phải của đợt này): integration test
  `internal/workflow` fail khi chạy song song trên một DB chung — race
  `CREATE EXTENSION pgcrypto` giữa các schema, helper không gọi
  `database.EnsurePartitions`, và một test dựng Handler với notify store nil.
  Đã xác nhận fail giống hệt trên bản code trước fix hash-chain.

## E. Nhật ký thực hiện — Fix hash-chain (2026-09-04)

Scope: bug 1 + bug 2 trong §A, kèm hai lỗi round-trip chỉ phát hiện được khi
chạy integration test thật.

1. **Hash bao phủ toàn bộ trường** (`hash-chain.go`): preimage v2 = JSON
   canonical của mọi cột (user, action, resource, event_type, status, ip, ua,
   region, metadata, event_time + prev hash). Sửa bất kỳ cột nào trong DB đều
   gãy chuỗi. `computeHashV1` giữ lại chỉ để verify các row cũ (v1 không bao
   phủ ip/metadata/time — ghi chú hạn chế trong code).
2. **Chain head nằm trong DB** (`audit_chain_state`, migration 000042):
   `Log()` chạy một transaction: `SELECT ... FOR UPDATE` state row → gán
   `seq = last_seq + 1` → insert → update state. seq liên tục, đúng thứ tự
   commit, sống qua restart; concurrency an toàn. Bảng `audit_logs` thêm cột
   `seq`, index thường (unique index không đặt được trên bảng partition không
   chứa partition key; tính duy nhất do cursor đảm bảo).
3. **Lỗi round-trip #1 — metadata JSONB**: Postgres render JSONB lại text
   (thêm dấu cách, sort key theo độ dài) làm hash đọc lại ≠ hash lúc ghi.
   Migration 000042 đổi cột sang TEXT (giữ nguyên byte); không có query dùng
   toán tử JSONB trên cột này.
4. **Lỗi round-trip #2 — INET**: `::TEXT` render `10.0.0.1/32`. Thêm
   `canonicalIP()`: lấy entry đầu của XFF, parse bằng netip, chuẩn hóa về dạng
   `addr/NN` trước khi hash; junk → NULL (trước đây junk/líst IP trong XFF
   làm INSERT lỗi cast inet).
5. **Dọn code chết**: xóa `internal/database/audit.go` (`WriteAuditLog` ghi
   row audit NGOÀI chuỗi hash, 0 caller — footgun). `Logger.Log` đổi signature
   trả về `(Entry, error)` kèm id/seq/hash; thêm `AppendChainNoTx` cho caller
   giữ tx riêng. `EnsurePartitions` xuất khẩu từ database package.
6. **Test**: unit (bao phủ trường, phát hiện tamper IP/metadata/time, chuỗi v1
   cũ vẫn verify, `canonicalIP`) + integration trên Postgres thật
   (sống qua restart — tái hiện bug 2; 8 goroutine × 5 log → seq 1..40 liền
   mạch, chuỗi valid; UPDATE thẳng DB → gãy đúng chỗ). Toàn bộ
   `go test ./internal/... ./pkg/...` pass (12 package).
