package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"Domicile/templates/partials"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Tables exposed in the settings UI.
var settingsTables = []models.TableInfo{
	{Name: "products", DisplayName: "Products", Description: "Inventory items you track"},
	{Name: "stock", DisplayName: "Stock", Description: "Current stock entries"},
	{Name: "shopping_list", DisplayName: "Shopping List", Description: "Shopping list items"},
	{Name: "recipes", DisplayName: "Recipes", Description: "Your saved recipes"},
	{Name: "chores", DisplayName: "Chores", Description: "Household chores"},
	{Name: "tasks", DisplayName: "Tasks", Description: "To-do items"},
	{Name: "batteries", DisplayName: "Batteries", Description: "Battery tracking"},
	{Name: "equipment", DisplayName: "Equipment", Description: "Equipment & manuals"},
	{Name: "meal_plan", DisplayName: "Meal Plan", Description: "Meal planning entries"},
}

// formFields defines every field that appears on each entity's form.
// This is the single source of truth — not the DB schema.
// Order here = order in settings UI.
var formFields = map[string][]models.FieldInfo{
	"products": {
		{Name: "name", DisplayName: "Name", Description: "The product name, must be unique"},
		{Name: "active", DisplayName: "Active", Description: "Inactive products are hidden from most views"},
		{Name: "parent_product", DisplayName: "Parent Product", Description: "A parent product that aggregates this product's stock"},
		{Name: "description", DisplayName: "Description", Description: "Optional notes or details about this product"},
		{Name: "location", DisplayName: "Default Location", Description: "The default storage location for this product"},
		{Name: "default_consume_location", DisplayName: "Default Consume Location", Description: "Stock at this location will be consumed first"},
		{Name: "move_on_open", DisplayName: "Move On Open", Description: "When marked as opened, the amount is automatically moved to the default consume location"},
		{Name: "default_store", DisplayName: "Default Store", Description: "The default store where this product is purchased"},
		{Name: "min_stock_amount", DisplayName: "Minimum Stock Amount", Description: "When current stock falls below this amount, the product is considered missing"},
		{Name: "cumulate_min_stock_amount_of_sub_products", DisplayName: "Accumulate Sub-Product Min Stock", Description: "Sub-product minimum stock amounts are rolled up into this product"},
		{Name: "treat_opened_as_out_of_stock", DisplayName: "Treat Opened As Out Of Stock", Description: "Opened items count as missing when checking minimum stock"},
		{Name: "due_type", DisplayName: "Due Date Type", Description: "Best-before means possibly safe after date. Expiration means not safe after date"},
		{Name: "default_best_before_days", DisplayName: "Default Due Days", Description: "On purchase, this many days are added to today for the due date suggestion. -1 means never overdue"},
		{Name: "default_best_before_days_after_open", DisplayName: "Default Due Days After Opened", Description: "When opened, the due date is replaced by today plus this many days (0 disables)"},
		{Name: "default_best_before_days_after_freezing", DisplayName: "Default Due Days After Freezing", Description: "When moved to a freezer, the due date is replaced by today plus this many days"},
		{Name: "default_best_before_days_after_thawing", DisplayName: "Default Due Days After Thawing", Description: "When moved out of a freezer, the due date is replaced by today plus this many days"},
		{Name: "should_not_be_frozen", DisplayName: "Should Not Be Frozen", Description: "A warning is shown when moving this product to a freezer location"},
		{Name: "product_group", DisplayName: "Product Group", Description: "Category or group this product belongs to"},
		{Name: "qu_stock", DisplayName: "Quantity Unit Stock", Description: "The unit used for tracking this product in stock"},
		{Name: "qu_purchase", DisplayName: "Default QU Purchase", Description: "The default unit used on purchase and shopping list"},
		{Name: "qu_consume", DisplayName: "Default QU Consume", Description: "The default unit used when consuming this product"},
		{Name: "qu_price", DisplayName: "QU For Prices", Description: "Prices are displayed relative to this unit"},
		{Name: "enable_tare_weight_handling", DisplayName: "Tare Weight Handling", Description: "Weigh the whole container and the net amount is calculated from the tare weight"},
		{Name: "tare_weight", DisplayName: "Tare Weight", Description: "The empty container weight, used when tare weight handling is enabled"},
		{Name: "not_check_stock_fulfillment_for_recipes", DisplayName: "Disable Stock Fulfillment Check", Description: "When used as a recipe ingredient, stock availability will not be checked"},
		{Name: "calories", DisplayName: "Energy (kcal)", Description: "Energy content per stock quantity unit"},
		{Name: "quick_consume_amount", DisplayName: "Quick Consume Amount", Description: "The amount used by the quick consume button on the stock overview"},
		{Name: "quick_open_amount", DisplayName: "Quick Open Amount", Description: "The amount used by the quick open button on the stock overview"},
		{Name: "default_stock_label_type", DisplayName: "Default Stock Label", Description: "The default label type on purchase: no label, single label, or label per unit"},
		{Name: "auto_reprint_stock_label", DisplayName: "Auto Reprint Stock Label", Description: "Reprint label when due date changes automatically"},
		{Name: "default_purchase_price_type", DisplayName: "Default Purchase Price Type", Description: "Default price type on purchase: unspecified, unit price, or total price"},
		{Name: "disable_open", DisplayName: "Can't Be Opened", Description: "This product does not support open/unopened tracking"},
		{Name: "hide_on_stock_overview", DisplayName: "Never Show On Stock Overview", Description: "This product never appears on the stock overview page"},
		{Name: "no_own_stock", DisplayName: "Disable Own Stock", Description: "This product can't have its own stock — useful for parent products"},
		{Name: "barcode", DisplayName: "Barcode", Description: "A barcode associated with this product"},
		{Name: "last_price", DisplayName: "Last Price", Description: "The last known price for this product"},
		{Name: "picture", DisplayName: "Picture", Description: "Product image"},
		{Name: "qu_factor_purchase_to_stock", DisplayName: "Purchase To Stock Factor", Description: "Conversion factor from purchase unit to stock unit"},
	},
	"stock": {
		{Name: "amount", DisplayName: "Amount", Description: "Quantity currently in stock"},
		{Name: "best_before_date", DisplayName: "Best Before Date", Description: "The due date for this stock entry"},
		{Name: "purchased_date", DisplayName: "Purchased Date", Description: "When this stock entry was purchased"},
		{Name: "price", DisplayName: "Price", Description: "The price paid for this stock entry"},
		{Name: "open", DisplayName: "Open", Description: "Whether this stock entry has been opened"},
		{Name: "opened_date", DisplayName: "Opened Date", Description: "When this stock entry was opened"},
		{Name: "location", DisplayName: "Location", Description: "Where this stock entry is stored"},
		{Name: "store", DisplayName: "Store", Description: "Where this stock was purchased"},
	},
	"shopping_list": {
		{Name: "product", DisplayName: "Product", Description: "The product to buy"},
		{Name: "amount", DisplayName: "Amount", Description: "How many to buy"},
		{Name: "note", DisplayName: "Note", Description: "A freeform note for this item"},
		{Name: "done", DisplayName: "Done", Description: "Whether this item has been purchased"},
	},
	"recipes": {
		{Name: "name", DisplayName: "Name", Description: "The recipe name"},
		{Name: "description", DisplayName: "Description", Description: "Instructions and details"},
		{Name: "picture", DisplayName: "Picture", Description: "Recipe image"},
		{Name: "base_servings", DisplayName: "Base Servings", Description: "The number of servings the ingredient amounts are based on"},
		{Name: "desired_servings", DisplayName: "Desired Servings", Description: "How many servings you want — ingredients scale accordingly"},
		{Name: "not_check_shoppinglist", DisplayName: "Don't Add To Shopping List", Description: "Missing ingredients won't be auto-added to the shopping list"},
	},
	"chores": {
		{Name: "name", DisplayName: "Name", Description: "The chore name"},
		{Name: "description", DisplayName: "Description", Description: "Details or instructions"},
		{Name: "period_type", DisplayName: "Period Type", Description: "How the chore repeats: manually, daily, weekly, monthly, etc."},
		{Name: "period_interval", DisplayName: "Period Interval", Description: "Multiplier for the period (e.g. every 2 weeks)"},
		{Name: "track_date_only", DisplayName: "Track Date Only", Description: "Only track the date, not the time"},
		{Name: "rollover", DisplayName: "Rollover", Description: "Missed occurrences carry over to the next due date"},
		{Name: "assignment_type", DisplayName: "Assignment Type", Description: "How the chore is assigned to users"},
		{Name: "consume_product_on_execution", DisplayName: "Consume Product On Execution", Description: "Automatically consume a linked product when completed"},
		{Name: "start_date", DisplayName: "Start Date", Description: "When this chore schedule begins"},
		{Name: "active", DisplayName: "Active", Description: "Inactive chores are hidden from most views"},
	},
	"tasks": {
		{Name: "name", DisplayName: "Name", Description: "The task name"},
		{Name: "description", DisplayName: "Description", Description: "Details about this task"},
		{Name: "due_date", DisplayName: "Due Date", Description: "When this task should be completed"},
		{Name: "category", DisplayName: "Category", Description: "Task category"},
		{Name: "assigned_to", DisplayName: "Assigned To", Description: "User this task is assigned to"},
		{Name: "done", DisplayName: "Done", Description: "Whether this task is complete"},
	},
	"batteries": {
		{Name: "name", DisplayName: "Name", Description: "The battery name"},
		{Name: "description", DisplayName: "Description", Description: "Details about this battery"},
		{Name: "used_in", DisplayName: "Used In", Description: "What device this battery is used in"},
		{Name: "charge_interval_days", DisplayName: "Charge Interval (Days)", Description: "How often this battery should be charged"},
		{Name: "active", DisplayName: "Active", Description: "Inactive batteries are hidden from most views"},
	},
	"equipment": {
		{Name: "name", DisplayName: "Name", Description: "The equipment name"},
		{Name: "description", DisplayName: "Description", Description: "Details about this equipment"},
		{Name: "instruction_manual", DisplayName: "Instruction Manual", Description: "Uploaded instruction manual file"},
	},
	"meal_plan": {
		{Name: "day", DisplayName: "Day", Description: "The date for this meal plan entry"},
		{Name: "type", DisplayName: "Type", Description: "Entry type: recipe, product, or note"},
		{Name: "recipe", DisplayName: "Recipe", Description: "The recipe for this meal"},
		{Name: "product", DisplayName: "Product", Description: "The product for this meal"},
		{Name: "product_amount", DisplayName: "Product Amount", Description: "Amount of the product"},
		{Name: "note", DisplayName: "Note", Description: "Optional note"},
		{Name: "done", DisplayName: "Done", Description: "Whether this meal has been prepared"},
	},
}

// getFieldsForTable returns the form fields for a table, merged with visibility settings from the DB.
func getFieldsForTable(tableName string) []models.FieldInfo {
	defs, ok := formFields[tableName]
	if !ok {
		return nil
	}

	// Load visibility overrides from DB.
	var vis []models.FieldVisibility
	database.DB.Select(&vis, "SELECT id, table_name, field_name, hidden FROM field_visibility WHERE table_name = ?", tableName)
	hiddenMap := make(map[string]bool)
	for _, v := range vis {
		hiddenMap[v.FieldName] = v.Hidden == 1
	}

	// Merge: start from the defined fields, apply DB overrides.
	result := make([]models.FieldInfo, len(defs))
	copy(result, defs)
	for i := range result {
		result[i].Hidden = hiddenMap[result[i].Name]
	}
	return result
}

// SettingsPage renders the top-level settings menu.
func SettingsPage(c echo.Context) error {
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.SettingsContent())
	}
	return render(c, http.StatusOK, templates.SettingsPage())
}

// SettingsFieldVisibility renders the field visibility sub-page (table list).
func SettingsFieldVisibility(c echo.Context) error {
	return render(c, http.StatusOK, partials.SettingsFieldVisibility(settingsTables))
}

// SettingsFieldList returns the field toggles for a specific table.
func SettingsFieldList(c echo.Context) error {
	tableName := c.QueryParam("table")
	if tableName == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	displayName := tableName
	for _, t := range settingsTables {
		if t.Name == tableName {
			displayName = t.DisplayName
			break
		}
	}

	fields := getFieldsForTable(tableName)
	return render(c, http.StatusOK, partials.FieldList(tableName, displayName, fields))
}

// ToggleFieldVisibility sets a field's hidden state. Expects hidden=0 or hidden=1.
func ToggleFieldVisibility(c echo.Context) error {
	tableName := c.FormValue("table_name")
	fieldName := c.FormValue("field_name")
	hiddenStr := c.FormValue("hidden")
	if tableName == "" || fieldName == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	hidden := 1
	if hiddenStr == "0" {
		hidden = 0
	}

	database.DB.Exec(`
		INSERT INTO field_visibility (table_name, field_name, hidden)
		VALUES (?, ?, ?)
		ON CONFLICT(table_name, field_name)
		DO UPDATE SET hidden = ?`,
		tableName, fieldName, hidden, hidden)

	return c.NoContent(http.StatusNoContent)
}

// GetProductFormFields returns the visible product fields with display metadata.
func GetProductFormFields() []models.ProductFormField {
	fields := getFieldsForTable("products")
	var result []models.ProductFormField
	for _, f := range fields {
		if f.Hidden {
			continue
		}
		result = append(result, models.ProductFormField{
			Name:        f.Name,
			DisplayName: f.DisplayName,
			Description: f.Description,
		})
	}
	return result
}

// IsFieldVisible checks if a specific field is visible for a table.
func IsFieldVisible(tableName, fieldName string) bool {
	fields := getFieldsForTable(tableName)
	for _, f := range fields {
		if f.Name == fieldName {
			return !f.Hidden
		}
	}
	return true
}

// ShoppingSettingsPage renders the shopping settings panel.
func ShoppingSettingsPage(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	autoStock := GetUserSetting(userID, "shopping_auto_stock", "1") == "1"
	return render(c, http.StatusOK, partials.ShoppingSettings(autoStock))
}

// SaveUserSetting handles POST /settings/user-setting.
func SaveUserSetting(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	key := c.FormValue("key")
	value := c.FormValue("value")
	if key == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	SetUserSetting(userID, key, value)
	return c.NoContent(http.StatusNoContent)
}
