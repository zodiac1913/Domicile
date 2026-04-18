package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// render is a helper to render templ components via Echo.
func render(c echo.Context, status int, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return component.Render(c.Request().Context(), c.Response())
}

// isHTMX returns true if the request was made by htmx.
func isHTMX(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

func getDashboardStats() models.DashboardStats {
	var stats models.DashboardStats
	database.DB.Get(&stats.ProductCount, "SELECT COUNT(*) FROM products")
	database.DB.Get(&stats.ShoppingItems, "SELECT COUNT(*) FROM shopping_list WHERE done = 0")
	database.DB.Get(&stats.ExpiringSoon, `
		SELECT COUNT(DISTINCT product_id) FROM stock
		WHERE best_before_date IS NOT NULL
		AND best_before_date > date('now')
		AND best_before_date <= date('now', '+7 days')`)
	database.DB.Get(&stats.BelowMinStock, `
		SELECT COUNT(*) FROM stock_missing_products`)
	return stats
}

func getLocations() []string {
	var locs []string
	database.DB.Select(&locs, "SELECT name FROM locations ORDER BY name")
	return locs
}

// HomePage renders the dashboard.
func HomePage(c echo.Context) error {
	stats := getDashboardStats()
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.HomeContent(stats))
	}
	return render(c, http.StatusOK, templates.HomePage(stats))
}
