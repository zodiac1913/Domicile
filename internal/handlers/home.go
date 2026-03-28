package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"Domicile/templates/partials"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// render is a helper to render templ components via Echo
func render(c echo.Context, status int, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return component.Render(c.Request().Context(), c.Response())
}

// isHTMX returns true if the request was made by htmx
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

// HomePage renders the dashboard
func HomePage(c echo.Context) error {
	stats := getDashboardStats()
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.HomeContent(stats))
	}
	return render(c, http.StatusOK, templates.HomePage(stats))
}

// ProductsPage renders the full products view
func ProductsPage(c echo.Context) error {
	var products []models.Product
	database.DB.Select(&products, `
		SELECT p.id, p.name, IFNULL(l.name, '') as location, 
			IFNULL(sc.amount, 0) as quantity,
			p.min_stock_amount,
			IFNULL(sc.best_before_date, '') as expires_at
		FROM products p
		LEFT JOIN locations l ON p.location_id = l.id
		LEFT JOIN stock_current sc ON p.id = sc.product_id
		WHERE p.active = 1
		ORDER BY p.name`)
	locs := getLocations()
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.ProductsContent(products, locs))
	}
	return render(c, http.StatusOK, templates.ProductsPage(products, locs))
}

// CreateProduct handles POST /products — returns a single row for htmx
func CreateProduct(c echo.Context) error {
	name := c.FormValue("name")
	location := c.FormValue("location")
	quantity, _ := strconv.Atoi(c.FormValue("quantity"))
	minStock, _ := strconv.Atoi(c.FormValue("min_stock"))

	var expires string
	if exp := c.FormValue("expires"); exp != "" {
		expires = exp
	}

	// Get or create location
	var locationID int
	err := database.DB.Get(&locationID, "SELECT id FROM locations WHERE name = ?", location)
	if err != nil {
		database.DB.Exec("INSERT INTO locations (name) VALUES (?)", location)
		database.DB.Get(&locationID, "SELECT id FROM locations WHERE name = ?", location)
	}

	// Get default quantity unit
	var quID int
	database.DB.Get(&quID, "SELECT id FROM quantity_units LIMIT 1")

	// Insert product
	result, err := database.DB.Exec(`
		INSERT INTO products (name, location_id, qu_id_purchase, qu_id_stock, qu_factor_purchase_to_stock, min_stock_amount)
		VALUES (?, ?, ?, ?, 1.0, ?)`,
		name, locationID, quID, quID, minStock)
	if err != nil {
		return c.String(http.StatusBadRequest, "Product already exists or invalid data")
	}

	productID, _ := result.LastInsertId()

	// Add stock entry if quantity > 0
	if quantity > 0 {
		stockID := uuid.New().String()
		database.DB.Exec(`
			INSERT INTO stock (product_id, amount, best_before_date, stock_id, location_id)
			VALUES (?, ?, NULLIF(?, ''), ?, ?)`,
			productID, quantity, expires, stockID, locationID)

		database.DB.Exec(`
			INSERT INTO stock_log (product_id, amount, best_before_date, stock_id, transaction_type, location_id)
			VALUES (?, ?, NULLIF(?, ''), ?, 'purchase', ?)`,
			productID, quantity, expires, stockID, locationID)
	}

	p := models.Product{
		ID:             int(productID),
		Name:           name,
		Location:       location,
		Quantity:       quantity,
		MinStockAmount: minStock,
	}
	if expires != "" {
		p.ExpiresAt, _ = time.Parse("2006-01-02", expires)
	}

	return render(c, http.StatusOK, partials.ProductRow(p))
}
