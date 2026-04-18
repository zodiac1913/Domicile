package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"Domicile/templates/partials"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// AdminMiddleware ensures only admins can access admin routes.
func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, _ := c.Get("user_id").(int)
		var isAdmin int
		database.DB.Get(&isAdmin, "SELECT is_admin FROM users WHERE id = ?", userID)
		if isAdmin != 1 {
			return c.String(http.StatusForbidden, "Admin access required")
		}
		return next(c)
	}
}

// AdminPage renders the admin panel.
func AdminPage(c echo.Context) error {
	var users []models.User
	database.DB.Select(&users, `
		SELECT id, username, IFNULL(first_name,'') as first_name,
			IFNULL(last_name,'') as last_name, is_admin,
			IFNULL(picture_file_name,'') as picture_file_name
		FROM users ORDER BY username`)
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.AdminContent(users))
	}
	return render(c, http.StatusOK, templates.AdminPage(users))
}

// AdminUserNewForm renders the create user form.
func AdminUserNewForm(c echo.Context) error {
	return render(c, http.StatusOK, partials.AdminUserForm(nil))
}

// AdminUserEditForm renders the edit user form.
func AdminUserEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var user models.User
	err := database.DB.Get(&user, `
		SELECT id, username, IFNULL(first_name,'') as first_name,
			IFNULL(last_name,'') as last_name, is_admin,
			IFNULL(picture_file_name,'') as picture_file_name
		FROM users WHERE id = ?`, id)
	if err != nil {
		return c.String(http.StatusNotFound, "User not found")
	}
	return render(c, http.StatusOK, partials.AdminUserForm(&user))
}

// AdminCreateUser handles creating a new user.
func AdminCreateUser(c echo.Context) error {
	username := c.FormValue("username")
	firstName := c.FormValue("first_name")
	lastName := c.FormValue("last_name")
	password := c.FormValue("password")
	isAdmin := checkboxVal(c, "is_admin")

	if username == "" || password == "" {
		return c.String(http.StatusBadRequest, "Username and password are required")
	}
	if len(password) < 6 {
		return c.String(http.StatusBadRequest, "Password must be at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to hash password")
	}

	result, err := database.DB.Exec(
		"INSERT INTO users (username, first_name, last_name, password, is_admin) VALUES (?, ?, ?, ?, ?)",
		username, firstName, lastName, string(hash), isAdmin)
	if err != nil {
		return c.String(http.StatusBadRequest, "Username already exists")
	}

	// Handle picture upload.
	userID, _ := result.LastInsertId()
	if pic, err := saveUploadedImage(c, "picture", "users"); err == nil && pic != "" {
		database.DB.Exec("UPDATE users SET picture_file_name = ? WHERE id = ?", pic, userID)
	}

	return AdminPage(c)
}

// AdminUpdateUser handles updating a user.
func AdminUpdateUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	username := c.FormValue("username")
	firstName := c.FormValue("first_name")
	lastName := c.FormValue("last_name")
	password := c.FormValue("password")
	isAdmin := checkboxVal(c, "is_admin")

	database.DB.Exec("UPDATE users SET username=?, first_name=?, last_name=?, is_admin=? WHERE id=?",
		username, firstName, lastName, isAdmin, id)

	// Update password only if provided.
	if password != "" {
		if len(password) < 6 {
			return c.String(http.StatusBadRequest, "Password must be at least 6 characters")
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		database.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hash), id)
	}

	// Handle picture upload.
	if pic, err := saveUploadedImage(c, "picture", "users"); err == nil && pic != "" {
		var oldPic string
		database.DB.Get(&oldPic, "SELECT IFNULL(picture_file_name,'') FROM users WHERE id = ?", id)
		deleteImage("users", oldPic)
		database.DB.Exec("UPDATE users SET picture_file_name = ? WHERE id = ?", pic, id)
	}

	return AdminPage(c)
}

// AdminDeleteUser handles deleting a user.
func AdminDeleteUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	currentUserID, _ := c.Get("user_id").(int)

	if id == currentUserID {
		return c.String(http.StatusBadRequest, "You cannot delete yourself")
	}

	// Delete user picture.
	var picFile string
	database.DB.Get(&picFile, "SELECT IFNULL(picture_file_name,'') FROM users WHERE id = ?", id)
	if picFile != "" {
		deleteImage("users", picFile)
	}

	// Delete sessions.
	database.DB.Exec("DELETE FROM sessions WHERE user_id = ?", id)
	database.DB.Exec("DELETE FROM users WHERE id = ?", id)
	return c.NoContent(http.StatusOK)
}
