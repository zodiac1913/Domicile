package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates/partials"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func getProductByID(id int) (models.Product, error) {
	var p models.Product
	err := database.DB.Get(&p, `
		SELECT p.id, p.name, IFNULL(l.name, '') as location,
			IFNULL(sc.amount, 0) as quantity,
			p.min_stock_amount,
			IFNULL(sc.best_before_date, '') as expires_at
		FROM products p
		LEFT JOIN locations l ON p.location_id = l.id
		LEFT JOIN stock_current sc ON p.id = sc.product_id
		WHERE p.id = ?`, id)
	return p, err
}

// ConsumeProduct decrements quantity by 1
func ConsumeProduct(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	// Find the oldest stock entry for this product
	var stockRow struct {
		ID      int     `db:"id"`
		StockID string  `db:"stock_id"`
		Amount  float64 `db:"amount"`
	}
	err := database.DB.Get(&stockRow, `
		SELECT id, stock_id, amount FROM stock
		WHERE product_id = ? AND amount > 0
		ORDER BY best_before_date ASC, purchased_date ASC
		LIMIT 1`, id)
	if err == nil {
		if stockRow.Amount <= 1 {
			database.DB.Exec("DELETE FROM stock WHERE id = ?", stockRow.ID)
		} else {
			database.DB.Exec("UPDATE stock SET amount = amount - 1 WHERE id = ?", stockRow.ID)
		}
		database.DB.Exec(`
			INSERT INTO stock_log (product_id, amount, stock_id, transaction_type)
			VALUES (?, -1, ?, 'consume')`, id, stockRow.StockID)
	}

	p, err := getProductByID(id)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}
	return render(c, http.StatusOK, partials.ProductRow(p))
}

// AddProduct increments quantity by 1
func AddProduct(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	// Get product's default location
	var locationID int
	database.DB.Get(&locationID, "SELECT location_id FROM products WHERE id = ?", id)

	stockID := uuid.New().String()
	database.DB.Exec(`
		INSERT INTO stock (product_id, amount, stock_id, location_id)
		VALUES (?, 1, ?, ?)`, id, stockID, locationID)
	database.DB.Exec(`
		INSERT INTO stock_log (product_id, amount, stock_id, transaction_type, location_id)
		VALUES (?, 1, ?, 'purchase', ?)`, id, stockID, locationID)

	p, err := getProductByID(id)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}
	return render(c, http.StatusOK, partials.ProductRow(p))
}

// DeleteProduct removes a product
func DeleteProduct(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM stock WHERE product_id = ?", id)
	database.DB.Exec("DELETE FROM stock_log WHERE product_id = ?", id)
	database.DB.Exec("DELETE FROM products WHERE id = ?", id)
	return c.NoContent(http.StatusOK)
}

// ExpiringSoonPartial returns the expiring-soon list for htmx
func ExpiringSoonPartial(c echo.Context) error {
	var products []models.Product
	database.DB.Select(&products, `
		SELECT p.id, p.name, IFNULL(l.name, '') as location,
			IFNULL(sc.amount, 0) as quantity,
			p.min_stock_amount,
			IFNULL(sc.best_before_date, '') as expires_at
		FROM products p
		LEFT JOIN locations l ON p.location_id = l.id
		LEFT JOIN stock_current sc ON p.id = sc.product_id
		WHERE sc.best_before_date IS NOT NULL
			AND sc.best_before_date > date('now')
			AND sc.best_before_date <= date('now', '+7 days')
		ORDER BY sc.best_before_date ASC`)
	return render(c, http.StatusOK, partials.ExpiringSoon(products))
}

// ShoppingPreviewPartial returns the shopping list preview for htmx
func ShoppingPreviewPartial(c echo.Context) error {
	var items []models.ShoppingItem
	database.DB.Select(&items, `
		SELECT sl.id, IFNULL(p.name, sl.note) as name, sl.amount as quantity, sl.done
		FROM shopping_list sl
		LEFT JOIN products p ON sl.product_id = p.id
		WHERE sl.done = 0
		ORDER BY sl.row_created_timestamp DESC
		LIMIT 10`)
	return render(c, http.StatusOK, partials.ShoppingPreview(items))
}
