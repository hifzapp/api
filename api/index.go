package handler

import (
	"net/http"
	"os"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"github.com/mohammed-ayoub-dz/hifz/config"
	"github.com/mohammed-ayoub-dz/hifz/middleware"

	authRoutes "github.com/mohammed-ayoub-dz/hifz/routes/auth"
	dailyRoutes "github.com/mohammed-ayoub-dz/hifz/routes/daily"
	dashboardRoutes "github.com/mohammed-ayoub-dz/hifz/routes/dashboard"
	hifzRoutes "github.com/mohammed-ayoub-dz/hifz/routes/hifz"
	"github.com/mohammed-ayoub-dz/hifz/routes/me"
	sessionRoutes "github.com/mohammed-ayoub-dz/hifz/routes/sessions"
	"github.com/mohammed-ayoub-dz/hifz/routes/user"
)

var (
	app      *fiber.App
	initOnce sync.Once
)

func getApp() *fiber.App {
	initOnce.Do(func() {

		if os.Getenv("JWT_SECRET") == "" {
			panic("JWT_SECRET is not configured")
		}

		if os.Getenv("GOOGLE_CLIENT_ID") == "" {
			panic("GOOGLE_CLIENT_ID is not configured")
		}

		config.Db()

		app = fiber.New()

		app.Use(cors.New(cors.Config{
			AllowOrigins: []string{
				"https://hifzapp.netlify.app",
			},
			AllowMethods: []string{
				"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
			},
			AllowHeaders: []string{
				"Origin",
				"Content-Type",
				"Accept",
				"Authorization",
				"X-Requested-With",
			},
			AllowCredentials: true,
		}))

		api := app.Group("/api/v1")

		authRoutes.RegisterRoutes(api)

		protectedAPI := api.Group(
			"/app",
			middleware.Protected(),
		)

		hifzRoutes.RegisterRoutes(protectedAPI)
		sessionRoutes.RegisterRoutes(protectedAPI)
		dailyRoutes.RegisterRoutes(protectedAPI)
		user.RegisterRoutes(protectedAPI)
		me.RegisterRoutes(protectedAPI)
		dashboardRoutes.RegisterRoutes(protectedAPI)
	})

	return app
}

func Handler(w http.ResponseWriter, r *http.Request) {
    r.RequestURI = r.URL.String()
    adaptor.FiberApp(getApp())(w, r)
}