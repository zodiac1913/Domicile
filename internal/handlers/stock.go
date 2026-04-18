package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"Domicile/templates/partials"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// StockOverview renders the stock overview page.
func StockOverview(c echo.Context) error {
	var items []models.StockItem
	database.DB.Select(&items, `
		SELECT p.id as product_id, p.name as product_name,
			IFNULL(l.name,'') as location_name,
			IFNULL(sc.amount, 0) as amount,
			IFNULL(sc.best_before_date,'') as best_before_date,
			IFNULL(p.last_price, 0) as last_price,
			p.min_stock_amount
		FROM products p
		LEFT JOIN locations l ON p.location_id = l.id
		LEFT JOIN stock_current sc ON p.id = sc.product_id
		WHERE p.active = 1
		ORDER BY p.name`)
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.StockContent(items))
	}
	return render(c, http.StatusOK, templates.StockPage(items))
}

// PurchaseForm renders the purchase form.
func PurchaseForm(c echo.Context) error {
	products := getProductsForDropdown()
	locs := getLocationsWithID()
	preselect, _ := strconv.Atoi(c.QueryParam("product"))
	if isHTMX(c) {
		return render(c, http.StatusOK, partials.PurchaseForm(products, locs, preselect))
	}
	return render(c, http.StatusOK, templates.PurchasePage(products, locs, preselect))
}

// CreatePurchase handles the stock purchase.
func CreatePurchase(c echo.Context) error {
	productID, _ := strconv.Atoi(c.FormValue("product_id"))
	amount := parseFloat(c.FormValue("amount"), 1)
	price := optionalFloat(c.FormValue("price"))
	bestBefore := c.FormValue("best_before_date")
	locationID, _ := strconv.Atoi(c.FormValue("location_id"))
	note := c.FormValue("note")

	if productID == 0 {
		return c.String(http.StatusBadRequest, "Select a product")
	}
	if amount <= 0 {
		amount = 1
	}

	// Default location from product if not specified.
	if locationID == 0 {
		database.DB.Get(&locationID, "SELECT location_id FROM products WHERE id = ?", productID)
	}

	stockID := uuid.New().String()
	var priceVal float64
	if price != nil {
		priceVal = *price
	}

	// Create stock entry.
	database.DB.Exec(`
		INSERT INTO stock (product_id, amount, best_before_date, stock_id, price, location_id)
		VALUES (?, ?, NULLIF(?, ''), ?, ?, ?)`,
		productID, amount, bestBefore, stockID, priceVal, locationID)

	// Log the purchase.
	database.DB.Exec(`
		INSERT INTO stock_log (product_id, amount, best_before_date, stock_id, transaction_type, price, location_id, note)
		VALUES (?, ?, NULLIF(?, ''), ?, 'purchase', ?, ?, ?)`,
		productID, amount, bestBefore, stockID, priceVal, locationID, note)

	// Update product's last_price if a price was entered.
	if price != nil && *price > 0 {
		database.DB.Exec("UPDATE products SET last_price = ? WHERE id = ?", *price, productID)
	}

	return StockOverview(c)
}

// ConsumeForm renders the consume form.
func ConsumeForm(c echo.Context) error {
	products := getProductsForDropdown()
	preselect, _ := strconv.Atoi(c.QueryParam("product"))
	if isHTMX(c) {
		return render(c, http.StatusOK, partials.ConsumeForm(products, preselect))
	}
	return render(c, http.StatusOK, templates.ConsumePage(products, preselect))
}

// CreateConsume handles stock consumption.
func CreateConsume(c echo.Context) error {
	productID, _ := strconv.Atoi(c.FormValue("product_id"))
	amount := parseFloat(c.FormValue("amount"), 1)
	spoiled := checkboxVal(c, "spoiled")

	if productID == 0 {
		return c.String(http.StatusBadRequest, "Select a product")
	}

	// Find oldest stock entry (FIFO).
	var stockRow struct {
		ID      int     `db:"id"`
		StockID string  `db:"stock_id"`
		Amount  float64 `db:"amount"`
	}
	err := database.DB.Get(&stockRow, `
		SELECT id, stock_id, amount FROM stock
		WHERE product_id = ? AND amount > 0
		ORDER BY best_before_date ASC, purchased_date ASC
		LIMIT 1`, productID)
	if err != nil {
		return c.String(http.StatusBadRequest, "No stock available for this product")
	}

	consumed := amount
	if consumed > stockRow.Amount {
		consumed = stockRow.Amount
	}

	if stockRow.Amount <= consumed {
		database.DB.Exec("DELETE FROM stock WHERE id = ?", stockRow.ID)
	} else {
		database.DB.Exec("UPDATE stock SET amount = amount - ? WHERE id = ?", consumed, stockRow.ID)
	}

	database.DB.Exec(`
		INSERT INTO stock_log (product_id, amount, stock_id, transaction_type, spoiled)
		VALUES (?, ?, ?, 'consume', ?)`,
		productID, -consumed, stockRow.StockID, spoiled)

	return StockOverview(c)
}

// TransferForm renders the transfer form.
func TransferForm(c echo.Context) error {
	products := getProductsForDropdown()
	locs := getLocationsWithID()
	if isHTMX(c) {
		return render(c, http.StatusOK, partials.TransferForm(products, locs))
	}
	return render(c, http.StatusOK, templates.TransferPage(products, locs))
}

// CreateTransfer handles stock transfer between locations.
func CreateTransfer(c echo.Context) error {
	productID, _ := strconv.Atoi(c.FormValue("product_id"))
	amount := parseFloat(c.FormValue("amount"), 1)
	toLocationID, _ := strconv.Atoi(c.FormValue("to_location_id"))

	if productID == 0 || toLocationID == 0 {
		return c.String(http.StatusBadRequest, "Select a product and destination")
	}

	// Find stock entry to transfer from.
	var stockRow struct {
		ID         int     `db:"id"`
		StockID    string  `db:"stock_id"`
		Amount     float64 `db:"amount"`
		LocationID int     `db:"location_id"`
	}
	err := database.DB.Get(&stockRow, `
		SELECT id, stock_id, amount, location_id FROM stock
		WHERE product_id = ? AND amount > 0
		ORDER BY best_before_date ASC
		LIMIT 1`, productID)
	if err != nil {
		return c.String(http.StatusBadRequest, "No stock available")
	}

	transferred := amount
	if transferred > stockRow.Amount {
		transferred = stockRow.Amount
	}

	// Reduce from source.
	if stockRow.Amount <= transferred {
		database.DB.Exec("DELETE FROM stock WHERE id = ?", stockRow.ID)
	} else {
		database.DB.Exec("UPDATE stock SET amount = amount - ? WHERE id = ?", transferred, stockRow.ID)
	}

	// Add to destination.
	newStockID := uuid.New().String()
	database.DB.Exec(`
		INSERT INTO stock (product_id, amount, stock_id, location_id)
		VALUES (?, ?, ?, ?)`,
		productID, transferred, newStockID, toLocationID)

	// Log transfer.
	database.DB.Exec(`
		INSERT INTO stock_log (product_id, amount, stock_id, transaction_type, location_id)
		VALUES (?, ?, ?, 'transfer_from', ?)`,
		productID, -transferred, stockRow.StockID, stockRow.LocationID)
	database.DB.Exec(`
		INSERT INTO stock_log (product_id, amount, stock_id, transaction_type, location_id)
		VALUES (?, ?, ?, 'transfer_to', ?)`,
		productID, transferred, newStockID, toLocationID)

	return StockOverview(c)
}
