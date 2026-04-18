package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"Domicile/templates/partials"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// EquipmentPage renders the equipment list.
func EquipmentPage(c echo.Context) error {
	var items []models.EquipmentItem
	database.DB.Select(&items, "SELECT id, name, IFNULL(description,'') as description, IFNULL(instruction_manual_file_name,'') as instruction_manual_file_name, IFNULL(picture_file_name,'') as picture_file_name FROM equipment ORDER BY name")
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.EquipmentContent(items))
	}
	return render(c, http.StatusOK, templates.EquipmentPage(items))
}

// EquipmentNewForm renders the create form.
func EquipmentNewForm(c echo.Context) error {
	return render(c, http.StatusOK, partials.EquipmentForm(nil))
}

// EquipmentEditForm renders the edit form.
func EquipmentEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.EquipmentItem
	err := database.DB.Get(&item, "SELECT id, name, IFNULL(description,'') as description, IFNULL(instruction_manual_file_name,'') as instruction_manual_file_name, IFNULL(picture_file_name,'') as picture_file_name FROM equipment WHERE id = ?", id)
	if err != nil {
		return c.String(http.StatusNotFound, "Equipment not found")
	}
	return render(c, http.StatusOK, partials.EquipmentForm(&item))
}

// CreateEquipment handles POST /equipment.
func CreateEquipment(c echo.Context) error {
	name := c.FormValue("name")
	description := c.FormValue("description")

	result, err := database.DB.Exec("INSERT INTO equipment (name, description) VALUES (?, ?)", name, description)
	if err != nil {
		return c.String(http.StatusBadRequest, "Equipment already exists or invalid data")
	}

	id, _ := result.LastInsertId()
	idStr := strconv.FormatInt(id, 10)

	// Handle picture upload.
	if pic, err := saveUploadedImage(c, "picture", "equipment"); err == nil && pic != "" {
		database.DB.Exec("UPDATE equipment SET picture_file_name = ? WHERE id = ?", pic, id)
	}

	// Handle manual upload.
	if manual, err := saveUploadedFile(c, "manual", "equipment_manuals", []string{".pdf", ".jpg", ".jpeg", ".png", ".gif", ".webp"}); err == nil && manual != "" {
		database.DB.Exec("UPDATE equipment SET instruction_manual_file_name = ? WHERE id = ?", manual, id)
	}

	_ = idStr
	return EquipmentPage(c)
}

// UpdateEquipment handles PUT /equipment/:id.
func UpdateEquipment(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := c.FormValue("name")
	description := c.FormValue("description")

	_, err := database.DB.Exec("UPDATE equipment SET name=?, description=? WHERE id=?", name, description, id)
	if err != nil {
		return c.String(http.StatusBadRequest, "Update failed: "+err.Error())
	}

	// Handle picture upload.
	if pic, err := saveUploadedImage(c, "picture", "equipment"); err == nil && pic != "" {
		var oldPic string
		database.DB.Get(&oldPic, "SELECT IFNULL(picture_file_name,'') FROM equipment WHERE id = ?", id)
		deleteImage("equipment", oldPic)
		database.DB.Exec("UPDATE equipment SET picture_file_name = ? WHERE id = ?", pic, id)
	}

	// Handle manual upload.
	if manual, err := saveUploadedFile(c, "manual", "equipment_manuals", []string{".pdf", ".jpg", ".jpeg", ".png", ".gif", ".webp"}); err == nil && manual != "" {
		var oldManual string
		database.DB.Get(&oldManual, "SELECT IFNULL(instruction_manual_file_name,'') FROM equipment WHERE id = ?", id)
		if oldManual != "" {
			os.Remove(filepath.Join("data/images/equipment_manuals", oldManual))
		}
		database.DB.Exec("UPDATE equipment SET instruction_manual_file_name = ? WHERE id = ?", manual, id)
	}

	return EquipmentPage(c)
}

// DeleteEquipment handles DELETE /equipment/:id.
func DeleteEquipment(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var picFile, manualFile string
	database.DB.Get(&picFile, "SELECT IFNULL(picture_file_name,'') FROM equipment WHERE id = ?", id)
	database.DB.Get(&manualFile, "SELECT IFNULL(instruction_manual_file_name,'') FROM equipment WHERE id = ?", id)
	if picFile != "" {
		deleteImage("equipment", picFile)
	}
	if manualFile != "" {
		os.Remove(filepath.Join("data/images/equipment_manuals", manualFile))
	}

	database.DB.Exec("DELETE FROM equipment WHERE id = ?", id)
	return c.NoContent(http.StatusOK)
}

// BulkUploadEquipment creates equipment entries from multiple uploaded files.
// Names are derived from filenames by splitting on underscores, dashes, and PascalCase.
func BulkUploadEquipment(c echo.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to read upload")
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.String(http.StatusBadRequest, "No files selected")
	}

	for _, file := range files {
		// Derive name from filename.
		name := deriveNameFromFilename(file.Filename)

		// Check if equipment with this name already exists.
		var exists int
		database.DB.Get(&exists, "SELECT COUNT(*) FROM equipment WHERE name = ?", name)
		if exists > 0 {
			continue // Skip duplicates.
		}

		// Save the file.
		ext := strings.ToLower(filepath.Ext(file.Filename))
		src, err := file.Open()
		if err != nil {
			continue
		}

		dir := filepath.Join("data/images/equipment_manuals")
		os.MkdirAll(dir, 0755)
		savedName := uuid.New().String() + ext
		dst, err := os.Create(filepath.Join(dir, savedName))
		if err != nil {
			src.Close()
			continue
		}
		io.Copy(dst, src)
		dst.Close()
		src.Close()

		// Create the equipment record.
		database.DB.Exec("INSERT INTO equipment (name, description, instruction_manual_file_name) VALUES (?, ?, ?)", name, name, savedName)
	}

	return EquipmentPage(c)
}

// deriveNameFromFilename converts a filename to a human-readable name.
// "Kitchen_Stove_Manual.pdf" → "Kitchen Stove Manual"
// "DishwasherGuide.pdf" → "Dishwasher Guide"
// "my-vacuum.pdf" → "my vacuum"
func deriveNameFromFilename(filename string) string {
	// Remove extension.
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Replace underscores and dashes with spaces.
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")

	// Split PascalCase: insert space before uppercase letters that follow lowercase.
	var result []byte
	for i, ch := range []byte(name) {
		if i > 0 && ch >= 'A' && ch <= 'Z' {
			prev := name[i-1]
			if prev >= 'a' && prev <= 'z' {
				result = append(result, ' ')
			}
		}
		result = append(result, ch)
	}
	name = string(result)

	// Clean up multiple spaces.
	for strings.Contains(name, "  ") {
		name = strings.ReplaceAll(name, "  ", " ")
	}

	return strings.TrimSpace(name)
}
