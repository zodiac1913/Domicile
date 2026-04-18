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

// ==================== Product Groups ====================

func ProductGroupsPage(c echo.Context) error {
	var items []models.ProductGroup
	database.DB.Select(&items, "SELECT id, name, IFNULL(description,'') as description FROM product_groups ORDER BY name")
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.ProductGroupsContent(items))
	}
	return render(c, http.StatusOK, templates.ProductGroupsPage(items))
}

func ProductGroupNewForm(c echo.Context) error {
	return render(c, http.StatusOK, partials.ProductGroupForm(nil))
}

func ProductGroupEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.ProductGroup
	err := database.DB.Get(&item, "SELECT id, name, IFNULL(description,'') as description FROM product_groups WHERE id = ?", id)
	if err != nil {
		return c.String(http.StatusNotFound, "Product group not found")
	}
	return render(c, http.StatusOK, partials.ProductGroupForm(&item))
}

func CreateProductGroup(c echo.Context) error {
	name := c.FormValue("name")
	description := c.FormValue("description")
	storeID, _ := strconv.Atoi(c.FormValue("store_id"))
	_, err := database.DB.Exec("INSERT INTO product_groups (name, description, store_id) VALUES (?, ?, ?)", name, description, storeID)
	if err != nil {
		return c.String(http.StatusBadRequest, "Product group already exists")
	}
	return c.NoContent(http.StatusOK)
}

func UpdateProductGroup(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := c.FormValue("name")
	description := c.FormValue("description")
	database.DB.Exec("UPDATE product_groups SET name=?, description=? WHERE id=?", name, description, id)
	return c.NoContent(http.StatusOK)
}

func DeleteProductGroup(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var count int
	database.DB.Get(&count, "SELECT COUNT(*) FROM products WHERE product_group_id = ?", id)
	if count > 0 {
		return c.String(http.StatusBadRequest, "Cannot delete — products are using this group")
	}
	database.DB.Exec("DELETE FROM product_groups WHERE id = ?", id)
	return c.NoContent(http.StatusOK)
}

// ==================== Quantity Units ====================

func QuantityUnitsPage(c echo.Context) error {
	var items []models.QuantityUnit
	database.DB.Select(&items, "SELECT id, name, IFNULL(description,'') as description, IFNULL(name_plural,'') as name_plural FROM quantity_units ORDER BY name")
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.QuantityUnitsContent(items))
	}
	return render(c, http.StatusOK, templates.QuantityUnitsPage(items))
}

func QuantityUnitNewForm(c echo.Context) error {
	return render(c, http.StatusOK, partials.QuantityUnitForm(nil))
}

func QuantityUnitEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.QuantityUnit
	err := database.DB.Get(&item, "SELECT id, name, IFNULL(description,'') as description, IFNULL(name_plural,'') as name_plural FROM quantity_units WHERE id = ?", id)
	if err != nil {
		return c.String(http.StatusNotFound, "Quantity unit not found")
	}
	return render(c, http.StatusOK, partials.QuantityUnitForm(&item))
}

func CreateQuantityUnit(c echo.Context) error {
	name := c.FormValue("name")
	description := c.FormValue("description")
	namePlural := c.FormValue("name_plural")
	_, err := database.DB.Exec("INSERT INTO quantity_units (name, description, name_plural) VALUES (?, ?, ?)", name, description, namePlural)
	if err != nil {
		return c.String(http.StatusBadRequest, "Quantity unit already exists")
	}
	return QuantityUnitsPage(c)
}

func UpdateQuantityUnit(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := c.FormValue("name")
	description := c.FormValue("description")
	namePlural := c.FormValue("name_plural")
	database.DB.Exec("UPDATE quantity_units SET name=?, description=?, name_plural=? WHERE id=?", name, description, namePlural, id)
	return QuantityUnitsPage(c)
}

func DeleteQuantityUnit(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var count int
	database.DB.Get(&count, "SELECT COUNT(*) FROM products WHERE qu_id_purchase = ? OR qu_id_stock = ?", id, id)
	if count > 0 {
		return c.String(http.StatusBadRequest, "Cannot delete — products are using this unit")
	}
	database.DB.Exec("DELETE FROM quantity_units WHERE id = ?", id)
	return c.NoContent(http.StatusOK)
}

// ==================== Stores (Shopping Locations) ====================

func StoresPage(c echo.Context) error {
	var items []models.ShoppingLocation
	database.DB.Select(&items, "SELECT id, name, IFNULL(description,'') as description FROM shopping_locations ORDER BY name")
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.StoresContent(items))
	}
	return render(c, http.StatusOK, templates.StoresPage(items))
}

func StoreNewForm(c echo.Context) error {
	return render(c, http.StatusOK, partials.StoreForm(nil, nil))
}

func StoreEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.ShoppingLocation
	err := database.DB.Get(&item, "SELECT id, name, IFNULL(description,'') as description FROM shopping_locations WHERE id = ?", id)
	if err != nil {
		return c.String(http.StatusNotFound, "Store not found")
	}
	var groups []models.ProductGroup
	database.DB.Select(&groups, "SELECT id, name, IFNULL(description,'') as description FROM product_groups WHERE store_id = ? ORDER BY name", id)
	return render(c, http.StatusOK, partials.StoreForm(&item, groups))
}

func CreateStore(c echo.Context) error {
	name := c.FormValue("name")
	description := c.FormValue("description")
	_, err := database.DB.Exec("INSERT INTO shopping_locations (name, description) VALUES (?, ?)", name, description)
	if err != nil {
		return c.String(http.StatusBadRequest, "Store already exists")
	}
	return StoresPage(c)
}

func UpdateStore(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := c.FormValue("name")
	description := c.FormValue("description")
	database.DB.Exec("UPDATE shopping_locations SET name=?, description=? WHERE id=?", name, description, id)
	return StoresPage(c)
}

func DeleteStore(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var count int
	database.DB.Get(&count, "SELECT COUNT(*) FROM products WHERE shopping_location_id = ?", id)
	if count > 0 {
		return c.String(http.StatusBadRequest, "Cannot delete — products are using this store")
	}
	database.DB.Exec("DELETE FROM shopping_locations WHERE id = ?", id)
	return c.NoContent(http.StatusOK)
}
