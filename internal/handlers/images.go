package handlers

import (
	"Domicile/internal/database"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const imageDir = "data/images"

// saveUploadedImage saves an uploaded image file and returns the filename.
func saveUploadedImage(c echo.Context, fieldName string, subDir string) (string, error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return "", nil // No file uploaded — not an error.
	}

	// Validate it's an image.
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		return "", fmt.Errorf("invalid image type: %s", ext)
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create directory.
	dir := filepath.Join(imageDir, subDir)
	os.MkdirAll(dir, 0755)

	// Generate unique filename.
	filename := uuid.New().String() + ext
	dstPath := filepath.Join(dir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}

	return filename, nil
}

// deleteImage removes an image file from disk.
func deleteImage(subDir, filename string) {
	if filename == "" {
		return
	}
	os.Remove(filepath.Join(imageDir, subDir, filename))
}

// ServeImage serves an image file from the data/images directory.
func ServeImage(c echo.Context) error {
	subDir := c.Param("entity")
	filename := c.Param("filename")

	// Sanitize to prevent path traversal.
	if strings.Contains(subDir, "..") || strings.Contains(filename, "..") {
		return c.NoContent(http.StatusForbidden)
	}

	path := filepath.Join(imageDir, subDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.NoContent(http.StatusNotFound)
	}
	return c.File(path)
}

// UploadProductImage handles image upload for a product.
func UploadProductImage(c echo.Context) error {
	id := c.Param("id")
	filename, err := saveUploadedImage(c, "picture", "products")
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if filename == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	var oldFile string
	database.DB.Get(&oldFile, "SELECT IFNULL(picture_file_name,'') FROM products WHERE id = ?", id)
	deleteImage("products", oldFile)

	database.DB.Exec("UPDATE products SET picture_file_name = ? WHERE id = ?", filename, id)

	// Redirect back to edit form to show the uploaded image.
	c.Response().Header().Set("HX-Redirect", "/products/"+id+"/edit")
	return c.NoContent(http.StatusOK)
}

// UploadLocationImage handles image upload for a location.
func UploadLocationImage(c echo.Context) error {
	id := c.Param("id")
	filename, err := saveUploadedImage(c, "picture", "locations")
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if filename == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	var oldFile string
	database.DB.Get(&oldFile, "SELECT IFNULL(picture_file_name,'') FROM locations WHERE id = ?", id)
	deleteImage("locations", oldFile)

	database.DB.Exec("UPDATE locations SET picture_file_name = ? WHERE id = ?", filename, id)

	c.Response().Header().Set("HX-Redirect", "/locations/"+id+"/edit")
	return c.NoContent(http.StatusOK)
}

// UploadEquipmentImage handles picture upload for equipment.
func UploadEquipmentImage(c echo.Context) error {
	id := c.Param("id")
	filename, err := saveUploadedImage(c, "picture", "equipment")
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if filename == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	var oldFile string
	database.DB.Get(&oldFile, "SELECT IFNULL(picture_file_name,'') FROM equipment WHERE id = ?", id)
	deleteImage("equipment", oldFile)

	database.DB.Exec("UPDATE equipment SET picture_file_name = ? WHERE id = ?", filename, id)

	c.Response().Header().Set("HX-Redirect", "/inventory/"+id+"/edit")
	return c.NoContent(http.StatusOK)
}

// UploadEquipmentManual handles manual/PDF upload for equipment.
func UploadEquipmentManual(c echo.Context) error {
	id := c.Param("id")

	file, err := c.FormFile("manual")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		return c.String(http.StatusBadRequest, "Invalid file type — use PDF or image")
	}

	src, err := file.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	dir := filepath.Join(imageDir, "equipment_manuals")
	os.MkdirAll(dir, 0755)

	filename := uuid.New().String() + ext
	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer dst.Close()
	io.Copy(dst, src)

	// Delete old manual.
	var oldFile string
	database.DB.Get(&oldFile, "SELECT IFNULL(instruction_manual_file_name,'') FROM equipment WHERE id = ?", id)
	if oldFile != "" {
		os.Remove(filepath.Join(imageDir, "equipment_manuals", oldFile))
	}

	database.DB.Exec("UPDATE equipment SET instruction_manual_file_name = ? WHERE id = ?", filename, id)

	c.Response().Header().Set("HX-Redirect", "/equipment/"+id+"/edit")
	return c.NoContent(http.StatusOK)
}

// saveUploadedFile saves any uploaded file with allowed extensions and returns the filename.
func saveUploadedFile(c echo.Context, fieldName string, subDir string, allowedExts []string) (string, error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return "", nil // No file uploaded.
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := false
	for _, a := range allowedExts {
		if ext == a {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", fmt.Errorf("invalid file type: %s", ext)
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dir := filepath.Join(imageDir, subDir)
	os.MkdirAll(dir, 0755)

	filename := uuid.New().String() + ext
	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return filename, err
}
