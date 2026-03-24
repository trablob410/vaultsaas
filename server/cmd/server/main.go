package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valt-dev/valt/server/internal/agent"
	"github.com/valt-dev/valt/server/internal/audit"
	"github.com/valt-dev/valt/server/internal/auth"
	"github.com/valt-dev/valt/server/internal/config"
	"github.com/valt-dev/valt/server/internal/consent"
	"github.com/valt-dev/valt/server/internal/database"
	"github.com/valt-dev/valt/server/internal/dynsecret"
	custommiddleware "github.com/valt-dev/valt/server/internal/middleware"
	"github.com/valt-dev/valt/server/internal/notify"
	"github.com/valt-dev/valt/server/internal/org"
	"github.com/valt-dev/valt/server/internal/policy"
	"github.com/valt-dev/valt/server/internal/project"
	"github.com/valt-dev/valt/server/internal/ratelimit"
	"github.com/valt-dev/valt/server/internal/scanner"
	"github.com/valt-dev/valt/server/internal/usage"
	"github.com/valt-dev/valt/server/internal/vault"
	"github.com/valt-dev/valt/server/internal/workflow"
	"github.com/valt-dev/valt/server/internal/workspace"
	"github.com/valt-dev/valt/server/pkg/apierror"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	database.StartPartitionManager(ctx, pool)

	// Load JWT keys
	privPEM, err := os.ReadFile(cfg.JWTPrivateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read JWT private key: %v", err)
	}
	pubPEM, err := os.ReadFile(cfg.JWTPublicKeyPath)
	if err != nil {
		log.Fatalf("Failed to read JWT public key: %v", err)
	}

	jwtMgr, err := auth.NewJWTManager(privPEM, pubPEM)
	if err != nil {
		log.Fatalf("Failed to init JWT manager: %v", err)
	}

	// Init MinIO storage
	storage, err := vault.NewMinIOStorage(
		cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey,
		cfg.MinioBucket, cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatalf("Failed to init MinIO storage: %v", err)
	}

	// Decode master key for envelope encryption
	masterKey, err := cfg.MasterKey()
	if err != nil {
		log.Fatalf("Failed to load master key: %v", err)
	}

	// Init services
	auditLogger := audit.NewLogger(pool)
	emailSender := notify.NewEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)
	var emailNotifier notify.Notifier
	if emailSender != nil {
		emailNotifier = emailSender
	}
	notifySvc := notify.NewService(emailNotifier)

	authHandler := auth.NewHandler(pool, jwtMgr, cfg)
	vaultService := vault.NewService(pool, storage)
	vaultHandler := vault.NewHandler(vaultService, masterKey)

	workflowSvc := workflow.NewService(pool, cfg.PolicyEnforcementV2Enabled)
	credMgr := workflow.NewCredentialManager(pool)
	workflowHandler := workflow.NewHandler(workflowSvc, credMgr, vaultService, auditLogger, notifySvc, masterKey, pool)

	auditHandler := audit.NewHandler(pool)

	consentSvc := consent.NewService(pool)
	consentHandler := consent.NewHandler(consentSvc)

	orgSvc := org.NewService(pool)
	orgHandler := org.NewHandler(orgSvc)
	workspaceSvc := workspace.NewService(pool)
	workspaceHandler := workspace.NewHandler(workspaceSvc)
	projectSvc := project.NewService(pool)
	projectHandler := project.NewHandler(projectSvc)
	agentSvc := agent.NewService(pool)
	agentHandler := agent.NewHandler(agentSvc)
	policySvc := policy.NewService(pool)
	policyHandler := policy.NewHandler(policySvc)
	scannerSvc := scanner.NewService(pool)
	scannerHandler := scanner.NewHandler(scannerSvc, pool)
	dynSvc := dynsecret.NewService(pool, masterKey)
	dynSvc.StartExpiryWorker(ctx)
	dynHandler := dynsecret.NewHandler(dynSvc, pool)

	// Redis rate limiter (optional — skip if REDIS_URL not set)
	var agentRateLimiter *ratelimit.RedisLimiter
	if cfg.RedisURL != "" {
		rl, err := ratelimit.NewRedisLimiter(cfg.RedisURL)
		if err != nil {
			log.Printf("Warning: Redis rate limiter init failed: %v", err)
		} else {
			agentRateLimiter = rl
			defer agentRateLimiter.Close()
		}
	}
	// agentRateLimiter is wired into the authenticated route group below

	// Usage tracking
	usageTracker := usage.NewTracker(pool)
	usageHandler := usage.NewHandler(usageTracker)

	// Rate limiters
	loginLimiter := custommiddleware.NewRateLimiter(5, 1*time.Minute)
	apiLimiter := custommiddleware.NewRateLimiter(100, 1*time.Minute)

	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(cfg.CORSOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", healthHandler(pool))

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(custommiddleware.SecurityHeaders)

		r.Route("/auth", func(r chi.Router) {
			r.Use(loginLimiter.IPMiddleware())
			r.Mount("/", authHandler.Routes())
		})

		// Dual-auth routes: accept either user JWT or agent bearer token
		// These routes allow both users and AI agents to read/list data
		r.Group(func(r chi.Router) {
			r.Use(dualAuthMiddleware(jwtMgr, agentSvc))
			r.Use(apiLimiter.Middleware())
			if agentRateLimiter != nil {
				r.Use(agentRateLimiter.Middleware(60))
			}

			// GET secrets - readable by both users and agents
			r.Get("/secrets", vaultHandler.ListSecrets)
			r.Get("/secrets/{id}", vaultHandler.GetSecret)

			// Create/modify secrets - JWT only (kept in separate group below)
			// Posted as separate routes with JWT-only auth

			// Access request flow - readable by both
			r.Post("/secrets/{secret_id}/access-requests", workflowHandler.CreateRequest)
			r.Get("/access-requests/{request_id}", workflowHandler.GetRequest)
			r.Get("/credentials/{request_id}", workflowHandler.GetCredential)
			r.Post("/credentials/{request_id}/revoke", workflowHandler.RevokeCredential)
		})

		// JWT-only authenticated routes: for write operations and user-specific operations
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware(jwtMgr))
			r.Use(apiLimiter.Middleware())
			if agentRateLimiter != nil {
				r.Use(agentRateLimiter.Middleware(60))
			}

			// Secret write operations - JWT/users only
			r.Post("/secrets", vaultHandler.CreateSecret)
			r.Put("/secrets/{id}", vaultHandler.UpdateSecret)
			r.Delete("/secrets/{id}", vaultHandler.DeleteSecret)

			// Access request management - JWT/users only
			r.Get("/access-requests", workflowHandler.ListPending)
			r.Post("/access-requests/{request_id}/approve", workflowHandler.Approve)
			r.Post("/access-requests/{request_id}/reject", workflowHandler.Reject)

			// Audit and user-specific operations
			r.Mount("/audit", auditHandler.Routes())
			r.Mount("/consent", consentHandler.Routes())
			r.Mount("/orgs", orgHandler.Routes())
			r.Mount("/orgs/{org_id}/workspaces", workspaceHandler.Routes())
			projectHandler.RegisterRoutes(r)
			agentHandler.RegisterRoutes(r)
			policyHandler.RegisterRoutes(r)
			scannerHandler.RegisterRoutes(r)
			dynHandler.RegisterRoutes(r)
			usageHandler.RegisterRoutes(r)
		})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Valt server starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}

func dualAuthMiddleware(jwtMgr *auth.JWTManager, agentSvc *agent.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				apierror.Unauthorized(w, "missing authorization header")
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				apierror.Unauthorized(w, "invalid authorization format")
				return
			}
			token := parts[1]

			// Try JWT first (no DB lookup)
			if userID, err := jwtMgr.ValidateAccessToken(token); err == nil {
				ctx := auth.WithUserID(r.Context(), userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Fallback: try agent token
			if agentToken, err := agentSvc.ValidateToken(r.Context(), token); err == nil && agentToken != nil {
				ctx := agent.WithAgentID(r.Context(), agentToken.AgentID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			apierror.Unauthorized(w, "invalid or expired token")
		})
	}
}

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database"`
}

func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		if err := pool.Ping(r.Context()); err != nil {
			dbStatus = "unavailable"
		}

		status := "ok"
		statusCode := http.StatusOK
		if dbStatus != "ok" {
			status = "degraded"
			statusCode = http.StatusServiceUnavailable
		}

		resp := healthResponse{
			Status:    status,
			Service:   "valt-server",
			Version:   "0.1.0",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Database:  dbStatus,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Failed to encode health response: %v", err)
		}
	}
}
