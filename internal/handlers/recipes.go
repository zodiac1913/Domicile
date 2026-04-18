package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"Domicile/templates/partials"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/html"
)

func getRecipeDetail(id int) (models.RecipeDetail, error) {
	var r models.RecipeDetail
	err := database.DB.Get(&r, `
		SELECT id, name, IFNULL(slug,'') as slug, IFNULL(description,'') as description,
			IFNULL(image,'') as image, IFNULL(total_time,'') as total_time,
			IFNULL(prep_time,'') as prep_time, IFNULL(cook_time,'') as cook_time,
			IFNULL(recipe_yield,'') as recipe_yield, servings,
			IFNULL(source_url,'') as source_url, rating,
			IFNULL(date_added,'') as date_added
		FROM recipes WHERE id = ?`, id)
	if err != nil {
		return r, err
	}
	database.DB.Select(&r.Ingredients, "SELECT id, position, text, IFNULL(title,'') as title FROM recipe_ingredients WHERE recipe_id = ? ORDER BY position", id)
	database.DB.Select(&r.Instructions, "SELECT id, position, IFNULL(title,'') as title, text FROM recipe_instructions WHERE recipe_id = ? ORDER BY position", id)
	database.DB.Select(&r.Tags, "SELECT tag FROM recipe_tags WHERE recipe_id = ? ORDER BY tag", id)
	database.DB.Select(&r.Categories, "SELECT category FROM recipe_categories WHERE recipe_id = ? ORDER BY category", id)
	return r, nil
}

// RecipesPage renders the recipe list.
func RecipesPage(c echo.Context) error {
	var recipes []models.Recipe
	database.DB.Select(&recipes, `
		SELECT id, name, IFNULL(slug,'') as slug, IFNULL(description,'') as description,
			IFNULL(total_time,'') as total_time, IFNULL(prep_time,'') as prep_time,
			IFNULL(cook_time,'') as cook_time, servings,
			IFNULL(source_url,'') as source_url, rating,
			IFNULL(image,'') as image, IFNULL(date_added,'') as date_added,
			CASE WHEN image_data IS NOT NULL THEN 1 ELSE 0 END as has_image
		FROM recipes ORDER BY name`)
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.RecipesContent(recipes))
	}
	return render(c, http.StatusOK, templates.RecipesPage(recipes))
}

// RecipeView renders a single recipe.
func RecipeView(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	r, err := getRecipeDetail(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Recipe not found")
	}
	if isHTMX(c) {
		return render(c, http.StatusOK, partials.RecipeView(r))
	}
	return render(c, http.StatusOK, templates.RecipeViewPage(r))
}

// RecipeNewForm renders the create form.
func RecipeNewForm(c echo.Context) error {
	return render(c, http.StatusOK, partials.RecipeForm(nil))
}

// RecipeEditForm renders the edit form.
func RecipeEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	r, err := getRecipeDetail(id)
	if err != nil {
		return c.String(http.StatusNotFound, "Recipe not found")
	}
	return render(c, http.StatusOK, partials.RecipeForm(&r))
}

// CreateRecipe handles POST /recipes.
func CreateRecipe(c echo.Context) error {
	name := c.FormValue("name")
	description := c.FormValue("description")
	prepTime := c.FormValue("prep_time")
	cookTime := c.FormValue("cook_time")
	totalTime := c.FormValue("total_time")
	recipeYield := c.FormValue("recipe_yield")
	servings := parseFloat(c.FormValue("servings"), 1)
	sourceURL := c.FormValue("source_url")
	ingredients := c.FormValue("ingredients")
	instructions := c.FormValue("instructions")
	tags := c.FormValue("tags")

	slug := slugify(name)

	result, err := database.DB.Exec(`
		INSERT INTO recipes (name, slug, description, prep_time, cook_time, total_time,
			recipe_yield, servings, source_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		name, slug, description, prepTime, cookTime, totalTime,
		recipeYield, servings, sourceURL)
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to create recipe")
	}

	recipeID, _ := result.LastInsertId()
	saveIngredientsAndInstructions(int(recipeID), ingredients, instructions, tags)
	saveRecipeImage(c, int(recipeID))

	return RecipesPage(c)
}

// UpdateRecipe handles PUT /recipes/:id.
func UpdateRecipe(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := c.FormValue("name")
	description := c.FormValue("description")
	prepTime := c.FormValue("prep_time")
	cookTime := c.FormValue("cook_time")
	totalTime := c.FormValue("total_time")
	recipeYield := c.FormValue("recipe_yield")
	servings := parseFloat(c.FormValue("servings"), 1)
	sourceURL := c.FormValue("source_url")
	ingredients := c.FormValue("ingredients")
	instructions := c.FormValue("instructions")
	tags := c.FormValue("tags")

	slug := slugify(name)

	database.DB.Exec(`
		UPDATE recipes SET name=?, slug=?, description=?, prep_time=?, cook_time=?,
			total_time=?, recipe_yield=?, servings=?, source_url=?
		WHERE id=?`,
		name, slug, description, prepTime, cookTime, totalTime,
		recipeYield, servings, sourceURL, id)

	// Clear and re-insert ingredients, instructions, tags.
	database.DB.Exec("DELETE FROM recipe_ingredients WHERE recipe_id = ?", id)
	database.DB.Exec("DELETE FROM recipe_instructions WHERE recipe_id = ?", id)
	database.DB.Exec("DELETE FROM recipe_tags WHERE recipe_id = ?", id)
	saveIngredientsAndInstructions(id, ingredients, instructions, tags)
	saveRecipeImage(c, id)

	return RecipesPage(c)
}

// DeleteRecipe handles DELETE /recipes/:id.
func DeleteRecipe(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM recipe_ingredients WHERE recipe_id = ?", id)
	database.DB.Exec("DELETE FROM recipe_instructions WHERE recipe_id = ?", id)
	database.DB.Exec("DELETE FROM recipe_tags WHERE recipe_id = ?", id)
	database.DB.Exec("DELETE FROM recipe_categories WHERE recipe_id = ?", id)
	database.DB.Exec("DELETE FROM recipes WHERE id = ?", id)
	return c.NoContent(http.StatusOK)
}

func saveRecipeImage(c echo.Context, recipeID int) {
	file, err := c.FormFile("recipe_image")
	if err != nil || file == nil {
		return
	}
	src, err := file.Open()
	if err != nil {
		return
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil || len(data) == 0 {
		return
	}
	contentType := http.DetectContentType(data)
	database.DB.Exec("UPDATE recipes SET image_data = ?, image_type = ? WHERE id = ?", data, contentType, recipeID)
}

func saveIngredientsAndInstructions(recipeID int, ingredients, instructions, tags string) {
	// Ingredients: one per line.
	for i, line := range strings.Split(ingredients, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		database.DB.Exec("INSERT INTO recipe_ingredients (recipe_id, position, text) VALUES (?, ?, ?)",
			recipeID, i, line)
	}
	// Instructions: one per line.
	for i, line := range strings.Split(instructions, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		database.DB.Exec("INSERT INTO recipe_instructions (recipe_id, position, text) VALUES (?, ?, ?)",
			recipeID, i, line)
	}
	// Tags: comma separated.
	for _, tag := range strings.Split(tags, ",") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		database.DB.Exec("INSERT OR IGNORE INTO recipe_tags (recipe_id, tag) VALUES (?, ?)",
			recipeID, tag)
	}
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	reg := regexp.MustCompile(`[^\w\s-]`)
	s = reg.ReplaceAllString(s, "")
	s = regexp.MustCompile(`[\s_]+`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// ScrapeRecipe fetches a URL and extracts recipe data from JSON-LD.
func ScrapeRecipe(c echo.Context) error {
	url := c.FormValue("url")
	if url == "" {
		return c.String(http.StatusBadRequest, "No URL provided")
	}

	resp, err := http.Get(url)
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to fetch URL: "+err.Error())
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	pageHTML := string(body)

	// Extract JSON-LD blocks.
	recipe := extractRecipeFromJSONLD(pageHTML)
	if recipe == nil {
		return c.String(http.StatusBadRequest, "No recipe found on this page")
	}

	// Save to DB.
	name := getString(recipe, "name")
	if name == "" {
		return c.String(http.StatusBadRequest, "Recipe has no name")
	}

	slug := slugify(name)
	desc := getString(recipe, "description")
	prepTime := formatDuration(getString(recipe, "prepTime"))
	cookTime := formatDuration(getString(recipe, "cookTime"))
	totalTime := formatDuration(getString(recipe, "totalTime"))
	recipeYield := getString(recipe, "recipeYield")
	image := extractImage(recipe)

	result, err := database.DB.Exec(`
		INSERT INTO recipes (name, slug, description, image, prep_time, cook_time,
			total_time, recipe_yield, servings, source_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?)`,
		name, slug, desc, image, prepTime, cookTime, totalTime, recipeYield, url)
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to save recipe: "+err.Error())
	}

	recipeID, _ := result.LastInsertId()

	// Ingredients.
	if ingrs, ok := recipe["recipeIngredient"].([]interface{}); ok {
		for i, v := range ingrs {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				database.DB.Exec("INSERT INTO recipe_ingredients (recipe_id, position, text) VALUES (?, ?, ?)",
					recipeID, i, strings.TrimSpace(s))
			}
		}
	}

	// Instructions.
	if instrs, ok := recipe["recipeInstructions"].([]interface{}); ok {
		for i, v := range instrs {
			switch step := v.(type) {
			case map[string]interface{}:
				text := getString(step, "text")
				if text != "" {
					database.DB.Exec("INSERT INTO recipe_instructions (recipe_id, position, text) VALUES (?, ?, ?)",
						recipeID, i, strings.TrimSpace(text))
				}
			case string:
				if strings.TrimSpace(step) != "" {
					database.DB.Exec("INSERT INTO recipe_instructions (recipe_id, position, text) VALUES (?, ?, ?)",
						recipeID, i, strings.TrimSpace(step))
				}
			}
		}
	}

	// Categories from recipeCategory.
	if cats, ok := recipe["recipeCategory"].([]interface{}); ok {
		for _, v := range cats {
			if s, ok := v.(string); ok && s != "" {
				database.DB.Exec("INSERT OR IGNORE INTO recipe_categories (recipe_id, category) VALUES (?, ?)",
					recipeID, strings.TrimSpace(s))
			}
		}
	}

	// Keywords as tags.
	if kw, ok := recipe["keywords"].(string); ok {
		for _, tag := range strings.Split(kw, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				database.DB.Exec("INSERT OR IGNORE INTO recipe_tags (recipe_id, tag) VALUES (?, ?)",
					recipeID, tag)
			}
		}
	}

	// Auto-fetch the image blob if we have a URL.
	if image != "" && strings.HasPrefix(image, "http") {
		client := &http.Client{Timeout: 15 * time.Second}
		req, _ := http.NewRequest("GET", image, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Domicile/1.0)")
		if imgResp, err := client.Do(req); err == nil {
			defer imgResp.Body.Close()
			if imgData, err := io.ReadAll(imgResp.Body); err == nil && len(imgData) > 0 {
				contentType := http.DetectContentType(imgData)
				database.DB.Exec("UPDATE recipes SET image_data = ?, image_type = ? WHERE id = ?", imgData, contentType, recipeID)
			}
		}
	}

	return RecipesPage(c)
}

func extractRecipeFromJSONLD(pageHTML string) map[string]interface{} {
	// Find all <script type="application/ld+json"> blocks.
	doc, err := html.Parse(strings.NewReader(pageHTML))
	if err != nil {
		return nil
	}

	var jsonBlocks []string
	var findScripts func(*html.Node)
	findScripts = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, attr := range n.Attr {
				if attr.Key == "type" && attr.Val == "application/ld+json" {
					if n.FirstChild != nil {
						jsonBlocks = append(jsonBlocks, n.FirstChild.Data)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findScripts(c)
		}
	}
	findScripts(doc)

	for _, block := range jsonBlocks {
		recipe := findRecipeInJSON(block)
		if recipe != nil {
			return recipe
		}
	}
	return nil
}

func findRecipeInJSON(jsonStr string) map[string]interface{} {
	// Try as single object.
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err == nil {
		if isRecipe(obj) {
			return obj
		}
		// Check @graph.
		if graph, ok := obj["@graph"].([]interface{}); ok {
			for _, item := range graph {
				if m, ok := item.(map[string]interface{}); ok && isRecipe(m) {
					return m
				}
			}
		}
	}

	// Try as array.
	var arr []interface{}
	if err := json.Unmarshal([]byte(jsonStr), &arr); err == nil {
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok && isRecipe(m) {
				return m
			}
		}
	}

	return nil
}

func isRecipe(obj map[string]interface{}) bool {
	t, _ := obj["@type"].(string)
	if t == "Recipe" {
		return true
	}
	if types, ok := obj["@type"].([]interface{}); ok {
		for _, tt := range types {
			if s, ok := tt.(string); ok && s == "Recipe" {
				return true
			}
		}
	}
	return false
}

func getString(obj map[string]interface{}, key string) string {
	if v, ok := obj[key]; ok {
		switch s := v.(type) {
		case string:
			return strings.TrimSpace(s)
		case []interface{}:
			if len(s) > 0 {
				if str, ok := s[0].(string); ok {
					return strings.TrimSpace(str)
				}
			}
		}
	}
	return ""
}

func extractImage(recipe map[string]interface{}) string {
	if img, ok := recipe["image"]; ok {
		switch v := img.(type) {
		case string:
			return v
		case []interface{}:
			if len(v) > 0 {
				if s, ok := v[0].(string); ok {
					return s
				}
			}
		case map[string]interface{}:
			if url, ok := v["url"].(string); ok {
				return url
			}
		}
	}
	return ""
}

func formatDuration(iso string) string {
	// Convert ISO 8601 duration (PT1H30M) to human readable.
	if iso == "" || !strings.HasPrefix(iso, "PT") {
		return iso
	}
	iso = strings.TrimPrefix(iso, "PT")
	var parts []string
	if i := strings.Index(iso, "H"); i >= 0 {
		parts = append(parts, iso[:i]+" hr")
		iso = iso[i+1:]
	}
	if i := strings.Index(iso, "M"); i >= 0 {
		parts = append(parts, iso[:i]+" min")
	}
	return strings.Join(parts, " ")
}

// RecipeImage serves a recipe's blob image.
func RecipeImage(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var img struct {
		Data []byte `db:"image_data"`
		Type string `db:"image_type"`
	}
	err := database.DB.Get(&img, "SELECT image_data, IFNULL(image_type,'image/jpeg') as image_type FROM recipes WHERE id = ? AND image_data IS NOT NULL", id)
	if err != nil || len(img.Data) == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	return c.Blob(http.StatusOK, img.Type, img.Data)
}

// RecipeImageUpload handles uploading an image for a recipe.
func RecipeImageUpload(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	file, err := c.FormFile("image")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	src, err := file.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	contentType := http.DetectContentType(data)
	database.DB.Exec("UPDATE recipes SET image_data = ?, image_type = ? WHERE id = ?", data, contentType, id)

	c.Response().Header().Set("HX-Redirect", "/recipes/"+strconv.Itoa(id))
	return c.NoContent(http.StatusOK)
}

// RecipeFetchImage fetches the image from the recipe's source URL and stores it as a blob.
func RecipeFetchImage(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var recipe struct {
		Image     string `db:"image"`
		SourceURL string `db:"source_url"`
	}
	database.DB.Get(&recipe, "SELECT IFNULL(image,'') as image, IFNULL(source_url,'') as source_url FROM recipes WHERE id = ?", id)

	imageURL := recipe.Image
	if !strings.HasPrefix(imageURL, "http") {
		if recipe.SourceURL == "" {
			return c.String(http.StatusBadRequest, "No image URL available")
		}
		// Re-scrape from source URL.
		client := &http.Client{Timeout: 15 * time.Second}
		req, _ := http.NewRequest("GET", recipe.SourceURL, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Domicile/1.0)")
		resp, err := client.Do(req)
		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to fetch source: "+err.Error())
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		scraped := extractRecipeFromJSONLD(string(body))
		if scraped != nil {
			imageURL = extractImage(scraped)
		}
		if !strings.HasPrefix(imageURL, "http") {
			return c.String(http.StatusBadRequest, "Could not find image URL on page")
		}
		database.DB.Exec("UPDATE recipes SET image = ? WHERE id = ?", imageURL, id)
	}

	// Fetch the image with User-Agent.
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", imageURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Domicile/1.0)")
	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to fetch image: "+err.Error())
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	contentType := http.DetectContentType(data)
	database.DB.Exec("UPDATE recipes SET image_data = ?, image_type = ? WHERE id = ?", data, contentType, id)

	c.Response().Header().Set("HX-Redirect", "/recipes/"+strconv.Itoa(id))
	return c.NoContent(http.StatusOK)
}
