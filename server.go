package main

import (
	"Domicile/internal/database"
	"Domicile/internal/handlers"
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize database
	if err := database.Init("data"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Static files
	e.Static("/static", "static")

	// Pages
	e.GET("/", handlers.HomePage)
	e.GET("/products", handlers.ProductsPage)

	// Product actions (htmx endpoints)
	e.POST("/products", handlers.CreateProduct)
	e.POST("/products/:id/consume", handlers.ConsumeProduct)
	e.POST("/products/:id/add", handlers.AddProduct)
	e.DELETE("/products/:id", handlers.DeleteProduct)

	// Dashboard partials (htmx)
	e.GET("/partials/expiring-soon", handlers.ExpiringSoonPartial)
	e.GET("/partials/shopping-preview", handlers.ShoppingPreviewPartial)

	// Find available port starting at 1323
	port := findAvailablePort(1323, 1423)
	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost:%d", port)

	log.Printf("Domicile running at %s", url)
	openBrowser(url)

	e.Logger.Fatal(e.Start(addr))
}

func findAvailablePort(start, max int) int {
	for port := start; port <= max; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port
		}
	}
	log.Fatalf("No available port found in range %d-%d", start, max)
	return 0
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		log.Printf("Open %s in your browser", url)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Could not open browser: %v — open %s manually", err, url)
	}
}
