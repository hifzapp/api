package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/mohammed-ayoub-dz/hifz/config"
	"github.com/mohammed-ayoub-dz/hifz/middleware"
	authRoutes "github.com/mohammed-ayoub-dz/hifz/routes/auth"
	dailyRoutes "github.com/mohammed-ayoub-dz/hifz/routes/daily"
	hifzRoutes "github.com/mohammed-ayoub-dz/hifz/routes/hifz"
	"github.com/mohammed-ayoub-dz/hifz/routes/me"
	sessionRoutes "github.com/mohammed-ayoub-dz/hifz/routes/sessions"
	"github.com/mohammed-ayoub-dz/hifz/routes/user"
	dashboardRoutes "github.com/mohammed-ayoub-dz/hifz/routes/dashboard"

)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load .env")
	}
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is not configured")
	}

	if os.Getenv("GOOGLE_CLIENT_ID") == "" {
		log.Fatal("GOOGLE_CLIENT_ID is not configured")
	}

	config.Db()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
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

	log.Fatal(app.Listen(":8080"))
}
