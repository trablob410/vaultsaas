# Research Report: Go REST API Security Patterns

Date: 2026-03-17

---

## 1. Authorization on Every Handler (chi router)

### Problem
Nothing in Go's type system forces a handler to be wrapped with authz middleware. Omission is a silent runtime bug.

### Patterns (ranked simplest to most robust)

**A. Subrouter grouping — enforce at mount point**
```go
// All routes under /api/v1 MUST pass through authz
r := chi.NewRouter()
r.Use(AuthnMiddleware)        // sets claims in ctx
r.Route("/api/v1", func(r chi.Router) {
    r.Use(AuthzMiddleware)    // fails closed if no claims
    r.Get("/secrets/{id}", GetSecret)
})
// Public routes mounted separately — explicit opt-out, not opt-in
r.Post("/auth/login", Login)
```
Key: public routes live in a separate subrouter; the protected subrouter has authz baked in. Forgetting to add a route to the protected group is visible in code review.

**B. Handler wrapper (compile-time enforced)**
```go
type AuthzHandler func(claims *Claims, w http.ResponseWriter, r *http.Request)

func Protected(ah AuthzHandler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims, ok := ClaimsFromCtx(r.Context())
        if !ok { http.Error(w, "forbidden", 403); return }
        ah(claims, w, r)
    }
}

// Caller MUST provide a function that accepts claims — no bypass possible
r.Get("/secrets/{id}", Protected(GetSecretHandler))
```
Forces every protected endpoint to accept `*Claims`; won't compile otherwise.

**C. Policy-based via Casbin or OPA (for fine-grained RBAC)**
```go
func CasbinMiddleware(e *casbin.Enforcer) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := ClaimsFromCtx(r.Context())
            ok, _ := e.Enforce(claims.Subject, r.URL.Path, r.Method)
            if !ok { http.Error(w, "forbidden", 403); return }
            next.ServeHTTP(w, r)
        })
    }
}
```
OPA is heavier; prefer Casbin for in-process policy without a sidecar.

### Recommendation for Valt
Use **pattern A + B**. Subrouter grouping prevents accidental exposure; `Protected()` wrapper enforces claims presence at handler signature level. Skip Casbin unless role matrix exceeds ~10 permission types.

---

## 2. Envelope Encryption for Config/Credentials at Rest (AES-256-GCM)

### Pattern
```
plaintext JSON → serialize → DEK encrypt (AES-256-GCM) → [nonce||ciphertext||tag]
DEK (32 bytes random) → KEK encrypt (AES-256-GCM) → store alongside blob
KEK → never in app memory at rest; loaded from env/KMS at startup only
```

### Go stdlib implementation
```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/json"
    "io"
)

// EncryptJSON: marshal v, generate ephemeral DEK, encrypt, wrap DEK with KEK
func EncryptJSON(kek []byte, v any) (encDEK, blob []byte, err error) {
    dek := make([]byte, 32)
    if _, err = io.ReadFull(rand.Reader, dek); err != nil { return }

    plain, err := json.Marshal(v)
    if err != nil { return }

    blob, err = aesGCMSeal(dek, plain)
    if err != nil { return }

    encDEK, err = aesGCMSeal(kek, dek)
    return
}

func aesGCMSeal(key, plaintext []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil { return nil, err }
    gcm, err := cipher.NewGCM(block)
    if err != nil { return nil, err }

    // nonce prepended to ciphertext: [12-byte nonce][ciphertext+16-byte tag]
    out := make([]byte, gcm.NonceSize()+len(plaintext)+gcm.Overhead())
    nonce := out[:gcm.NonceSize()]
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil { return nil, err }
    gcm.Seal(out[gcm.NonceSize():gcm.NonceSize():], nonce, plaintext, nil)
    return out, nil
}
```

### Key rules
- **Nonce**: 12 bytes, random per encryption. Never counter-based unless you track state durably.
- **Limit**: max 2^32 encryptions per DEK under random nonces (GCM birthday bound). Rotate DEKs proactively.
- **Additional data (AAD)**: bind ciphertext to its storage context (e.g., `[]byte("provider_config:"+providerID)`) to prevent ciphertext transplanting across records.
- **KEK storage**: env var or AWS/GCP KMS — never in DB. Load once at startup, keep in `[]byte` (not `string` — strings are immutable and harder to zero).
- **DEK rotation**: store `encDEK` alongside blob; re-wrapping DEK with new KEK requires only re-encrypting 32 bytes, not the whole blob.

### Sources
- [Go crypto/cipher pkg docs](https://pkg.go.dev/crypto/cipher)
- [Lambrospetrou Go envelope encryption](https://www.lambrospetrou.com/articles/encryption/)
- [Google Cloud KMS envelope encryption](https://cloud.google.com/kms/docs/envelope-encryption)

---

## 3. Redis Sliding-Window Rate Limiter Wiring

### Algorithm (ZSET-based)
```
key = "rl:{agentID}:{windowBucket}"
ZREMRANGEBYSCORE key -inf (now_ms - window_ms)   // evict expired
count = ZCARD key
if count >= limit → reject
ZADD key now_ms <uuid>
EXPIRE key window_seconds+1
```
All ops in a single Lua script for atomicity.

### Middleware wiring pattern (chi)
```go
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            agentID := AgentIDFromCtx(r.Context()) // set by AuthnMiddleware upstream
            allowed, err := rl.Allow(r.Context(), agentID)
            if err != nil {
                // FAIL-OPEN: log + allow (availability > strict limiting)
                // FAIL-CLOSED: http.Error(w, "service unavailable", 503)
                log.Warn("rate limiter redis error", "err", err)
                next.ServeHTTP(w, r) // fail-open choice
                return
            }
            if !allowed {
                w.Header().Set("Retry-After", "1")
                http.Error(w, "rate limit exceeded", 429)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```
Wire **after** AuthnMiddleware (needs agentID) but **before** business logic handlers.

### Fail-open vs fail-closed decision for Valt
| Scenario | Recommendation |
|---|---|
| Agent secret reads | **Fail-closed** (429/503) — abuse surface too high |
| Agent heartbeat/status | **Fail-open** — don't degrade observability |
| Redis unavailable | Fallback to in-memory limiter (NVIDIA go-ratelimit pattern) |

In-memory fallback: use `golang.org/x/time/rate` (token bucket) per-agent in a `sync.Map`. Evict entries after TTL to avoid unbounded memory growth.

### Per-agent key construction
```go
key := fmt.Sprintf("rl:agent:%s:%s", agentID, planTier) // tier-aware limits
```
Avoid using raw user input in Redis keys — sanitize to `[a-zA-Z0-9_-]` max 64 chars.

### Sources
- [Leapcell Go+Redis sliding window](https://leapcell.io/blog/implementing-a-go-and-redis-powered-sliding-window-rate-limiter)
- [NVIDIA go-ratelimit](https://github.com/NVIDIA/go-ratelimit)
- [Redis rate limiting tutorial](https://redis.io/tutorials/howtos/ratelimiting/)

---

## 4. Filesystem Path Traversal Prevention

### Core invariant
```
canonicalized(base + userInput) must have base as prefix
```

### Go implementation
```go
import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// SafeJoin returns error if resolved path escapes base.
func SafeJoin(base, userInput string) (string, error) {
    // 1. Clean user input before joining — removes ../ sequences
    clean := filepath.Clean(userInput)

    // 2. Join to base
    joined := filepath.Join(base, clean)

    // 3. Resolve symlinks (OS-level canonicalization)
    abs, err := filepath.EvalSymlinks(joined)
    if err != nil {
        // File may not exist yet — use Abs instead for pre-creation checks
        abs, err = filepath.Abs(joined)
        if err != nil { return "", err }
    }

    // 4. Prefix check — MUST use os.PathSeparator suffix to avoid /base-evil bypass
    baseAbs, _ := filepath.Abs(base)
    if !strings.HasPrefix(abs, baseAbs+string(os.PathSeparator)) &&
        abs != baseAbs {
        return "", fmt.Errorf("path traversal detected: %q escapes base %q", abs, baseAbs)
    }
    return abs, nil
}
```

### Depth limit (for recursive scans)
```go
func walkWithDepthLimit(root string, maxDepth int, fn fs.WalkDirFunc) error {
    rootDepth := strings.Count(root, string(os.PathSeparator))
    return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        depth := strings.Count(path, string(os.PathSeparator)) - rootDepth
        if d.IsDir() && depth >= maxDepth {
            return filepath.SkipDir
        }
        return fn(path, d, err)
    })
}
```

### Whitelist pattern (for known extensions only)
```go
var allowedExts = map[string]bool{".json": true, ".yaml": true, ".toml": true}

func isAllowedFile(path string) bool {
    ext := strings.ToLower(filepath.Ext(path))
    return allowedExts[ext]
}
```

### Critical gotchas
- `filepath.Clean` alone is **not** sufficient on Windows — symlinks and UNC paths bypass it. Always `EvalSymlinks` for write operations.
- `strings.HasPrefix(abs, base)` without the trailing separator is exploitable: `/base` prefix matches `/base-evil`.
- On Windows, compare paths case-insensitively: `strings.EqualFold`.
- Never `os.Open(filepath.Join(base, userInput))` without the prefix check — `filepath.Join` calls `Clean` but does not enforce containment.

### Sources
- [OWASP Path Traversal](https://owasp.org/www-community/attacks/Path_Traversal)
- [Go path/filepath pkg docs](https://pkg.go.dev/path/filepath)
- Go stdlib `filepath.EvalSymlinks` / `filepath.Rel` documentation

---

## Unresolved Questions

1. **KEK source for Valt**: Is KEK loaded from env var or a KMS (AWS/GCP)? If env-only, what's the key rotation story? No KMS wiring is documented in current arch.
2. **Rate limiter Redis instance**: Is the rate limiter Redis the same instance as the session/cache Redis, or isolated? Shared instance risks noisy-neighbor eviction affecting rate limit state.
3. **AAD context for envelope encryption**: What stable identifier is available to bind ciphertext to its row? Need a guaranteed immutable column (e.g., provider_config UUID) before AAD can be wired.
4. **Symlink following**: Does Valt's filesystem scanner need to follow symlinks? If yes, `EvalSymlinks` must be called; if no, skip it and stat the link target explicitly.
5. **Authorization policy source**: Static role matrix in code vs. DB-driven? Casbin adoption only justified if policy is dynamic/configurable at runtime.
