// project_name/api.go
package project_name

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mviner000/eyygo/project_name/notes"
	auth "github.com/mviner000/eyygo/src/auth"
)

// SetupAPIRoutes sets up all the API routes under the /api prefix
func SetupAPIRoutes(app *fiber.App) {
	// Group all API routes under /api
	apiGroup := app.Group("/api")

	// Public API routes
	apiGroup.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to project_name API!")
	})

	// Token generation endpoint
	apiGroup.Post("/token/pair", auth.GenerateTokenPairHandler)

	// Group notes-related routes under /api/notes
	noteGroup := apiGroup.Group("/notes")

	// Call the function to set up note routes
	notes.SetupNoteRoutes(noteGroup) // Pass the noteGroup to the SetupNoteRoutes function
}
