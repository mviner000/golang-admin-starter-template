package cmd

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	conf "github.com/mviner000/eyygo/src"
	"github.com/mviner000/eyygo/src/core"

	"github.com/spf13/cobra"
)

var ShowRoutesCmd = &cobra.Command{
	Use:   "show_routes [app_name]",
	Short: "Display all routes in the application or for a specific app",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app := fiber.New()    // Create a new Fiber app
		core.SetupRoutes(app) // Set up all routes

		if len(args) == 1 {
			appName := args[0]
			if isAppInstalled(appName) {
				core.LogAppRoutes(app, appName)
			} else {
				fmt.Printf("Error: App '%s' is not installed or does not exist.\n", appName)
			}
		} else {
			core.LogAvailableRoutes(app) // Log all routes
		}
	},
}

func isAppInstalled(appName string) bool {
	// Check if the app is installed
	for _, installedApp := range conf.GetSettings().InstalledApps {
		if installedApp == appName {
			return true
		}
	}
	return false
}
