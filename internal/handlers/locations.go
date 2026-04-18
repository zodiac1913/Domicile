package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"Domicile/templates/partials"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// LocationsPage renders the locations list.
func LocationsPage(c echo.Context) error {
	var locs []models.LocationDetail
	database.DB.Select(&locs, "SELECT id, name, IFNULL(description,'') as description, is_freezer, IFNULL(picture_file_name,'') as picture_file_name FROM locations ORDER BY name")
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.LocationsContent(locs))
	}
	return render(c, http.StatusOK, templates.LocationsPage(locs))
}

// LocationNewForm renders the create form.
func LocationNewForm(c echo.Context) error {
	return render(c, http.StatusOK, partials.LocationForm(nil))
}

// LocationEditForm renders the edit form.
func LocationEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var loc models.LocationDetail
	err := database.DB.Get(&loc, "SELECT id, name, IFNULL(description,'') as description, is_freezer, IFNULL(picture_file_name,'') as picture_file_name FROM locations WHERE id = ?", id)
	if err != nil {
		return c.String(http.StatusNotFound, "Location not found")
	}
	return render(c, http.StatusOK, partials.LocationForm(&loc))
}

// CreateLocation handles POST /locations.
func CreateLocation(c echo.Context) error {
	name := c.FormValue("name")
	description := c.FormValue("description")
	isFreezer := checkboxVal(c, "is_freezer")

	result, err := database.DB.Exec("INSERT INTO locations (name, description, is_freezer) VALUES (?, ?, ?)",
		name, description, isFreezer)
	if err != nil {
		return c.String(http.StatusBadRequest, "Location already exists or invalid data")
	}

	id, _ := result.LastInsertId()
	if pic, err := saveUploadedImage(c, "picture", "locations"); err == nil && pic != "" {
		database.DB.Exec("UPDATE locations SET picture_file_name = ? WHERE id = ?", pic, id)
	}

	return LocationsPage(c)
}

// UpdateLocation handles PUT /locations/:id.
func UpdateLocation(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := c.FormValue("name")
	description := c.FormValue("description")
	isFreezer := checkboxVal(c, "is_freezer")

	_, err := database.DB.Exec("UPDATE locations SET name=?, description=?, is_freezer=? WHERE id=?",
		name, description, isFreezer, id)
	if err != nil {
		return c.String(http.StatusBadRequest, "Update failed: "+err.Error())
	}

	if pic, err := saveUploadedImage(c, "picture", "locations"); err == nil && pic != "" {
		var oldPic string
		database.DB.Get(&oldPic, "SELECT IFNULL(picture_file_name,'') FROM locations WHERE id = ?", id)
		deleteImage("locations", oldPic)
		database.DB.Exec("UPDATE locations SET picture_file_name = ? WHERE id = ?", pic, id)
	}

	return LocationsPage(c)
}

// DeleteLocation handles DELETE /locations/:id.
func DeleteLocation(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var count int
	database.DB.Get(&count, "SELECT COUNT(*) FROM products WHERE location_id = ?", id)
	if count > 0 {
		return c.String(http.StatusBadRequest, "Cannot delete — products are using this location")
	}

	var picFile string
	database.DB.Get(&picFile, "SELECT IFNULL(picture_file_name,'') FROM locations WHERE id = ?", id)
	if picFile != "" {
		deleteImage("locations", picFile)
	}

	database.DB.Exec("DELETE FROM locations WHERE id = ?", id)
	return c.NoContent(http.StatusOK)
}
