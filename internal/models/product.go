package models

import (
	"strconv"
	"strings"
	"time"
)

// Product is the slim list-view model.
type Product struct {
	ID             int       `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Location       string    `json:"location" db:"location"`
	Quantity       int       `json:"quantity" db:"quantity"`
	MinStockAmount int       `json:"min_stock_amount" db:"min_stock_amount"`
	ExpiresAt      time.Time `json:"expires_at,omitempty" db:"-"`
	ExpiresAtStr   string    `json:"-" db:"expires_at"`
}

func (p Product) IDStr() string       { return strconv.Itoa(p.ID) }
func (p Product) QuantityStr() string { return strconv.Itoa(p.Quantity) }

func (p Product) ExpiresDisplay() string {
	if p.ExpiresAtStr != "" {
		t, err := time.Parse("2006-01-02", p.ExpiresAtStr)
		if err == nil {
			return t.Format("Jan 2, 2006")
		}
	}
	if !p.ExpiresAt.IsZero() {
		return p.ExpiresAt.Format("Jan 2, 2006")
	}
	return "—"
}

func (p Product) IsExpiringSoon() bool {
	var t time.Time
	if p.ExpiresAtStr != "" {
		t, _ = time.Parse("2006-01-02", p.ExpiresAtStr)
	} else {
		t = p.ExpiresAt
	}
	if t.IsZero() {
		return false
	}
	return time.Until(t) < 7*24*time.Hour && time.Until(t) > 0
}

// ProductDetail is the full model for create/edit/view forms.
type ProductDetail struct {
	ID                                  int      `db:"id"`
	Name                                string   `db:"name"`
	Description                         string   `db:"description"`
	LocationID                          int      `db:"location_id"`
	LocationName                        string   `db:"location_name"`
	ProductGroupID                      *int     `db:"product_group_id"`
	QuIDPurchase                        int      `db:"qu_id_purchase"`
	QuIDStock                           int      `db:"qu_id_stock"`
	QuIDConsume                         *int     `db:"qu_id_consume"`
	QuIDPrice                           *int     `db:"qu_id_price"`
	QuFactorPurchaseToStock             float64  `db:"qu_factor_purchase_to_stock"`
	Barcode                             string   `db:"barcode"`
	MinStockAmount                      int      `db:"min_stock_amount"`
	DefaultBestBeforeDays               int      `db:"default_best_before_days"`
	DefaultBestBeforeDaysAfterOpen      int      `db:"default_best_before_days_after_open"`
	DefaultBestBeforeDaysAfterFreezing  int      `db:"default_best_before_days_after_freezing"`
	DefaultBestBeforeDaysAfterThawing   int      `db:"default_best_before_days_after_thawing"`
	PictureFileName                     string   `db:"picture_file_name"`
	EnableTareWeightHandling            int      `db:"enable_tare_weight_handling"`
	TareWeight                          float64  `db:"tare_weight"`
	NotCheckStockFulfillmentForRecipes  int      `db:"not_check_stock_fulfillment_for_recipes"`
	ParentProductID                     *int     `db:"parent_product_id"`
	Calories                            *int     `db:"calories"`
	CumulateMinStockAmountOfSubProducts int      `db:"cumulate_min_stock_amount_of_sub_products"`
	DueType                             int      `db:"due_type"`
	QuickConsumeAmount                  float64  `db:"quick_consume_amount"`
	QuickOpenAmount                     float64  `db:"quick_open_amount"`
	HideOnStockOverview                 int      `db:"hide_on_stock_overview"`
	DefaultStockLabelType               int      `db:"default_stock_label_type"`
	ShouldNotBeFrozen                   int      `db:"should_not_be_frozen"`
	TreatOpenedAsOutOfStock             int      `db:"treat_opened_as_out_of_stock"`
	NoOwnStock                          int      `db:"no_own_stock"`
	DefaultConsumeLocationID            *int     `db:"default_consume_location_id"`
	MoveOnOpen                          int      `db:"move_on_open"`
	ShoppingLocationID                  *int     `db:"shopping_location_id"`
	DefaultPurchasePriceType            *int     `db:"default_purchase_price_type"`
	Active                              int      `db:"active"`
	AutoReprintStockLabel               int      `db:"auto_reprint_stock_label"`
	DisableOpen                         int      `db:"disable_open"`
	LastPrice                           *float64 `db:"last_price"`
	PurchaseCount                       int      `db:"purchase_count"`
	// Stock info (joined, not in products table)
	StockQuantity   int    `db:"stock_quantity"`
	StockBestBefore string `db:"stock_best_before"`
}

func (p ProductDetail) IDStr() string { return strconv.Itoa(p.ID) }

func (p ProductDetail) CaloriesStr() string {
	if p.Calories != nil {
		return strconv.Itoa(*p.Calories)
	}
	return "0"
}

func (p ProductDetail) LastPriceStr() string {
	if p.LastPrice != nil {
		return strconv.FormatFloat(*p.LastPrice, 'f', 2, 64)
	}
	return ""
}

func (p ProductDetail) PurchaseCountStr() string {
	return strconv.Itoa(p.PurchaseCount)
}

// Location is a simple lookup for dropdowns.
type Location struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type ShoppingItem struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Quantity int    `json:"quantity" db:"quantity"`
	Done     bool   `json:"done" db:"done"`
}

func (s ShoppingItem) IDStr() string       { return strconv.Itoa(s.ID) }
func (s ShoppingItem) QuantityStr() string { return strconv.Itoa(s.Quantity) }

type DashboardStats struct {
	ProductCount  int
	ExpiringSoon  int
	BelowMinStock int
	ShoppingItems int
}

func (d DashboardStats) ProductCountStr() string  { return strconv.Itoa(d.ProductCount) }
func (d DashboardStats) ExpiringSoonStr() string  { return strconv.Itoa(d.ExpiringSoon) }
func (d DashboardStats) BelowMinStockStr() string { return strconv.Itoa(d.BelowMinStock) }
func (d DashboardStats) ShoppingItemsStr() string { return strconv.Itoa(d.ShoppingItems) }

// LocationDetail is the full model for location CRUD.
type LocationDetail struct {
	ID              int    `db:"id"`
	Name            string `db:"name"`
	Description     string `db:"description"`
	IsFreezer       int    `db:"is_freezer"`
	PictureFileName string `db:"picture_file_name"`
}

func (l LocationDetail) IDStr() string { return strconv.Itoa(l.ID) }

// ShoppingListItem is the full model for shopping list CRUD.
type ShoppingListItem struct {
	ID              int      `db:"id"`
	ProductID       *int     `db:"product_id"`
	ProductName     string   `db:"product_name"`
	ProductGroup    string   `db:"product_group"`
	ProductGroupID  *int     `db:"product_group_id"`
	StoreID         *int     `db:"store_id"`
	HomeLoc         string   `db:"home_loc"`
	LastPrice       *float64 `db:"last_price"`
	Note            string   `db:"note"`
	Amount          float64  `db:"amount"`
	AmountAutoadded float64  `db:"amount_autoadded"`
	Done            int      `db:"done"`
}

func (s ShoppingListItem) IDStr() string { return strconv.Itoa(s.ID) }
func (s ShoppingListItem) AmountStr() string {
	if s.Amount == float64(int(s.Amount)) {
		return strconv.Itoa(int(s.Amount))
	}
	return strconv.FormatFloat(s.Amount, 'f', 1, 64)
}
func (s ShoppingListItem) IsDone() bool { return s.Done == 1 }
func (s ShoppingListItem) DisplayName() string {
	if s.ProductName != "" {
		return s.ProductName
	}
	if s.Note != "" {
		return s.Note
	}
	return "(no name)"
}

func (s ShoppingListItem) PriceStr() string {
	if s.LastPrice != nil && *s.LastPrice > 0 {
		return "$" + strconv.FormatFloat(*s.LastPrice, 'f', 2, 64)
	}
	return ""
}

// ShoppingGroup is a product group with its shopping items, for aisle-sorted display.
type ShoppingGroup struct {
	GroupName string
	Items     []ShoppingListItem
}

// EquipmentItem is the model for equipment CRUD.
type EquipmentItem struct {
	ID                        int    `db:"id"`
	Name                      string `db:"name"`
	Description               string `db:"description"`
	InstructionManualFileName string `db:"instruction_manual_file_name"`
	PictureFileName           string `db:"picture_file_name"`
}

func (e EquipmentItem) IDStr() string { return strconv.Itoa(e.ID) }

// ProductGroup is the model for product group CRUD.
type ProductGroup struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

func (g ProductGroup) IDStr() string { return strconv.Itoa(g.ID) }

// QuantityUnit is the model for quantity unit CRUD.
type QuantityUnit struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	NamePlural  string `db:"name_plural"`
}

func (q QuantityUnit) IDStr() string { return strconv.Itoa(q.ID) }

// ShoppingLocation is the model for store CRUD.
type ShoppingLocation struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

func (s ShoppingLocation) IDStr() string { return strconv.Itoa(s.ID) }

// StockItem is the model for the stock overview.
type StockItem struct {
	ProductID      int     `db:"product_id"`
	ProductName    string  `db:"product_name"`
	LocationName   string  `db:"location_name"`
	Amount         float64 `db:"amount"`
	BestBeforeDate string  `db:"best_before_date"`
	LastPrice      float64 `db:"last_price"`
	MinStockAmount int     `db:"min_stock_amount"`
}

func (s StockItem) ProductIDStr() string { return strconv.Itoa(s.ProductID) }
func (s StockItem) AmountStr() string {
	if s.Amount == float64(int(s.Amount)) {
		return strconv.Itoa(int(s.Amount))
	}
	return strconv.FormatFloat(s.Amount, 'f', 1, 64)
}
func (s StockItem) PriceStr() string {
	if s.LastPrice > 0 {
		return "$" + strconv.FormatFloat(s.LastPrice, 'f', 2, 64)
	}
	return ""
}
func (s StockItem) IsBelowMin() bool { return int(s.Amount) < s.MinStockAmount && s.MinStockAmount > 0 }

// Recipe is the model for recipe list view.
type Recipe struct {
	ID          int      `db:"id"`
	Name        string   `db:"name"`
	Slug        string   `db:"slug"`
	Description string   `db:"description"`
	TotalTime   string   `db:"total_time"`
	PrepTime    string   `db:"prep_time"`
	CookTime    string   `db:"cook_time"`
	Servings    float64  `db:"servings"`
	SourceURL   string   `db:"source_url"`
	Rating      *float64 `db:"rating"`
	Image       string   `db:"image"`
	DateAdded   string   `db:"date_added"`
	HasImage    int      `db:"has_image"`
}

func (r Recipe) IDStr() string { return strconv.Itoa(r.ID) }

// RecipeDetail is the full model for recipe view/edit.
type RecipeDetail struct {
	Recipe
	RecipeYield  string `db:"recipe_yield"`
	Ingredients  []RecipeIngredient
	Instructions []RecipeInstruction
	Tags         []string
	Categories   []string
}

type RecipeIngredient struct {
	ID       int    `db:"id"`
	Position int    `db:"position"`
	Text     string `db:"text"`
	Title    string `db:"title"`
}

type RecipeInstruction struct {
	ID       int    `db:"id"`
	Position int    `db:"position"`
	Title    string `db:"title"`
	Text     string `db:"text"`
}

// User is the model for admin user management.
type User struct {
	ID              int    `db:"id"`
	Username        string `db:"username"`
	FirstName       string `db:"first_name"`
	LastName        string `db:"last_name"`
	IsAdmin         int    `db:"is_admin"`
	PictureFileName string `db:"picture_file_name"`
}

func (u User) IDStr() string { return strconv.Itoa(u.ID) }
func (u User) FullName() string {
	if u.FirstName != "" || u.LastName != "" {
		return strings.TrimSpace(u.FirstName + " " + u.LastName)
	}
	return u.Username
}
func (u User) IsAdminBool() bool { return u.IsAdmin == 1 }

// MealSlot is a configurable meal slot (Breakfast, Lunch, Dinner, etc.)
type MealSlot struct {
	ID         int    `db:"id"`
	Name       string `db:"name"`
	SortNumber int    `db:"sort_number"`
}

func (m MealSlot) IDStr() string { return strconv.Itoa(m.ID) }

// MealPlanEntry is a single entry on the meal plan.
type MealPlanEntry struct {
	ID         int     `db:"id"`
	Day        string  `db:"day"`
	RecipeID   *int    `db:"recipe_id"`
	RecipeName string  `db:"recipe_name"`
	Note       string  `db:"note"`
	Type       string  `db:"type"`
	SectionID  *int    `db:"section_id"`
	SlotName   string  `db:"slot_name"`
	Done       int     `db:"done"`
	Servings   float64 `db:"servings"`
}

func (e MealPlanEntry) IDStr() string { return strconv.Itoa(e.ID) }
func (e MealPlanEntry) IsDone() bool  { return e.Done == 1 }

// MealPlanDay groups entries by slot for a single day.
type MealPlanDay struct {
	Date  string
	Label string // e.g. "Mon Jan 2"
	Slots []MealPlanSlotGroup
}

// MealPlanSlotGroup is a meal slot with its entries for a day.
type MealPlanSlotGroup struct {
	Slot    MealSlot
	Entries []MealPlanEntry
}

// MealPlanIngredient is a merged ingredient line for the shopping export.
type MealPlanIngredient struct {
	Key         string   // normalised ingredient name used for grouping
	DisplayQty  string   // combined quantity string e.g. "1 cup"
	Name        string   // ingredient name as parsed
	Recipes     []string // source recipe names
	ProductID   *int     // matched product ID if found
	ProductName string   // matched product name
	StoreName   string   // shopping_locations.name (the store)
	AisleName   string   // product_groups.name (aisle within store)
	HomeLoc     string   // locations.name (where to put it at home)
}

// TaskList is a named list of tasks owned by a user.
type TaskList struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	UserID      int    `db:"user_id"`
	TaskCount   int    `db:"task_count"`
	DoneCount   int    `db:"done_count"`
}

func (t TaskList) IDStr() string { return strconv.Itoa(t.ID) }

// Task is a single task within a task list.
type Task struct {
	ID               int    `db:"id"`
	TaskListID       int    `db:"task_list_id"`
	Name             string `db:"name"`
	Description      string `db:"description"`
	DueDate          string `db:"due_date"`
	Done             int    `db:"done"`
	DoneTimestamp    string `db:"done_timestamp"`
	AssignedToUserID *int   `db:"assigned_to_user_id"`
	AssignedUsername string `db:"assigned_username"`
}

func (t Task) IDStr() string         { return strconv.Itoa(t.ID) }
func (t Task) IsDone() bool          { return t.Done == 1 }
func (t Task) TaskListIDStr() string { return strconv.Itoa(t.TaskListID) }

// ChoreList is a named list of chores owned by a user.
type ChoreList struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	UserID      int    `db:"user_id"`
	ChoreCount  int    `db:"chore_count"`
	DoneCount   int    `db:"done_count"`
}

func (c ChoreList) IDStr() string { return strconv.Itoa(c.ID) }

// Chore is a single chore within a chore list.
type Chore struct {
	ID               int    `db:"id"`
	ChoreListID      int    `db:"chore_list_id"`
	Name             string `db:"name"`
	Description      string `db:"description"`
	DueDate          string `db:"due_date"`
	Done             int    `db:"done"`
	DoneTimestamp    string `db:"done_timestamp"`
	AssignedToUserID *int   `db:"assigned_to_user_id"`
	AssignedUsername string `db:"assigned_username"`
}

func (c Chore) IDStr() string          { return strconv.Itoa(c.ID) }
func (c Chore) IsDone() bool           { return c.Done == 1 }
func (c Chore) ChoreListIDStr() string { return strconv.Itoa(c.ChoreListID) }

// Battery is the model for battery CRUD.
type Battery struct {
	ID                 int    `db:"id"`
	Name               string `db:"name"`
	Description        string `db:"description"`
	UsedIn             string `db:"used_in"`
	ChargeIntervalDays int    `db:"charge_interval_days"`
	Active             int    `db:"active"`
}

func (b Battery) IDStr() string { return strconv.Itoa(b.ID) }

// BatteryOverview adds last-charged and next-due computed fields.
type BatteryOverview struct {
	Battery
	LastCharged string `db:"last_charged"`
	DaysAgo     int    `db:"days_ago"`
	NextDue     string `db:"next_due"`
	DaysUntil   int    `db:"days_until"`
}

func (b BatteryOverview) RowClass() string {
	if b.ChargeIntervalDays == 0 || b.LastCharged == "" {
		return ""
	}
	if b.DaysUntil < 0 {
		return "table-danger"
	}
	if b.DaysUntil <= 3 {
		return "table-warning"
	}
	return ""
}

func (b BatteryOverview) StatusLabel() string {
	if b.ChargeIntervalDays == 0 {
		return ""
	}
	if b.LastCharged == "" {
		return "Never charged"
	}
	if b.DaysUntil < 0 {
		return "Overdue by " + strconv.Itoa(-b.DaysUntil) + "d"
	}
	if b.DaysUntil == 0 {
		return "Due today"
	}
	if b.DaysUntil <= 3 {
		return "Due in " + strconv.Itoa(b.DaysUntil) + "d"
	}
	return "OK"
}

func (b BatteryOverview) LastChargedDisplay() string {
	if b.LastCharged == "" {
		return "Never"
	}
	if len(b.LastCharged) >= 10 {
		return b.LastCharged[:10]
	}
	return b.LastCharged
}

func (b BatteryOverview) NextDueDisplay() string {
	if b.NextDue == "" {
		return "—"
	}
	return b.NextDue
}

// BatteryChargeCycle is a single charge log entry.
type BatteryChargeCycle struct {
	ID          int    `db:"id"`
	TrackedTime string `db:"tracked_time"`
	Undone      int    `db:"undone"`
}
