package core

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/mviner000/eyymi/config"
	"github.com/mviner000/eyymi/eyygo/constants"
	"github.com/mviner000/eyymi/eyygo/core/decorators"
	"github.com/mviner000/eyymi/eyygo/reverb"
	"github.com/mviner000/eyymi/project_name"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	appLogger *log.Logger
	db        *gorm.DB
)

// Define an interface that all apps should implement
type App interface {
	SetupRoutes(app *fiber.App)
}

func init() {
	if appLogger == nil {
		appLogger = log.New(os.Stdout, "CORE: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	// Set the default time zone
	loc, err := time.LoadLocation(project_name.AppSettings.TimeZone)
	if err != nil {
		appLogger.Fatalf("Invalid time zone: %v", err)
	}
	time.Local = loc

	// Log the time zone if DEBUG is true
	if project_name.AppSettings.Debug {
		config.DebugLogf("Time zone set to: %s", project_name.AppSettings.TimeZone)
	}

	db, err = gorm.Open(sqlite.Open(project_name.AppSettings.Database.Name), &gorm.Config{})
	if err != nil {
		appLogger.Fatalf("Failed to connect to database: %v", err)
	}
}

func ReloadSettings() {
	project_name.LoadSettings() // Reload settings using the function from project_name
	log.Println("Settings reloaded")
}

func RunCommand() {
	ReloadSettings() // Ensure settings are reloaded at the start

	nodeEnv := os.Getenv("NODE_ENV")
	isProduction := nodeEnv == "production"

	if isProduction {
		config.DebugLogf("Running in production mode")
		setupProductionServer()
	} else {
		config.DebugLogf("Running in development mode")
		setupDevelopmentServer()
	}
}

// NewApp initializes and returns a new Fiber application
func NewApp() *fiber.App {
	app := fiber.New(fiber.Config{
		Views:       html.New("./", ".html"),
		ReadTimeout: 5 * time.Second,
	})

	setupMiddleware(app)
	SetupRoutes(app)

	// Set up WebSocket route
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		defer c.Close()
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, msg)
			if err != nil {
				break
			}
		}
	}))

	// Log that the WebSocket route is set up
	log.Println(constants.ColorYellow + "WebSocket route set up at /ws" + constants.ColorReset)

	return app
}

func customCORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

func setupMiddleware(app *fiber.App) {
	// Recover middleware
	app.Use(recover.New())

	// Logger middleware
	app.Use(logger.New(logger.Config{
		Format:     "${time} ${status} - ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   project_name.AppSettings.TimeZone,
	}))

	// Use custom CORS middleware
	app.Use(customCORS())

	// Custom middlewares
	// app.Use(decorators.RequireHTTPS())
	app.Use(decorators.Logger())
	app.Use(decorators.Throttle(100, 60)) // 100 requests per minute
	app.Use(decorators.DatabaseTransaction(db))
}

func setupDevelopmentServer() {
	httpPort := os.Getenv("HTTP_PORT")
	wsPort := os.Getenv("WS_PORT")

	if httpPort == "" {
		httpPort = "8000"
	}

	if wsPort == "" {
		wsPort = "3333"
	}

	// Set up HTTP server
	go func() {
		app := fiber.New(fiber.Config{
			Views:       html.New("./", ".html"),
			ReadTimeout: 5 * time.Second,
		})

		setupMiddleware(app)
		reverb.SetupWebSocket(app)
		SetupRoutes(app)

		if project_name.AppSettings.Debug {
			appLogger.Printf("Development server started on http://127.0.0.1:%s", httpPort)
		}

		err := app.Listen(":" + httpPort)
		if err != nil {
			appLogger.Fatalf("Failed to start development server: %v", err)
		}
	}()

	// Set up WebSocket server
	go func() {
		app := fiber.New(fiber.Config{
			Views:       html.New("./", ".html"),
			ReadTimeout: 5 * time.Second,
		})

		setupMiddleware(app)
		reverb.SetupWebSocket(app)
		SetupRoutes(app)

		if project_name.AppSettings.Debug {
			appLogger.Printf("WebSocket server started on ws://127.0.0.1:%s", wsPort)
		}

		err := app.Listen(":" + wsPort)
		if err != nil {
			appLogger.Fatalf("Failed to start WebSocket server: %v", err)
		}
	}()

	// Block forever
	select {}
}

func setupProductionServer() {
	httpPort := os.Getenv("HTTP_PORT")
	wsPort := os.Getenv("WS_PORT")

	if httpPort == "" {
		httpPort = "8000"
	}

	if wsPort == "" {
		wsPort = "3333"
	}

	// Set up HTTP server
	go func() {
		app := fiber.New(fiber.Config{
			Views:        html.New("./", ".html"),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		})

		setupMiddleware(app)
		reverb.SetupWebSocket(app)
		SetupRoutes(app)

		certFile := project_name.AppSettings.CertFile
		keyFile := project_name.AppSettings.KeyFile

		// Check for wildcard in AllowedOrigins
		for _, origin := range project_name.AppSettings.AllowedOrigins {
			if origin == "*" {
				appLogger.Println(constants.ColorRed + "WARNING: Using wildcard '*' in AllowedOrigins in production is not recommended!" + constants.ColorReset)
				break
			}
		}

		if project_name.AppSettings.Debug {
			appLogger.Printf("Allowed origins: %v", project_name.AppSettings.AllowedOrigins)
		}

		if certFile != "" && keyFile != "" {
			if project_name.AppSettings.Debug {
				appLogger.Printf("Starting HTTPS server on port %s", httpPort)
			}
			err := app.ListenTLS(":"+httpPort, certFile, keyFile)
			if err != nil {
				appLogger.Fatalf("Failed to start HTTPS server: %v", err)
			}
		} else {
			if project_name.AppSettings.Debug {
				appLogger.Printf("Starting HTTP server on port %s", httpPort)
			}
			err := app.Listen(":" + httpPort)
			if err != nil {
				appLogger.Fatalf("Failed to start HTTP server: %v", err)
			}
		}
	}()

	// Set up WebSocket server
	go func() {
		app := fiber.New(fiber.Config{
			Views:        html.New("./", ".html"),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		})

		setupMiddleware(app)
		reverb.SetupWebSocket(app)
		SetupRoutes(app)

		certFile := project_name.AppSettings.CertFile
		keyFile := project_name.AppSettings.KeyFile

		// Check for wildcard in AllowedOrigins
		for _, origin := range project_name.AppSettings.AllowedOrigins {
			if origin == "*" {
				appLogger.Println(constants.ColorRed + "WARNING: Using wildcard '*' in AllowedOrigins in production is not recommended!" + constants.ColorReset)
				break
			}
		}

		if project_name.AppSettings.Debug {
			appLogger.Printf("Allowed origins: %v", project_name.AppSettings.AllowedOrigins)
		}

		if certFile != "" && keyFile != "" {
			if project_name.AppSettings.Debug {
				appLogger.Printf("Starting HTTPS WebSocket server on port %s", wsPort)
			}
			err := app.ListenTLS(":"+wsPort, certFile, keyFile)
			if err != nil {
				appLogger.Fatalf("Failed to start HTTPS WebSocket server: %v", err)
			}
		} else {
			if project_name.AppSettings.Debug {
				appLogger.Printf("Starting HTTP WebSocket server on port %s", wsPort)
			}
			err := app.Listen(":" + wsPort)
			if err != nil {
				appLogger.Fatalf("Failed to start HTTP WebSocket server: %v", err)
			}
		}
	}()

	// Block forever
	select {}
}
