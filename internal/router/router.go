package router

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nghiaduong/finai/internal/config"
	"github.com/nghiaduong/finai/internal/handler"
	"github.com/nghiaduong/finai/internal/middleware"
	"github.com/nghiaduong/finai/internal/repository/postgres"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
	"github.com/nghiaduong/finai/internal/service"
	"github.com/nghiaduong/finai/templates/pages"
)

// New creates the router and returns the handler plus a cleanup function
// that must be called on shutdown to stop background goroutines.
func New(cfg *config.Config, pool *pgxpool.Pool) (http.Handler, func()) {
	r := chi.NewRouter()

	// ── Repositories ────────────────────────────────────────
	userRepo := postgres.NewUserRepo(pool)
	sessionRepo := postgres.NewSessionRepo(pool)
	tokenRepo := postgres.NewTokenRepo(pool)
	accountRepo := postgres.NewAccountRepo(pool)
	auditRepo := postgres.NewAuditRepo(pool)
	txnRepo := postgres.NewTransactionRepo(pool)
	budgetRepo := postgres.NewBudgetRepo(pool)
	subRepo := postgres.NewSubscriptionRepo(pool)
	billRepo := postgres.NewBillRepo(pool)
	savingsRepo := postgres.NewSavingsRepo(pool)

	// ── Shared Queries (for budget spent calculation, token revocation) ─
	queries := generated.New(pool)

	// ── Services ────────────────────────────────────────────
	authService := service.NewAuthService(userRepo, sessionRepo, tokenRepo, &cfg.Auth)
	encService := service.NewEncryptionService(cfg.Auth.EncryptionKey)
	plaidService := service.NewPlaidService(&cfg.Plaid)
	txnSyncService := service.NewTransactionSyncService(txnRepo, accountRepo, plaidService, encService)
	accountService := service.NewAccountService(accountRepo, plaidService, encService, auditRepo, txnSyncService)
	txnService := service.NewTransactionService(txnRepo)
	budgetService := service.NewBudgetService(budgetRepo, queries)
	subService := service.NewSubscriptionService(subRepo)
	billService := service.NewBillService(billRepo)
	savingsService := service.NewSavingsService(savingsRepo)
	aiService := service.NewAIService(&cfg.AI)
	insightsService := service.NewInsightsService(queries, aiService)
	netWorthService := service.NewNetWorthService(queries)
	chatRepo := postgres.NewChatRepo(pool)
	chatService := service.NewChatService(chatRepo, aiService, insightsService)
	settingsService := service.NewSettingsService(userRepo, queries)

	// ── Handlers ────────────────────────────────────────────
	handler.SetSecureCookies(cfg.Server.IsProd())
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	txnHandler := handler.NewTransactionHandler(txnService)
	budgetHandler := handler.NewBudgetHandler(budgetService)
	subHandler := handler.NewSubscriptionHandler(subService)
	billHandler := handler.NewBillHandler(billService)
	savingsHandler := handler.NewSavingsHandler(savingsService)
	insightsHandler := handler.NewInsightsHandler(insightsService)
	netWorthHandler := handler.NewNetWorthHandler(netWorthService)
	chatHandler := handler.NewChatHandler(chatService)
	settingsHandler := handler.NewSettingsHandler(settingsService)

	// ── Auth Middleware ─────────────────────────────────────
	authMW := middleware.NewAuth(cfg.Auth.JWTSecret, func(ctx context.Context, jti string) (bool, error) {
		jtiUUID, err := uuid.Parse(jti)
		if err != nil {
			return false, fmt.Errorf("invalid jti format: %w", err)
		}
		return queries.IsTokenRevoked(ctx, jtiUUID)
	})

	// ── Rate Limiter ────────────────────────────────────────
	apiLimiter := middleware.NewRateLimiter(2.0, 120)  // 120 req/min
	appLimiter := middleware.NewRateLimiter(1.0, 60)   // 60 req/min for pages
	authLimiter := middleware.NewRateLimiter(0.17, 10) // 10 req/min

	// ── Global Middleware Chain ──────────────────────────────
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	// Note: chimw.RealIP intentionally omitted — it trusts X-Forwarded-For
	// from any client, which allows rate-limiter bypass via IP spoofing.
	// Render.com sets RemoteAddr correctly at the platform level.
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(chimw.Compress(5))
	r.Use(middleware.BodyLimit(1 << 20)) // 1MB
	r.Use(middleware.ContentType)
	r.Use(middleware.Logger)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.NewCSRF(cfg.Server.IsProd()))

	// ── Health Check ────────────────────────────────────────
	r.Get("/health", healthHandler(pool))
	r.Get("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"alive"}`))
	})
	r.Get("/health/ready", readinessHandler(pool))

	// ── Static Files ────────────────────────────────────────
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// ── Well-Known ──────────────────────────────────────────
	r.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("User-agent: *\nAllow: /\nDisallow: /app/\nDisallow: /api/\nDisallow: /auth/\n"))
	})
	r.Get("/.well-known/security.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Contact: mailto:security@finai.app\nPreferred-Languages: en\n"))
	})

	// ── Public Routes ───────────────────────────────────────
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		csrf := middleware.GetCSRFToken(r)
		templ.Handler(pages.Landing(csrf)).ServeHTTP(w, r)
	})

	// ── Auth Routes (no auth required) ──────────────────────
	r.Route("/auth", func(r chi.Router) {
		r.Use(authLimiter.Handler)

		r.Get("/login", authHandler.LoginPage)
		r.Get("/register", authHandler.RegisterPage)
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		// Logout requires auth to extract JTI for token revocation
		r.With(authMW.RequireAuth).Post("/logout", authHandler.Logout)
	})

	// ── Protected App Routes (return full HTML pages) ───────
	r.Route("/app", func(r chi.Router) {
		r.Use(authMW.RequireAuth)
		r.Use(appLimiter.Handler)

		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			csrf := middleware.GetCSRFToken(r)
			templ.Handler(pages.Dashboard(csrf)).ServeHTTP(w, r)
		})
		r.Get("/accounts", accountHandler.AccountsPage)
		r.Get("/transactions", txnHandler.TransactionsPage)
		r.Get("/subscriptions", subHandler.SubscriptionsPage)
		r.Get("/budgets", budgetHandler.BudgetsPage)
		r.Get("/bills", billHandler.BillsPage)
		r.Get("/savings", savingsHandler.SavingsPage)
		r.Get("/insights", insightsHandler.InsightsPage)
		r.Get("/networth", netWorthHandler.NetWorthPage)
		r.Get("/chat", chatHandler.ChatPage)
		r.Get("/settings", settingsHandler.SettingsPage)

		// Placeholder pages for routes not yet implemented
		r.Get("/onboarding", placeholderPage("Onboarding"))
	})

	// ── Protected API Routes (return HTML partials or JSON) ─
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(authMW.RequireAuth)
		r.Use(apiLimiter.Handler)

		// Accounts
		r.Post("/accounts/link-token", accountHandler.CreateLinkToken)
		r.Post("/accounts/exchange-token", accountHandler.ExchangeToken)
		r.Get("/accounts", accountHandler.ListAccounts)
		r.Post("/accounts/{id}/sync", accountHandler.SyncAccount)
		r.Delete("/accounts/{id}", accountHandler.UnlinkAccount)

		// Transactions
		r.Get("/transactions", txnHandler.ListTransactions)
		r.Post("/transactions", txnHandler.CreateTransaction)
		r.Get("/transactions/search", txnHandler.SearchTransactions)
		r.Get("/transactions/{id}", txnHandler.GetTransaction)
		r.Patch("/transactions/{id}/category", txnHandler.UpdateTransactionCategory)
		r.Patch("/transactions/{id}/notes", txnHandler.UpdateTransactionNotes)
		r.Delete("/transactions/{id}", txnHandler.DeleteTransaction)

		// Budgets
		r.Get("/budgets", budgetHandler.ListBudgets)
		r.Post("/budgets", budgetHandler.CreateBudget)
		r.Delete("/budgets/{id}", budgetHandler.DeleteBudget)

		// Subscriptions
		r.Get("/subscriptions", subHandler.ListSubscriptions)
		r.Post("/subscriptions", subHandler.CreateSubscription)
		r.Post("/subscriptions/{id}/cancel", subHandler.CancelSubscription)
		r.Delete("/subscriptions/{id}", subHandler.DeleteSubscription)

		// Bills
		r.Get("/bills", billHandler.ListBills)
		r.Post("/bills", billHandler.CreateBill)
		r.Post("/bills/{id}/pay", billHandler.MarkBillPaid)
		r.Delete("/bills/{id}", billHandler.DeleteBill)

		// Savings Goals
		r.Get("/savings", savingsHandler.ListSavingsGoals)
		r.Post("/savings", savingsHandler.CreateSavingsGoal)
		r.Post("/savings/{id}/add", savingsHandler.AddFunds)
		r.Post("/savings/{id}/withdraw", savingsHandler.WithdrawFunds)
		r.Delete("/savings/{id}", savingsHandler.DeleteSavingsGoal)

		// Insights
		r.Get("/insights/spending", insightsHandler.SpendingCard)
		r.Get("/insights/income", insightsHandler.IncomeCard)
		r.Get("/insights/subscriptions-total", insightsHandler.SubscriptionsTotalCard)
		r.Get("/insights/savings-progress", insightsHandler.SavingsProgressCard)
		r.Get("/insights/categories", insightsHandler.CategoriesBreakdown)
		r.Get("/insights/forecast", insightsHandler.Forecast)

		// Net Worth
		r.Get("/networth", netWorthHandler.GetLatest)
		r.Post("/networth/snapshot", netWorthHandler.CalculateSnapshot)
		r.Get("/networth/history", netWorthHandler.ListHistory)

		// Chat
		r.Post("/chat", chatHandler.SendMessage)
		r.Get("/chat/history/{sessionId}", chatHandler.GetHistory)
		r.Post("/chat/session", chatHandler.NewSession)

		// Settings
		r.Get("/settings/profile", settingsHandler.GetProfile)
		r.Put("/settings/profile", settingsHandler.UpdateProfile)
		r.Put("/settings/password", settingsHandler.ChangePassword)
	})

	log.Info().Msg("router initialized")
	cleanup := func() {
		apiLimiter.Stop()
		appLimiter.Stop()
		authLimiter.Stop()
	}
	return r, cleanup
}

func placeholderPage(title string) http.HandlerFunc {
	safe := html.EscapeString(title)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><body><h1>` + safe + `</h1><p>Coming soon...</p><a href="/app/dashboard">Back to Dashboard</a></body></html>`))
	}
}

func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "up"
		if err := pool.Ping(r.Context()); err != nil {
			dbStatus = "down"
		}

		w.Header().Set("Content-Type", "application/json")
		if dbStatus == "down" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		w.Write([]byte(`{"status":"` + dbStatus + `","service":"finai-api"}`))
	}
}

func readinessHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"ready":false}`))
			return
		}
		w.Write([]byte(`{"ready":true}`))
	}
}
