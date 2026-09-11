package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"Domicile/templates/partials"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const shoppingStoreKey = "shopping_store_id"

func getStores() []models.ShoppingLocation {
	var stores []models.ShoppingLocation
	database.DB.Select(&stores, "SELECT id, name FROM shopping_locations ORDER BY name")
	return stores
}

// activeStoreID returns the store filter for the current user: query param > saved setting > 0 (all).
func activeStoreID(c echo.Context) int {
	if s := c.QueryParam("store_id"); s != "" {
		id, _ := strconv.Atoi(s)
		return id
	}
	userID, _ := c.Get("user_id").(int)
	saved := GetUserSetting(userID, shoppingStoreKey, "0")
	id, _ := strconv.Atoi(saved)
	return id
}

func getShoppingItems(storeID int) []models.ShoppingListItem {
	var items []models.ShoppingListItem
	query := `
		SELECT sl.id, sl.product_id, IFNULL(p.name,'') as product_name,
			IFNULL(pg.name,'Ungrouped') as product_group,
			pg.id as product_group_id,
			pg.store_id as store_id,
			IFNULL(l.name,'') as home_loc,
			p.last_price,
			IFNULL(sl.note,'') as note, sl.amount, sl.amount_autoadded, sl.done
		FROM shopping_list sl
		LEFT JOIN products p ON sl.product_id = p.id
		LEFT JOIN product_groups pg ON p.product_group_id = pg.id
		LEFT JOIN locations l ON p.location_id = l.id`
	if storeID > 0 {
		query += ` WHERE (pg.store_id = ` + strconv.Itoa(storeID) + ` OR pg.id IS NULL)`
	}
	query += ` ORDER BY sl.done ASC, pg.name ASC, p.name ASC, sl.note ASC`
	database.DB.Select(&items, query)
	return items
}

func getShoppingGrouped(storeID int) []models.ShoppingGroup {
	items := getShoppingItems(storeID)
	groupMap := make(map[string][]models.ShoppingListItem)
	var groupOrder []string
	for _, item := range items {
		g := item.ProductGroup
		if _, exists := groupMap[g]; !exists {
			groupOrder = append(groupOrder, g)
		}
		groupMap[g] = append(groupMap[g], item)
	}
	var groups []models.ShoppingGroup
	for _, name := range groupOrder {
		groups = append(groups, models.ShoppingGroup{GroupName: name, Items: groupMap[name]})
	}
	return groups
}

func getProductsForDropdown() []models.Location {
	var products []models.Location
	database.DB.Select(&products, "SELECT id, name FROM products WHERE active = 1 ORDER BY name")
	return products
}

func getFrequentProducts() []models.Location {
	var products []models.Location
	database.DB.Select(&products, "SELECT id, name FROM products WHERE active = 1 AND purchase_count > 0 ORDER BY purchase_count DESC, name")
	return products
}

// ShoppingPage renders the shopping list.
func ShoppingPage(c echo.Context) error {
	storeID := activeStoreID(c)
	// Persist store choice.
	if s := c.QueryParam("store_id"); s != "" {
		userID, _ := c.Get("user_id").(int)
		SaveUserSettingDirect(userID, shoppingStoreKey, s)
	}
	items := getShoppingItems(storeID)
	stores := getStores()
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.ShoppingContent(items, stores, storeID))
	}
	return render(c, http.StatusOK, templates.ShoppingPage(items, stores, storeID))
}

// ShoppingNewForm renders the add item form.
func ShoppingNewForm(c echo.Context) error {
	products := getProductsForDropdown()
	frequentProducts := getFrequentProducts()
	storeID := activeStoreID(c)
	currentItems := getShoppingItems(storeID)
	return render(c, http.StatusOK, partials.ShoppingItemForm(nil, products, frequentProducts, currentItems))
}

// ShoppingEditForm renders the edit item form.
func ShoppingEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.ShoppingListItem
	err := database.DB.Get(&item, `
		SELECT sl.id, sl.product_id, IFNULL(p.name,'') as product_name,
			IFNULL(pg.name,'Ungrouped') as product_group,
			pg.id as product_group_id, pg.store_id as store_id,
			IFNULL(l.name,'') as home_loc,
			p.last_price,
			IFNULL(sl.note,'') as note, sl.amount, sl.amount_autoadded, sl.done
		FROM shopping_list sl
		LEFT JOIN products p ON sl.product_id = p.id
		LEFT JOIN product_groups pg ON p.product_group_id = pg.id
		LEFT JOIN locations l ON p.location_id = l.id
		WHERE sl.id = ?`, id)
	if err != nil {
		return c.String(http.StatusNotFound, "Item not found")
	}
	products := getProductsForDropdown()
	return render(c, http.StatusOK, partials.ShoppingItemForm(&item, products, nil, nil))
}

// CreateShoppingItem handles POST /shopping.
func CreateShoppingItem(c echo.Context) error {
	productIDStr := c.FormValue("product_id")
	note := c.FormValue("note")
	amount, _ := strconv.ParseFloat(c.FormValue("amount"), 64)
	if amount <= 0 {
		amount = 1
	}

	var productID *int
	if productIDStr != "" && productIDStr != "0" {
		pid, _ := strconv.Atoi(productIDStr)
		productID = &pid
	}

	if productID == nil && strings.TrimSpace(note) == "" {
		return c.String(http.StatusBadRequest, `<div class="alert alert-danger">Did not provide an item — select a product or enter a note.</div>`)
	}

	database.DB.Exec(`
		INSERT INTO shopping_list (shopping_list_id, product_id, note, amount)
		VALUES (1, ?, NULLIF(?, ''), ?)`,
		productID, note, amount)

	return ShoppingPage(c)
}

// CreateShoppingItemContinue handles POST /shopping/add-continue.
// It adds the item then re-renders the add form so the user can keep adding.
func CreateShoppingItemContinue(c echo.Context) error {
	productIDStr := c.FormValue("product_id")
	note := c.FormValue("note")
	amount, _ := strconv.ParseFloat(c.FormValue("amount"), 64)
	if amount <= 0 {
		amount = 1
	}

	var productID *int
	if productIDStr != "" && productIDStr != "0" {
		pid, _ := strconv.Atoi(productIDStr)
		productID = &pid
	}

	if productID == nil && strings.TrimSpace(note) == "" {
		return c.String(http.StatusBadRequest, `<div class="alert alert-danger">Did not provide an item — select a product or enter a note.</div>`)
	}

	database.DB.Exec(`
		INSERT INTO shopping_list (shopping_list_id, product_id, note, amount)
		VALUES (1, ?, NULLIF(?, ''), ?)`,
		productID, note, amount)

	// Re-render the add form with refreshed current items list
	products := getProductsForDropdown()
	frequentProducts := getFrequentProducts()
	storeID := activeStoreID(c)
	currentItems := getShoppingItems(storeID)
	return render(c, http.StatusOK, partials.ShoppingItemForm(nil, products, frequentProducts, currentItems))
}

// UpdateShoppingItem handles PUT /shopping/:id.
func UpdateShoppingItem(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	productIDStr := c.FormValue("product_id")
	note := c.FormValue("note")
	amount, _ := strconv.ParseFloat(c.FormValue("amount"), 64)
	if amount <= 0 {
		amount = 1
	}

	var productID *int
	if productIDStr != "" && productIDStr != "0" {
		pid, _ := strconv.Atoi(productIDStr)
		productID = &pid
	}

	database.DB.Exec("UPDATE shopping_list SET product_id=?, note=?, amount=? WHERE id=?",
		productID, note, amount, id)

	return ShoppingPage(c)
}

// ToggleShoppingDone toggles the done state of an item.
func ToggleShoppingDone(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	userID, _ := c.Get("user_id").(int)

	var wasDone int
	database.DB.Get(&wasDone, "SELECT done FROM shopping_list WHERE id = ?", id)
	database.DB.Exec("UPDATE shopping_list SET done = CASE WHEN done = 1 THEN 0 ELSE 1 END WHERE id = ?", id)

	if wasDone == 0 {
		autoStock := GetUserSetting(userID, "shopping_auto_stock", "1")
		if autoStock == "1" {
			autoAddToStock(id)
		}
	}

	var item models.ShoppingListItem
	database.DB.Get(&item, `
		SELECT sl.id, sl.product_id, IFNULL(p.name,'') as product_name,
			IFNULL(pg.name,'Ungrouped') as product_group,
			pg.id as product_group_id, pg.store_id as store_id,
			IFNULL(l.name,'') as home_loc,
			p.last_price,
			IFNULL(sl.note,'') as note, sl.amount, sl.amount_autoadded, sl.done
		FROM shopping_list sl
		LEFT JOIN products p ON sl.product_id = p.id
		LEFT JOIN product_groups pg ON p.product_group_id = pg.id
		LEFT JOIN locations l ON p.location_id = l.id
		WHERE sl.id = ?`, id)
	return render(c, http.StatusOK, partials.ShoppingRow(item))
}

func autoAddToStock(shoppingItemID int) {
	var item struct {
		ProductID  *int    `db:"product_id"`
		Amount     float64 `db:"amount"`
		LocationID int     `db:"location_id"`
	}
	err := database.DB.Get(&item, `
		SELECT sl.product_id, sl.amount, IFNULL(p.location_id, 1) as location_id
		FROM shopping_list sl
		LEFT JOIN products p ON sl.product_id = p.id
		WHERE sl.id = ?`, shoppingItemID)
	if err != nil || item.ProductID == nil {
		return
	}

	stockID := uuid.New().String()
	database.DB.Exec(`
		INSERT INTO stock (product_id, amount, stock_id, location_id)
		VALUES (?, ?, ?, ?)`,
		*item.ProductID, item.Amount, stockID, item.LocationID)
	database.DB.Exec(`
		INSERT INTO stock_log (product_id, amount, stock_id, transaction_type, location_id)
		VALUES (?, ?, ?, 'purchase', ?)`,
		*item.ProductID, item.Amount, stockID, item.LocationID)
}

// DeleteShoppingItem handles DELETE /shopping/:id.
func DeleteShoppingItem(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM shopping_list WHERE id = ?", id)
	return c.NoContent(http.StatusOK)
}

// ClearDoneItems removes all done items.
func ClearDoneItems(c echo.Context) error {
	// Increment purchase_count for products that were on the done list
	database.DB.Exec(`
		UPDATE products SET purchase_count = purchase_count + 1
		WHERE id IN (SELECT product_id FROM shopping_list WHERE done = 1 AND product_id IS NOT NULL)`)
	database.DB.Exec("DELETE FROM shopping_list WHERE done = 1")
	return ShoppingPage(c)
}

// ClearList removes every item from the shopping list.
func ClearList(c echo.Context) error {
	// Increment purchase_count for all products on the list
	database.DB.Exec(`
		UPDATE products SET purchase_count = purchase_count + 1
		WHERE id IN (SELECT product_id FROM shopping_list WHERE product_id IS NOT NULL)`)
	database.DB.Exec("DELETE FROM shopping_list")
	return ShoppingPage(c)
}

// AddDoneToStock pushes all done items with a product_id into stock, then removes them.
func AddDoneToStock(c echo.Context) error {
	// Increment purchase_count for products being added to stock
	database.DB.Exec(`
		UPDATE products SET purchase_count = purchase_count + 1
		WHERE id IN (SELECT product_id FROM shopping_list WHERE done = 1 AND product_id IS NOT NULL)`)
	var doneIDs []int
	database.DB.Select(&doneIDs, "SELECT id FROM shopping_list WHERE done = 1 AND product_id IS NOT NULL")
	for _, id := range doneIDs {
		autoAddToStock(id)
	}
	database.DB.Exec("DELETE FROM shopping_list WHERE done = 1")
	return ShoppingPage(c)
}

// AddMissingStock adds products below min stock to the shopping list.
func AddMissingStock(c echo.Context) error {
	type missing struct {
		ID            int     `db:"id"`
		AmountMissing float64 `db:"amount_missing"`
	}
	var items []missing
	database.DB.Select(&items, `
		SELECT smp.id, smp.amount_missing
		FROM stock_missing_products smp
		WHERE smp.id NOT IN (
			SELECT COALESCE(product_id, 0) FROM shopping_list WHERE done = 0
		)`)

	for _, m := range items {
		database.DB.Exec(`
			INSERT INTO shopping_list (shopping_list_id, product_id, amount, amount_autoadded)
			VALUES (1, ?, ?, ?)`,
			m.ID, m.AmountMissing, m.AmountMissing)
	}

	return ShoppingPage(c)
}

// ShoppingShopMode renders the phone-friendly shopping mode grouped by aisle.
func ShoppingShopMode(c echo.Context) error {
	storeID := activeStoreID(c)
	groups := getShoppingGrouped(storeID)
	stores := getStores()
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.ShopModeContent(groups, stores, storeID))
	}
	return render(c, http.StatusOK, templates.ShopModePage(groups, stores, storeID))
}

// ShoppingPrint renders the printable shopping list grouped by aisle.
func ShoppingPrint(c echo.Context) error {
	storeID := activeStoreID(c)
	groups := getShoppingGrouped(storeID)
	storeName := ""
	if storeID > 0 {
		database.DB.Get(&storeName, "SELECT name FROM shopping_locations WHERE id = ?", storeID)
	}
	return render(c, http.StatusOK, templates.ShoppingPrintPage(groups, storeName))
}

// ShopModeToggle toggles done and returns the updated shop mode row.
func ShopModeToggle(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	userID, _ := c.Get("user_id").(int)

	var wasDone int
	database.DB.Get(&wasDone, "SELECT done FROM shopping_list WHERE id = ?", id)
	database.DB.Exec("UPDATE shopping_list SET done = CASE WHEN done = 1 THEN 0 ELSE 1 END WHERE id = ?", id)

	if wasDone == 0 {
		autoStock := GetUserSetting(userID, "shopping_auto_stock", "1")
		if autoStock == "1" {
			autoAddToStock(id)
		}
	}

	var item models.ShoppingListItem
	database.DB.Get(&item, `
		SELECT sl.id, sl.product_id, IFNULL(p.name,'') as product_name,
			IFNULL(pg.name,'Ungrouped') as product_group,
			pg.id as product_group_id, pg.store_id as store_id,
			IFNULL(l.name,'') as home_loc,
			p.last_price,
			IFNULL(sl.note,'') as note, sl.amount, sl.amount_autoadded, sl.done
		FROM shopping_list sl
		LEFT JOIN products p ON sl.product_id = p.id
		LEFT JOIN product_groups pg ON p.product_group_id = pg.id
		LEFT JOIN locations l ON p.location_id = l.id
		WHERE sl.id = ?`, id)
	return render(c, http.StatusOK, partials.ShopModeItem(item))
}

// AssignProductGroup assigns a product group to a product from the shopping list.
func AssignProductGroup(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	groupID, _ := strconv.Atoi(c.FormValue("product_group_id"))
	if groupID == 0 {
		return c.NoContent(http.StatusBadRequest)
	}

	var productID *int
	database.DB.Get(&productID, "SELECT product_id FROM shopping_list WHERE id = ?", id)
	if productID == nil {
		return c.NoContent(http.StatusBadRequest)
	}

	database.DB.Exec("UPDATE products SET product_group_id = ? WHERE id = ?", groupID, *productID)
	return c.NoContent(http.StatusOK)
}

// ProductGroupsJSON returns product groups filtered by store if provided.
func ProductGroupsJSON(c echo.Context) error {
	storeID, _ := strconv.Atoi(c.QueryParam("store_id"))
	var groups []models.Location
	if storeID > 0 {
		database.DB.Select(&groups, "SELECT id, name FROM product_groups WHERE store_id = ? ORDER BY name", storeID)
	} else {
		database.DB.Select(&groups, "SELECT id, name FROM product_groups ORDER BY name")
	}
	return c.JSON(http.StatusOK, groups)
}

// SaveUserSettingDirect saves a user setting without an HTTP context (internal use).
func SaveUserSettingDirect(userID int, key, value string) {
	database.DB.Exec(`
		INSERT INTO user_settings (user_id, key, value) VALUES (?, ?, ?)
		ON CONFLICT(user_id, key) DO UPDATE SET value=excluded.value, row_updated_timestamp=datetime('now','localtime')`,
		userID, key, value)
}
