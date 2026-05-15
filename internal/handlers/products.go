package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"Domicile/templates/partials"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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

func getProductDetailByID(id int) (models.ProductDetail, error) {
	var p models.ProductDetail
	err := database.DB.Get(&p, `
		SELECT p.id, p.name, IFNULL(p.description,'') as description,
			p.location_id, IFNULL(l.name, '') as location_name,
			p.product_group_id, p.qu_id_purchase, p.qu_id_stock,
			p.qu_id_consume, p.qu_id_price, p.qu_factor_purchase_to_stock,
			IFNULL(p.barcode,'') as barcode, p.min_stock_amount,
			p.default_best_before_days, p.default_best_before_days_after_open,
			p.default_best_before_days_after_freezing, p.default_best_before_days_after_thawing,
			IFNULL(p.picture_file_name,'') as picture_file_name,
			p.enable_tare_weight_handling, p.tare_weight,
			p.not_check_stock_fulfillment_for_recipes, p.parent_product_id,
			p.calories, p.cumulate_min_stock_amount_of_sub_products,
			p.due_type, p.quick_consume_amount, p.quick_open_amount,
			p.hide_on_stock_overview, p.default_stock_label_type,
			p.should_not_be_frozen, p.treat_opened_as_out_of_stock,
			p.no_own_stock, p.default_consume_location_id, p.move_on_open,
			p.shopping_location_id, p.default_purchase_price_type,
			p.active, p.auto_reprint_stock_label, p.disable_open, p.last_price,
			IFNULL(sc.amount, 0) as stock_quantity,
			IFNULL(sc.best_before_date, '') as stock_best_before
		FROM products p
		LEFT JOIN locations l ON p.location_id = l.id
		LEFT JOIN stock_current sc ON p.id = sc.product_id
		WHERE p.id = ?`, id)
	return p, err
}

func getLocationsWithID() []models.Location {
	var locs []models.Location
	database.DB.Select(&locs, "SELECT id, name FROM locations ORDER BY name")
	return locs
}

func getProductGroups() []models.Location {
	var groups []models.Location
	database.DB.Select(&groups, "SELECT id, name FROM product_groups ORDER BY name")
	return groups
}

func buildProductFormData() models.ProductFormData {
	return models.ProductFormData{
		Locations:     getLocationsWithID(),
		ProductGroups: getProductGroups(),
		Fields:        GetProductFormFields(),
	}
}

// ProductsPage renders the product list (L in CRUDL).
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
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.ProductsContent(products))
	}
	return render(c, http.StatusOK, templates.ProductsPage(products))
}

// ProductNewForm renders the create form (C in CRUDL).
func ProductNewForm(c echo.Context) error {
	fd := buildProductFormData()
	return render(c, http.StatusOK, partials.ProductForm(nil, fd))
}

// ProductView renders the read-only detail view (R in CRUDL).
func ProductView(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	p, err := getProductDetailByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Product not found")
	}
	fd := buildProductFormData()
	return render(c, http.StatusOK, partials.ProductView(p, fd))
}

// ProductEditForm renders the edit form (U in CRUDL).
func ProductEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	p, err := getProductDetailByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Product not found")
	}
	fd := buildProductFormData()
	return render(c, http.StatusOK, partials.ProductForm(&p, fd))
}

// CreateProduct handles POST /products (C in CRUDL).
func CreateProduct(c echo.Context) error {
	name := c.FormValue("name")
	locationID, _ := strconv.Atoi(c.FormValue("location_id"))
	productGroupID := optionalInt(c.FormValue("product_group_id"))
	minStock, _ := strconv.Atoi(c.FormValue("min_stock_amount"))
	quantity, _ := strconv.Atoi(c.FormValue("stock_quantity"))
	expires := c.FormValue("stock_best_before")
	description := c.FormValue("description")
	barcode := c.FormValue("barcode")
	lastPrice := optionalFloat(c.FormValue("last_price"))
	calories, _ := strconv.Atoi(c.FormValue("calories"))
	defaultBBDays, _ := strconv.Atoi(c.FormValue("default_best_before_days"))
	defaultBBDaysOpen, _ := strconv.Atoi(c.FormValue("default_best_before_days_after_open"))
	defaultBBDaysFreezing, _ := strconv.Atoi(c.FormValue("default_best_before_days_after_freezing"))
	defaultBBDaysThawing, _ := strconv.Atoi(c.FormValue("default_best_before_days_after_thawing"))
	quickConsume := parseFloat(c.FormValue("quick_consume_amount"), 1)
	quickOpen := parseFloat(c.FormValue("quick_open_amount"), 1)
	tareWeight := parseFloat(c.FormValue("tare_weight"), 0)

	enableTare := checkboxVal(c, "enable_tare_weight_handling")
	shouldNotBeFrozen := checkboxVal(c, "should_not_be_frozen")
	treatOpenedOOS := checkboxVal(c, "treat_opened_as_out_of_stock")
	hideOnOverview := checkboxVal(c, "hide_on_stock_overview")
	noOwnStock := checkboxVal(c, "no_own_stock")
	moveOnOpen := checkboxVal(c, "move_on_open")
	disableOpen := checkboxVal(c, "disable_open")
	cumulateMin := checkboxVal(c, "cumulate_min_stock_amount_of_sub_products")
	notCheckRecipes := checkboxVal(c, "not_check_stock_fulfillment_for_recipes")
	autoReprint := checkboxVal(c, "auto_reprint_stock_label")

	var quID int
	database.DB.Get(&quID, "SELECT id FROM quantity_units LIMIT 1")

	result, err := database.DB.Exec(`
		INSERT INTO products (name, description, location_id, product_group_id, qu_id_purchase, qu_id_stock,
			qu_factor_purchase_to_stock, barcode, last_price, min_stock_amount,
			default_best_before_days, default_best_before_days_after_open,
			default_best_before_days_after_freezing, default_best_before_days_after_thawing,
			enable_tare_weight_handling, tare_weight, not_check_stock_fulfillment_for_recipes,
			calories, cumulate_min_stock_amount_of_sub_products,
			quick_consume_amount, quick_open_amount, hide_on_stock_overview,
			should_not_be_frozen, treat_opened_as_out_of_stock, no_own_stock,
			move_on_open, disable_open, auto_reprint_stock_label)
		VALUES (?, ?, ?, ?, ?, ?, 1.0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		name, description, locationID, productGroupID, quID, quID, barcode, lastPrice, minStock,
		defaultBBDays, defaultBBDaysOpen, defaultBBDaysFreezing, defaultBBDaysThawing,
		enableTare, tareWeight, notCheckRecipes, calories, cumulateMin,
		quickConsume, quickOpen, hideOnOverview,
		shouldNotBeFrozen, treatOpenedOOS, noOwnStock,
		moveOnOpen, disableOpen, autoReprint)
	if err != nil {
		return c.String(http.StatusBadRequest, "Product already exists or invalid data")
	}

	productID, _ := result.LastInsertId()

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

	// Handle picture upload.
	if pic, err := saveUploadedImage(c, "picture", "products"); err == nil && pic != "" {
		database.DB.Exec("UPDATE products SET picture_file_name = ? WHERE id = ?", pic, productID)
	}

	// Return the updated list
	c.Response().Header().Set("HX-Trigger", "productChanged")
	return ProductsPage(c)
}

// UpdateProduct handles PUT /products/:id (U in CRUDL).
func UpdateProduct(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	name := c.FormValue("name")
	locationID, _ := strconv.Atoi(c.FormValue("location_id"))
	productGroupID := optionalInt(c.FormValue("product_group_id"))
	minStock, _ := strconv.Atoi(c.FormValue("min_stock_amount"))
	description := c.FormValue("description")
	barcode := c.FormValue("barcode")
	lastPrice := optionalFloat(c.FormValue("last_price"))
	calories, _ := strconv.Atoi(c.FormValue("calories"))
	defaultBBDays, _ := strconv.Atoi(c.FormValue("default_best_before_days"))
	defaultBBDaysOpen, _ := strconv.Atoi(c.FormValue("default_best_before_days_after_open"))
	defaultBBDaysFreezing, _ := strconv.Atoi(c.FormValue("default_best_before_days_after_freezing"))
	defaultBBDaysThawing, _ := strconv.Atoi(c.FormValue("default_best_before_days_after_thawing"))
	quickConsume := parseFloat(c.FormValue("quick_consume_amount"), 1)
	quickOpen := parseFloat(c.FormValue("quick_open_amount"), 1)
	tareWeight := parseFloat(c.FormValue("tare_weight"), 0)

	enableTare := checkboxVal(c, "enable_tare_weight_handling")
	shouldNotBeFrozen := checkboxVal(c, "should_not_be_frozen")
	treatOpenedOOS := checkboxVal(c, "treat_opened_as_out_of_stock")
	hideOnOverview := checkboxVal(c, "hide_on_stock_overview")
	noOwnStock := checkboxVal(c, "no_own_stock")
	moveOnOpen := checkboxVal(c, "move_on_open")
	disableOpen := checkboxVal(c, "disable_open")
	cumulateMin := checkboxVal(c, "cumulate_min_stock_amount_of_sub_products")
	notCheckRecipes := checkboxVal(c, "not_check_stock_fulfillment_for_recipes")
	autoReprint := checkboxVal(c, "auto_reprint_stock_label")

	_, err := database.DB.Exec(`
		UPDATE products SET name=?, description=?, location_id=?, product_group_id=?, barcode=?, last_price=?, min_stock_amount=?,
			default_best_before_days=?, default_best_before_days_after_open=?,
			default_best_before_days_after_freezing=?, default_best_before_days_after_thawing=?,
			enable_tare_weight_handling=?, tare_weight=?, not_check_stock_fulfillment_for_recipes=?,
			calories=?, cumulate_min_stock_amount_of_sub_products=?,
			quick_consume_amount=?, quick_open_amount=?, hide_on_stock_overview=?,
			should_not_be_frozen=?, treat_opened_as_out_of_stock=?, no_own_stock=?,
			move_on_open=?, disable_open=?, auto_reprint_stock_label=?
		WHERE id=?`,
		name, description, locationID, productGroupID, barcode, lastPrice, minStock,
		defaultBBDays, defaultBBDaysOpen, defaultBBDaysFreezing, defaultBBDaysThawing,
		enableTare, tareWeight, notCheckRecipes, calories, cumulateMin,
		quickConsume, quickOpen, hideOnOverview,
		shouldNotBeFrozen, treatOpenedOOS, noOwnStock,
		moveOnOpen, disableOpen, autoReprint, id)
	if err != nil {
		return c.String(http.StatusBadRequest, "Update failed: "+err.Error())
	}

	// Handle picture upload.
	if pic, err := saveUploadedImage(c, "picture", "products"); err == nil && pic != "" {
		var oldPic string
		database.DB.Get(&oldPic, "SELECT IFNULL(picture_file_name,'') FROM products WHERE id = ?", id)
		deleteImage("products", oldPic)
		database.DB.Exec("UPDATE products SET picture_file_name = ? WHERE id = ?", pic, id)
	}

	c.Response().Header().Set("HX-Trigger", "productChanged")
	return ProductsPage(c)
}

// ConsumeProduct decrements quantity by 1.
func ConsumeProduct(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
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

// AddProduct increments quantity by 1.
func AddProduct(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
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

// AddToCart adds one of this product to the shopping list each time it's called.
func AddToCart(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	// If already on the list (not done), increment amount by 1.
	var existingID int
	err := database.DB.Get(&existingID, "SELECT id FROM shopping_list WHERE product_id=? AND done=0 LIMIT 1", id)
	if err == nil {
		database.DB.Exec("UPDATE shopping_list SET amount = amount + 1 WHERE id=?", existingID)
	} else {
		database.DB.Exec(`INSERT INTO shopping_list (shopping_list_id, product_id, amount) VALUES (1, ?, 1)`, id)
	}
	p, err := getProductByID(id)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}
	return render(c, http.StatusOK, partials.ProductRow(p))
}

// DeleteProduct removes a product (D in CRUDL).
func DeleteProduct(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	// Delete product image if exists.
	var picFile string
	database.DB.Get(&picFile, "SELECT IFNULL(picture_file_name,'') FROM products WHERE id = ?", id)
	if picFile != "" {
		deleteImage("products", picFile)
	}

	database.DB.Exec("DELETE FROM stock WHERE product_id = ?", id)
	database.DB.Exec("DELETE FROM stock_log WHERE product_id = ?", id)
	database.DB.Exec("DELETE FROM products WHERE id = ?", id)
	return c.NoContent(http.StatusOK)
}

// ExpiringSoonPartial returns the expiring-soon list for htmx.
func ExpiringSoonPartial(c echo.Context) error {
	var products []models.Product
	database.DB.Select(&products, `
		SELECT p.id, p.name, IFNULL(l.name, '') as location,
			IFNULL(sc.amount, 0) as quantity, p.min_stock_amount,
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

// ShoppingPreviewPartial returns the shopping list preview for htmx.
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

// Helper functions.
func parseFloat(s string, def float64) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}

func checkboxVal(c echo.Context, name string) int {
	if c.FormValue(name) == "1" || c.FormValue(name) == "on" {
		return 1
	}
	return 0
}

func optionalInt(s string) *int {
	if s == "" || s == "0" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}

func optionalFloat(s string) *float64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

// MergeProductForm shows the merge UI for a product.
func MergeProductForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var product models.Product
	err := database.DB.Get(&product, `
		SELECT p.id, p.name, IFNULL(l.name, '') as location,
			IFNULL(sc.amount, 0) as quantity, p.min_stock_amount,
			IFNULL(sc.best_before_date, '') as expires_at
		FROM products p
		LEFT JOIN locations l ON p.location_id = l.id
		LEFT JOIN stock_current sc ON p.id = sc.product_id
		WHERE p.id = ?`, id)
	if err != nil {
		return c.String(http.StatusNotFound, "Product not found")
	}

	// Get all other products as merge candidates.
	var candidates []models.Location
	database.DB.Select(&candidates, "SELECT id, name FROM products WHERE id != ? AND active = 1 ORDER BY name", id)

	return render(c, http.StatusOK, partials.MergeProductForm(product, candidates))
}

// MergeProduct merges one product into another.
func MergeProduct(c echo.Context) error {
	keepID, _ := strconv.Atoi(c.Param("id"))
	mergeID, _ := strconv.Atoi(c.FormValue("merge_id"))
	newName := strings.TrimSpace(c.FormValue("new_name"))

	// Get the keep product's current name.
	var keepName string
	database.DB.Get(&keepName, "SELECT name FROM products WHERE id = ?", keepID)
	if keepName == "" {
		return c.String(http.StatusBadRequest, "Product not found")
	}

	// If a new name is provided, rename the keep product.
	if newName != "" && newName != keepName {
		database.DB.Exec("UPDATE products SET name = ? WHERE id = ?", newName, keepID)
		// Also replace old name in recipes.
		replaceInRecipes(keepName, newName)
		keepName = newName
	}

	// If a merge target is selected, absorb it.
	if mergeID > 0 && mergeID != keepID {
		var mergeName string
		database.DB.Get(&mergeName, "SELECT name FROM products WHERE id = ?", mergeID)
		if mergeName != "" {
			// Replace in recipe ingredients (case-insensitive).
			replaceInRecipes(mergeName, keepName)

			// Transfer shopping list items.
			database.DB.Exec("UPDATE shopping_list SET product_id = ? WHERE product_id = ?", keepID, mergeID)

			// Transfer stock.
			database.DB.Exec("UPDATE stock SET product_id = ? WHERE product_id = ?", keepID, mergeID)
			database.DB.Exec("UPDATE stock_log SET product_id = ? WHERE product_id = ?", keepID, mergeID)

			// Delete the merged product.
			var picFile string
			database.DB.Get(&picFile, "SELECT IFNULL(picture_file_name,'') FROM products WHERE id = ?", mergeID)
			if picFile != "" {
				deleteImage("products", picFile)
			}
			database.DB.Exec("DELETE FROM products WHERE id = ?", mergeID)
		}
	}

	return ProductsPage(c)
}

// replaceInRecipes does a case-insensitive replacement of oldName with newName in all recipe ingredient text.
func replaceInRecipes(oldName, newName string) {
	// SQLite doesn't have case-insensitive REPLACE, so we fetch, replace in Go, and update.
	type ingRow struct {
		ID   int    `db:"id"`
		Text string `db:"text"`
	}
	var ingredients []ingRow
	database.DB.Select(&ingredients, "SELECT id, text FROM recipe_ingredients WHERE LOWER(text) LIKE ?",
		"%"+strings.ToLower(oldName)+"%")

	for _, ing := range ingredients {
		// Case-insensitive replace.
		re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(oldName))
		updated := re.ReplaceAllString(ing.Text, newName)
		if updated != ing.Text {
			database.DB.Exec("UPDATE recipe_ingredients SET text = ? WHERE id = ?", updated, ing.ID)
		}
	}
}
