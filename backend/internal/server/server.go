package server

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"apex/internal/auth"
	"apex/internal/cache"
	"apex/internal/config"
	"apex/internal/features"
	"apex/internal/goals"
	"apex/internal/handler"
	"apex/internal/iracing"
	"apex/internal/locales"
	"apex/internal/metrics"
	"apex/internal/middleware"
	"apex/internal/racing"
	"apex/internal/secretbox"
	"apex/internal/setups"
)

// iracingFlag gates all iRacing-OAuth-dependent routes.
const iracingFlag = "iracing_oauth"

// New builds the application's HTTP router with all routes and middleware.
func New(cfg *config.Config, db *sql.DB) http.Handler {
	authSvc := auth.NewService(db)
	// Redis cache is fail-open: an empty REDIS_ADDR (or a downed Redis) just
	// means every read falls through to the DB.
	redisCache := cache.New(cfg.RedisAddr)
	featuresSvc := features.NewService(db).WithCache(redisCache)

	h := handler.New(db, authSvc)
	h.CookieSecure = cfg.CookieSecure
	h.Racing = buildRacing(cfg, db)
	h.Features = featuresSvc
	h.Setups = setups.New(db)
	h.Goals = goals.New(db)
	h.Locales = locales.New(db)
	h.Cache = redisCache
	h.DeveloperKey = cfg.DeveloperKey

	// requireIRacing 404s OAuth-dependent routes when the feature flag is off.
	requireIRacing := requireFeature(featuresSvc, iracingFlag)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(metrics.Middleware)
	r.Use(middleware.CORS(cfg.CORSOrigin))
	r.Use(middleware.Auth(authSvc))

	// Prometheus exposition. At the root (not under /api) and not proxied by the
	// frontend nginx, so it's only reachable inside apex-net for scraping.
	r.Handle("/metrics", metrics.Handler())

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", h.Health)
		r.Get("/features", h.ListFeatures)
		r.Post("/fuel/plan", h.FuelPlan)

		// Backend-driven i18n: the language list + bundles live in the DB.
		r.Get("/locales", h.ListLocales)
		r.Get("/locales/{code}", h.GetLocale)

		// Cockpit dev-overlay: gated by the developer cookie matching
		// DEVELOPER_KEY (each handler calls devAuth → 404 otherwise). No
		// feature-flag gate here — that would be a chicken-and-egg, since the
		// toggle endpoint is how you'd flip flags in the first place.
		r.Get("/features/all", h.AllFeatures)
		r.Put("/features/{key}", h.ToggleFeature)
		r.Get("/health/cockpit", h.HealthCockpit)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Post("/logout", h.Logout)
			r.Get("/me", h.Me)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAuth)
				r.Patch("/profile", h.UpdateProfile)
				r.Put("/avatar", h.UpdateAvatar)
				r.Post("/password", h.ChangePassword)
			})
		})

		// Planner works WITHOUT iRacing (catalog is seeded); only the live sync
		// from iRacing is gated behind the flag.
		r.Route("/planner", func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.With(requireIRacing).Post("/catalog/sync", h.SyncCatalog)

			r.Get("/cars", h.ListCars)
			r.Put("/cars/{id}", h.SetCarOwned)
			r.Get("/tracks", h.ListTracks)
			r.Put("/tracks/{id}", h.SetTrackOwned)
			r.Get("/series", h.ListSeries)
			r.Put("/series/{id}", h.SetSeriesFavorite)

			r.Get("/season", h.SeasonView)
			r.Put("/season/plan", h.SetRacePlanned)
		})

		// Setups showroom — save private setups and share them publicly. Works
		// without iRacing (uses the seeded catalog for car/track names).
		r.Route("/setups", func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Get("/", h.ListSetups)
			r.Post("/", h.CreateSetup)
			r.Post("/generate", h.GenerateSetup)
			r.Post("/generate/pack", h.GenerateSetupPack)
			r.Get("/{id}", h.GetSetup)
			r.Put("/{id}/public", h.SetSetupPublic)
			r.Delete("/{id}", h.DeleteSetup)
		})

		// Goal tracker — personal numeric goals with progress.
		r.Route("/goals", func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Get("/", h.ListGoals)
			r.Post("/", h.CreateGoal)
			r.Put("/{id}", h.UpdateGoal)
			r.Delete("/{id}", h.DeleteGoal)
		})

		// iRacing OAuth linking + live data — gated behind the feature flag.
		r.Route("/iracing", func(r chi.Router) {
			r.With(middleware.RequireAuth).Get("/", h.IRacingStatus)

			r.Group(func(r chi.Router) {
				r.Use(requireIRacing)
				r.Get("/callback", h.CallbackIRacing)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireAuth)
					r.Get("/authorize", h.AuthorizeIRacing)
					r.Delete("/", h.UnlinkIRacing)
					r.Get("/stats", h.IRacingStats)
					r.Post("/sync", h.IRacingSync)
				})
			})
		})

		r.Route("/drivers", func(r chi.Router) {
			r.Use(requireIRacing)
			r.Use(middleware.RequireAuth)
			r.Get("/search", h.SearchDrivers)
			r.Get("/{custId}", h.DriverProfile)
		})

		r.Route("/compare", func(r chi.Router) {
			r.Use(requireIRacing)
			r.Use(middleware.RequireAuth)
			r.Get("/categories", h.CompareCategories)
			r.Get("/cars", h.CompareCars)
			r.Get("/tracks", h.CompareTracks)
		})
	})

	return r
}

// requireFeature is chi middleware that 404s a route when a flag is off.
func requireFeature(f *features.Service, key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if f == nil || !f.Enabled(r.Context(), key) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"feature unavailable"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// buildRacing always returns a service. The planner works with the DB alone;
// the OAuth-dependent methods gate themselves when encryption/client settings
// are absent (see racing.Service.oauthReady).
func buildRacing(cfg *config.Config, db *sql.DB) *racing.Service {
	var box *secretbox.Box
	if cfg.EncryptionKey != "" {
		if b, err := secretbox.New(cfg.EncryptionKey); err == nil {
			box = b
		} else {
			log.Printf("iRacing OAuth disabled (bad APP_ENCRYPTION_KEY: %v)", err)
		}
	}
	oauth := iracing.OAuthConfig{
		ClientID:     cfg.IRacingClientID,
		ClientSecret: cfg.IRacingClientSecret,
		RedirectURI:  cfg.IRacingRedirectURI,
	}
	return racing.NewService(db, box, racing.DefaultFactory, oauth)
}
