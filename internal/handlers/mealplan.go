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
	"time"

	"github.com/labstack/echo/v4"
)

const dateLayout = "2006-01-02"

func getMealSlots() []models.MealSlot {
	var slots []models.MealSlot
	database.DB.Select(&slots, "SELECT id, name, sort_number FROM meal_plan_sections ORDER BY sort_number, name")
	return slots
}

func getMealPlanWeek(startDate string) []models.MealPlanDay {
	start, err := time.Parse(dateLayout, startDate)
	if err != nil {
		start = time.Now()
	}
	// Normalize to Monday of the week.
	for start.Weekday() != time.Monday {
		start = start.AddDate(0, 0, -1)
	}

	slots := getMealSlots()
	var days []models.MealPlanDay

	for i := 0; i < 7; i++ {
		d := start.AddDate(0, 0, i)
		dateStr := d.Format(dateLayout)

		var entries []models.MealPlanEntry
		database.DB.Select(&entries, `
			SELECT mp.id, mp.day, mp.recipe_id, IFNULL(r.name,'') as recipe_name,
				IFNULL(mp.note,'') as note, mp.type,
				mp.section_id, IFNULL(mps.name,'') as slot_name,
				mp.done, IFNULL(r.servings,1) as servings
			FROM meal_plan mp
			LEFT JOIN recipes r ON mp.recipe_id = r.id
			LEFT JOIN meal_plan_sections mps ON mp.section_id = mps.id
			WHERE mp.day = ?
			ORDER BY mps.sort_number, mps.name, mp.id`, dateStr)

		// Group by slot.
		slotMap := make(map[int][]models.MealPlanEntry)
		var unslotted []models.MealPlanEntry
		for _, e := range entries {
			if e.SectionID != nil {
				slotMap[*e.SectionID] = append(slotMap[*e.SectionID], e)
			} else {
				unslotted = append(unslotted, e)
			}
		}

		var slotGroups []models.MealPlanSlotGroup
		for _, s := range slots {
			slotGroups = append(slotGroups, models.MealPlanSlotGroup{
				Slot:    s,
				Entries: slotMap[s.ID],
			})
		}
		// Unslotted entries go in a catch-all group.
		if len(unslotted) > 0 {
			slotGroups = append(slotGroups, models.MealPlanSlotGroup{
				Slot:    models.MealSlot{ID: 0, Name: "Other"},
				Entries: unslotted,
			})
		}

		days = append(days, models.MealPlanDay{
			Date:  dateStr,
			Label: d.Format("Mon Jan 2"),
			Slots: slotGroups,
		})
	}
	return days
}

// MealPlanPage renders the full meal plan page.
func MealPlanPage(c echo.Context) error {
	week := c.QueryParam("week")
	if week == "" {
		week = time.Now().Format(dateLayout)
	}
	days := getMealPlanWeek(week)
	slots := getMealSlots()
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.MealPlanContent(days, slots, week))
	}
	return render(c, http.StatusOK, templates.MealPlanPage(days, slots, week))
}

// MealPlanWeek returns just the week grid (htmx navigation).
func MealPlanWeek(c echo.Context) error {
	week := c.QueryParam("week")
	if week == "" {
		week = time.Now().Format(dateLayout)
	}
	days := getMealPlanWeek(week)
	slots := getMealSlots()
	return render(c, http.StatusOK, partials.MealPlanWeekGrid(days, slots, week))
}

// MealPlanEntryForm renders the add/edit entry form.
func MealPlanEntryForm(c echo.Context) error {
	day := c.QueryParam("day")
	slotID, _ := strconv.Atoi(c.QueryParam("slot_id"))
	slots := getMealSlots()
	var recipes []models.Recipe
	database.DB.Select(&recipes, "SELECT id, name, servings FROM recipes ORDER BY name")
	return render(c, http.StatusOK, partials.MealPlanEntryForm(nil, day, slotID, slots, recipes))
}

// MealPlanEntryEditForm renders the edit form for an existing entry.
func MealPlanEntryEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var entry models.MealPlanEntry
	err := database.DB.Get(&entry, `
		SELECT mp.id, mp.day, mp.recipe_id, IFNULL(r.name,'') as recipe_name,
			IFNULL(mp.note,'') as note, mp.type,
			mp.section_id, IFNULL(mps.name,'') as slot_name,
			mp.done, IFNULL(r.servings,1) as servings
		FROM meal_plan mp
		LEFT JOIN recipes r ON mp.recipe_id = r.id
		LEFT JOIN meal_plan_sections mps ON mp.section_id = mps.id
		WHERE mp.id = ?`, id)
	if err != nil {
		return c.String(http.StatusNotFound, "Entry not found")
	}
	slots := getMealSlots()
	var recipes []models.Recipe
	database.DB.Select(&recipes, "SELECT id, name, servings FROM recipes ORDER BY name")
	return render(c, http.StatusOK, partials.MealPlanEntryForm(&entry, entry.Day, 0, slots, recipes))
}

// CreateMealPlanEntry handles POST /meal-plan/entries.
func CreateMealPlanEntry(c echo.Context) error {
	day := c.FormValue("day")
	entryType := c.FormValue("type")
	note := c.FormValue("note")
	slotIDStr := c.FormValue("slot_id")
	recipeIDStr := c.FormValue("recipe_id")

	var slotID *int
	if slotIDStr != "" && slotIDStr != "0" {
		sid, _ := strconv.Atoi(slotIDStr)
		slotID = &sid
	}
	var recipeID *int
	if recipeIDStr != "" && recipeIDStr != "0" {
		rid, _ := strconv.Atoi(recipeIDStr)
		recipeID = &rid
		entryType = "recipe"
	} else {
		entryType = "note"
	}

	database.DB.Exec(`
		INSERT INTO meal_plan (day, recipe_id, note, type, section_id)
		VALUES (?, ?, NULLIF(?, ''), ?, ?)`,
		day, recipeID, note, entryType, slotID)

	week := weekStart(day)
	days := getMealPlanWeek(week)
	slots := getMealSlots()
	return render(c, http.StatusOK, partials.MealPlanWeekGrid(days, slots, week))
}

// UpdateMealPlanEntry handles POST /meal-plan/entries/:id/update.
func UpdateMealPlanEntry(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	note := c.FormValue("note")
	slotIDStr := c.FormValue("slot_id")
	recipeIDStr := c.FormValue("recipe_id")
	entryType := "note"

	var slotID *int
	if slotIDStr != "" && slotIDStr != "0" {
		sid, _ := strconv.Atoi(slotIDStr)
		slotID = &sid
	}
	var recipeID *int
	if recipeIDStr != "" && recipeIDStr != "0" {
		rid, _ := strconv.Atoi(recipeIDStr)
		recipeID = &rid
		entryType = "recipe"
	}

	var day string
	database.DB.Get(&day, "SELECT day FROM meal_plan WHERE id = ?", id)
	database.DB.Exec(`UPDATE meal_plan SET recipe_id=?, note=NULLIF(?,''), type=?, section_id=? WHERE id=?`,
		recipeID, note, entryType, slotID, id)

	week := weekStart(day)
	days := getMealPlanWeek(week)
	slots := getMealSlots()
	return render(c, http.StatusOK, partials.MealPlanWeekGrid(days, slots, week))
}

// DeleteMealPlanEntry handles DELETE /meal-plan/entries/:id.
func DeleteMealPlanEntry(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var day string
	database.DB.Get(&day, "SELECT day FROM meal_plan WHERE id = ?", id)
	database.DB.Exec("DELETE FROM meal_plan WHERE id = ?", id)

	week := weekStart(day)
	days := getMealPlanWeek(week)
	slots := getMealSlots()
	return render(c, http.StatusOK, partials.MealPlanWeekGrid(days, slots, week))
}

// ToggleMealPlanDone toggles the done state of an entry.
func ToggleMealPlanDone(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("UPDATE meal_plan SET done = CASE WHEN done=1 THEN 0 ELSE 1 END WHERE id=?", id)
	var day string
	database.DB.Get(&day, "SELECT day FROM meal_plan WHERE id = ?", id)
	week := weekStart(day)
	days := getMealPlanWeek(week)
	slots := getMealSlots()
	return render(c, http.StatusOK, partials.MealPlanWeekGrid(days, slots, week))
}

// MealPlanIngredientsForm renders the ingredient export form.
func MealPlanIngredientsForm(c echo.Context) error {
	today := time.Now().Format(dateLayout)
	return render(c, http.StatusOK, partials.MealPlanIngredientsForm(today))
}

// MealPlanExportIngredients collects, merges, and enriches ingredients from planned recipes.
func MealPlanExportIngredients(c echo.Context) error {
	fromStr := c.FormValue("from")
	toStr := c.FormValue("to")
	if fromStr == "" || toStr == "" {
		return c.String(http.StatusBadRequest, "Date range required")
	}

	var recipeIDs []int
	database.DB.Select(&recipeIDs, `
		SELECT DISTINCT recipe_id FROM meal_plan
		WHERE day >= ? AND day <= ? AND recipe_id IS NOT NULL`, fromStr, toStr)

	if len(recipeIDs) == 0 {
		return render(c, http.StatusOK, partials.MealPlanIngredientsResult(nil, "No recipes planned in that range."))
	}

	// Load all products once for matching.
	type productRow struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Store string `db:"store"`
		Aisle string `db:"aisle"`
		Home  string `db:"home"`
	}
	var allProducts []productRow
	database.DB.Select(&allProducts, `
		SELECT p.id, p.name,
			IFNULL(sl.name,'') as store,
			IFNULL(pg.name,'') as aisle,
			IFNULL(l.name,'') as home
		FROM products p
		LEFT JOIN product_groups pg ON p.product_group_id = pg.id
		LEFT JOIN shopping_locations sl ON pg.store_id = sl.id
		LEFT JOIN locations l ON p.location_id = l.id
		WHERE p.active = 1 ORDER BY p.name`)

	// Accumulate parsed ingredients keyed by normalised name.
	type accumEntry struct {
		qtyByUnit map[string]float64
		recipes   []string
	}
	accum := map[string]*accumEntry{}
	var keyOrder []string

	for _, rid := range recipeIDs {
		var recipeName string
		database.DB.Get(&recipeName, "SELECT name FROM recipes WHERE id = ?", rid)
		var lines []string
		database.DB.Select(&lines, "SELECT text FROM recipe_ingredients WHERE recipe_id = ? ORDER BY position", rid)
		for _, line := range lines {
			qty, unit, name := parseIngredientLine(line)
			key := normaliseIngredientKey(name)
			if _, exists := accum[key]; !exists {
				accum[key] = &accumEntry{qtyByUnit: map[string]float64{}}
				keyOrder = append(keyOrder, key)
			}
			accum[key].qtyByUnit[unit] += qty
			// Add recipe name if not already listed.
			found := false
			for _, r := range accum[key].recipes {
				if r == recipeName {
					found = true
					break
				}
			}
			if !found {
				accum[key].recipes = append(accum[key].recipes, recipeName)
			}
		}
	}

	var merged []models.MealPlanIngredient
	for _, key := range keyOrder {
		e := accum[key]
		mi := models.MealPlanIngredient{
			Key:     key,
			Name:    key,
			Recipes: e.recipes,
			DisplayQty: buildQtyString(e.qtyByUnit),
		}
		// Product lookup — find best match.
		for _, p := range allProducts {
			if ingredientMatchesProduct(key, strings.ToLower(p.Name)) {
				pid := p.ID
				mi.ProductID = &pid
				mi.ProductName = p.Name
				mi.StoreName = p.Store
				mi.AisleName = p.Aisle
				mi.HomeLoc = p.Home
				break
			}
		}
		merged = append(merged, mi)
	}

	return render(c, http.StatusOK, partials.MealPlanIngredientsResult(merged, ""))
}

// MealPlanAddToShopping adds selected ingredients to the shopping list.
func MealPlanAddToShopping(c echo.Context) error {
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse form")
	}
	items := c.Request().Form["items"]
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		// item is encoded as "productID:qty:note" or "0:qty:note" when no product.
		parts := strings.SplitN(item, ":", 3)
		if len(parts) != 3 {
			continue
		}
		pidStr, qtyStr, note := parts[0], parts[1], parts[2]
		qty, _ := strconv.ParseFloat(qtyStr, 64)
		if qty <= 0 {
			qty = 1
		}
		var productID *int
		if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
			productID = &pid
		}
		database.DB.Exec(
			`INSERT INTO shopping_list (shopping_list_id, product_id, note, amount) VALUES (1, ?, NULLIF(?,''), ?)`,
			productID, note, qty,
		)
	}
	c.Response().Header().Set("HX-Redirect", "/shopping")
	return c.NoContent(http.StatusOK)
}

// --- ingredient parsing helpers ---

// knownUnits maps unit strings (and common abbreviations/plurals) to a canonical form.
var knownUnits = map[string]string{
	"cup": "cup", "cups": "cup", "c": "cup",
	"tablespoon": "tbsp", "tablespoons": "tbsp", "tbsp": "tbsp", "tbs": "tbsp",
	"teaspoon": "tsp", "teaspoons": "tsp", "tsp": "tsp",
	"pound": "lb", "pounds": "lb", "lb": "lb", "lbs": "lb",
	"ounce": "oz", "ounces": "oz", "oz": "oz",
	"gram": "g", "grams": "g", "g": "g",
	"kilogram": "kg", "kilograms": "kg", "kg": "kg",
	"liter": "l", "liters": "l", "litre": "l", "litres": "l", "l": "l",
	"milliliter": "ml", "milliliters": "ml", "ml": "ml",
	"clove": "clove", "cloves": "clove",
	"slice": "slice", "slices": "slice",
	"piece": "piece", "pieces": "piece",
	"can": "can", "cans": "can",
	"package": "pkg", "packages": "pkg", "pkg": "pkg",
	"bunch": "bunch", "bunches": "bunch",
	"sprig": "sprig", "sprigs": "sprig",
	"pinch": "pinch",
	"dash": "dash",
	"handful": "handful",
	"head": "head", "heads": "head",
	"stalk": "stalk", "stalks": "stalk",
	"strip": "strip", "strips": "strip",
}

// fractionMap converts common unicode/text fractions to float64.
var fractionMap = map[string]float64{
	"½": 0.5, "⅓": 1.0 / 3, "⅔": 2.0 / 3,
	"¼": 0.25, "¾": 0.75,
	"⅛": 0.125, "⅜": 0.375, "⅝": 0.625, "⅞": 0.875,
	"1/2": 0.5, "1/3": 1.0 / 3, "2/3": 2.0 / 3,
	"1/4": 0.25, "3/4": 0.75,
	"1/8": 0.125, "3/8": 0.375, "5/8": 0.625, "7/8": 0.875,
}

var ingredientLineRe = regexp.MustCompile(
	`^([\d\s½⅓⅔¼¾⅛⅜⅝⅞/\.]+)?\s*([a-zA-Z]+\.?)?\s*(.+)$`,
)

// parseIngredientLine splits "1/2 cup carrots, diced" → (0.5, "cup", "carrots")
func parseIngredientLine(line string) (qty float64, unit, name string) {
	line = strings.TrimSpace(line)
	m := ingredientLineRe.FindStringSubmatch(line)
	if m == nil {
		return 1, "", line
	}
	qtyStr := strings.TrimSpace(m[1])
	unitStr := strings.ToLower(strings.TrimSpace(m[2]))
	nameRaw := strings.TrimSpace(m[3])

	// Parse quantity.
	qty = parseQty(qtyStr)
	if qty == 0 {
		qty = 1
	}

	// Resolve unit.
	if canonical, ok := knownUnits[unitStr]; ok {
		unit = canonical
	} else {
		// unitStr wasn't a known unit — it's part of the name.
		if unitStr != "" {
			nameRaw = unitStr + " " + nameRaw
		}
		unit = ""
	}

	// Strip trailing descriptors after comma.
	name = strings.TrimSpace(strings.SplitN(nameRaw, ",", 2)[0])
	return
}

func parseQty(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// Check unicode/slash fractions first.
	for frac, val := range fractionMap {
		if strings.Contains(s, frac) {
			// Handle "1 1/2" style.
			rest := strings.TrimSpace(strings.Replace(s, frac, "", 1))
			whole, _ := strconv.ParseFloat(rest, 64)
			return whole + val
		}
	}
	// Plain float/int.
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func normaliseIngredientKey(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	// Strip common trailing words.
	for _, suffix := range []string{" fresh", " dried", " chopped", " diced", " minced",
		" sliced", " grated", " peeled", " cooked", " raw", " large", " small", " medium"} {
		s = strings.TrimSuffix(s, suffix)
	}
	// Naive depluralize: trailing 's' on words > 3 chars.
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") {
			words[i] = w[:len(w)-1]
		}
	}
	return strings.Join(words, " ")
}

func buildQtyString(qtyByUnit map[string]float64) string {
	var parts []string
	for unit, qty := range qtyByUnit {
		qtyStr := formatQty(qty)
		if unit == "" {
			parts = append(parts, qtyStr)
		} else {
			parts = append(parts, qtyStr+" "+unit)
		}
	}
	return strings.Join(parts, " + ")
}

func formatQty(qty float64) string {
	if qty == float64(int(qty)) {
		return strconv.Itoa(int(qty))
	}
	return strconv.FormatFloat(qty, 'f', 2, 64)
}

// ingredientMatchesProduct returns true if the ingredient key is contained in
// the product name or vice-versa (simple substring match on normalised strings).
func ingredientMatchesProduct(ingKey, productNameLower string) bool {
	if ingKey == "" || productNameLower == "" {
		return false
	}
	return strings.Contains(productNameLower, ingKey) || strings.Contains(ingKey, productNameLower)
}

// --- Slot (meal_plan_sections) management ---

// MealPlanSlotsPage renders the slot management page.
func MealPlanSlotsPage(c echo.Context) error {
	slots := getMealSlots()
	if isHTMX(c) {
		return render(c, http.StatusOK, partials.MealPlanSlotsPanel(slots))
	}
	return render(c, http.StatusOK, templates.MealPlanSlotsPage(slots))
}

// CreateMealSlot handles POST /meal-plan/slots.
func CreateMealSlot(c echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	sort, _ := strconv.Atoi(c.FormValue("sort_number"))
	if name == "" {
		return c.String(http.StatusBadRequest, "Name required")
	}
	database.DB.Exec("INSERT INTO meal_plan_sections (name, sort_number) VALUES (?, ?)", name, sort)
	return render(c, http.StatusOK, partials.MealPlanSlotsPanel(getMealSlots()))
}

// UpdateMealSlot handles POST /meal-plan/slots/:id/update.
func UpdateMealSlot(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := strings.TrimSpace(c.FormValue("name"))
	sort, _ := strconv.Atoi(c.FormValue("sort_number"))
	database.DB.Exec("UPDATE meal_plan_sections SET name=?, sort_number=? WHERE id=?", name, sort, id)
	return render(c, http.StatusOK, partials.MealPlanSlotsPanel(getMealSlots()))
}

// DeleteMealSlot handles DELETE /meal-plan/slots/:id.
func DeleteMealSlot(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("UPDATE meal_plan SET section_id=NULL WHERE section_id=?", id)
	database.DB.Exec("DELETE FROM meal_plan_sections WHERE id=?", id)
	return render(c, http.StatusOK, partials.MealPlanSlotsPanel(getMealSlots()))
}

// weekStart returns the Monday of the week containing the given date string.
func weekStart(dateStr string) string {
	t, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		return time.Now().Format(dateLayout)
	}
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, -1)
	}
	return t.Format(dateLayout)
}
