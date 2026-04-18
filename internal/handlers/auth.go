package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"Domicile/templates/partials"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const sessionCookieName = "domicile_session"
const sessionDuration = 30 * 24 * time.Hour // 30 days

// CurrentUserID returns the logged-in user's ID from the session, or 0.
func CurrentUserID(c echo.Context) int {
	cookie, err := c.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return 0
	}
	var userID int
	err = database.DB.Get(&userID, `
		SELECT user_id FROM sessions
		WHERE session_key = ? AND (expires IS NULL OR expires > datetime('now'))`,
		cookie.Value)
	if err != nil {
		return 0
	}
	// Update last_used.
	database.DB.Exec("UPDATE sessions SET last_used = datetime('now') WHERE session_key = ?", cookie.Value)
	return userID
}

// HasUsers returns true if any users exist in the database.
func HasUsers() bool {
	var count int
	database.DB.Get(&count, "SELECT COUNT(*) FROM users")
	return count > 0
}

// AuthMiddleware checks if the user is logged in. Redirects to login or setup.
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().URL.Path
		// Allow auth routes and static files without login.
		if path == "/login" || path == "/setup" || path == "/logout" ||
			strings.HasPrefix(path, "/static/") ||
			strings.HasPrefix(path, "/sml/") ||
			strings.HasPrefix(path, "/images/") {
			return next(c)
		}

		if !HasUsers() {
			return c.Redirect(http.StatusFound, "/setup")
		}

		userID := CurrentUserID(c)
		if userID == 0 {
			return c.Redirect(http.StatusFound, "/login")
		}

		c.Set("user_id", userID)

		// Load user info for navbar.
		var user models.User
		database.DB.Get(&user, `
			SELECT id, username, IFNULL(first_name,'') as first_name,
				IFNULL(last_name,'') as last_name, is_admin,
				IFNULL(picture_file_name,'') as picture_file_name
			FROM users WHERE id = ?`, userID)
		c.Set("current_user", user)

		return next(c)
	}
}

// SetupPage renders the first-run admin setup screen.
func SetupPage(c echo.Context) error {
	if HasUsers() {
		return c.Redirect(http.StatusFound, "/login")
	}
	return render(c, http.StatusOK, templates.SetupPage())
}

// SetupCreate handles the first admin account creation.
func SetupCreate(c echo.Context) error {
	if HasUsers() {
		return c.Redirect(http.StatusFound, "/login")
	}

	username := c.FormValue("username")
	password := c.FormValue("password")
	confirm := c.FormValue("confirm_password")

	if username == "" || password == "" {
		return c.String(http.StatusBadRequest, "Username and password are required")
	}
	if password != confirm {
		return c.String(http.StatusBadRequest, "Passwords do not match")
	}
	if len(password) < 6 {
		return c.String(http.StatusBadRequest, "Password must be at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to hash password")
	}

	result, err := database.DB.Exec(
		"INSERT INTO users (username, password, is_admin) VALUES (?, ?, 1)",
		username, string(hash))
	if err != nil {
		return c.String(http.StatusBadRequest, "Username already exists")
	}

	userID, _ := result.LastInsertId()

	// Assign all existing data to this user.
	for _, table := range []string{"product_groups", "shopping_locations", "locations", "equipment",
		"shopping_list", "stock", "stock_log", "field_visibility",
		"chores", "chores_log", "tasks", "batteries", "battery_charge_cycles", "meal_plan"} {
		database.DB.Exec("UPDATE "+table+" SET user_id = ? WHERE user_id = 1 OR user_id IS NULL", userID)
	}

	// Create session and log in.
	createSession(c, int(userID))
	return c.Redirect(http.StatusFound, "/")
}

// LoginPage renders the login screen.
func LoginPage(c echo.Context) error {
	if !HasUsers() {
		return c.Redirect(http.StatusFound, "/setup")
	}
	// Already logged in?
	if CurrentUserID(c) > 0 {
		return c.Redirect(http.StatusFound, "/")
	}
	return render(c, http.StatusOK, templates.LoginPage())
}

// LoginSubmit handles login form submission.
func LoginSubmit(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	var user struct {
		ID       int    `db:"id"`
		Password string `db:"password"`
	}
	err := database.DB.Get(&user, "SELECT id, password FROM users WHERE username = ?", username)
	if err != nil {
		return render(c, http.StatusOK, templates.LoginPageWithError("Invalid username or password"))
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return render(c, http.StatusOK, templates.LoginPageWithError("Invalid username or password"))
	}

	createSession(c, user.ID)
	return c.Redirect(http.StatusFound, "/")
}

// Logout destroys the session.
func Logout(c echo.Context) error {
	cookie, err := c.Cookie(sessionCookieName)
	if err == nil {
		database.DB.Exec("DELETE FROM sessions WHERE session_key = ?", cookie.Value)
	}
	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	return c.Redirect(http.StatusFound, "/login")
}

func createSession(c echo.Context, userID int) {
	key := generateSessionKey()
	expires := time.Now().Add(sessionDuration)

	database.DB.Exec(
		"INSERT INTO sessions (session_key, user_id, expires, last_used) VALUES (?, ?, ?, datetime('now'))",
		key, userID, expires)

	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    key,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func generateSessionKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetUserSetting returns a user setting value, or the default if not set.
func GetUserSetting(userID int, key string, defaultVal string) string {
	var val string
	err := database.DB.Get(&val, "SELECT value FROM user_settings WHERE user_id = ? AND key = ?", userID, key)
	if err != nil {
		return defaultVal
	}
	return val
}

// SetUserSetting saves a user setting.
func SetUserSetting(userID int, key string, value string) {
	database.DB.Exec(`
		INSERT INTO user_settings (user_id, key, value)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, key)
		DO UPDATE SET value = ?, row_updated_timestamp = datetime('now', 'localtime')`,
		userID, key, value, value)
}

// NavbarUser returns the user avatar HTML for the navbar.
func NavbarUser(c echo.Context) error {
	user, ok := c.Get("current_user").(models.User)
	if !ok {
		return c.NoContent(http.StatusOK)
	}
	return render(c, http.StatusOK, partials.NavbarUserBubble(user))
}
