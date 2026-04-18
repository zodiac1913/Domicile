package handlers

import (
	"Domicile/internal/database"
	"Domicile/internal/models"
	"Domicile/templates"
	"Domicile/templates/partials"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func getBatteries() []models.Battery {
	var items []models.Battery
	database.DB.Select(&items, `
		SELECT id, name, IFNULL(description,'') as description,
			IFNULL(used_in,'') as used_in, charge_interval_days, active
		FROM batteries WHERE active=1 ORDER BY name`)
	return items
}

func getBatteryOverview() []models.BatteryOverview {
	var items []models.BatteryOverview
	database.DB.Select(&items, `
		SELECT b.id, b.name, IFNULL(b.description,'') as description,
			IFNULL(b.used_in,'') as used_in, b.charge_interval_days, b.active,
			IFNULL(MAX(bc.tracked_time),'') as last_charged,
			CASE WHEN MAX(bc.tracked_time) IS NOT NULL
				THEN CAST(JULIANDAY('now','localtime') - JULIANDAY(MAX(bc.tracked_time)) AS INTEGER)
				ELSE -1
			END as days_ago,
			CASE WHEN b.charge_interval_days > 0 AND MAX(bc.tracked_time) IS NOT NULL
				THEN DATE(MAX(bc.tracked_time), '+' || b.charge_interval_days || ' days')
				ELSE ''
			END as next_due,
			CASE WHEN b.charge_interval_days > 0 AND MAX(bc.tracked_time) IS NOT NULL
				THEN CAST(JULIANDAY(DATE(MAX(bc.tracked_time), '+' || b.charge_interval_days || ' days')) - JULIANDAY('now','localtime') AS INTEGER)
				ELSE 9999
			END as days_until
		FROM batteries b
		LEFT JOIN battery_charge_cycles bc ON bc.battery_id = b.id AND bc.undone = 0
		WHERE b.active = 1
		GROUP BY b.id
		ORDER BY days_until ASC, b.name ASC`)
	return items
}

func BatteriesPage(c echo.Context) error {
	items := getBatteries()
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.BatteriesContent(items))
	}
	return render(c, http.StatusOK, templates.BatteriesPage(items))
}

func BatteriesOverview(c echo.Context) error {
	items := getBatteryOverview()
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.BatteriesOverviewContent(items))
	}
	return render(c, http.StatusOK, templates.BatteriesOverviewPage(items))
}

func BatteryNewForm(c echo.Context) error {
	return render(c, http.StatusOK, partials.BatteryForm(nil))
}

func BatteryEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var b models.Battery
	err := database.DB.Get(&b, `SELECT id, name, IFNULL(description,'') as description,
		IFNULL(used_in,'') as used_in, charge_interval_days, active FROM batteries WHERE id=?`, id)
	if err != nil {
		return c.String(http.StatusNotFound, "Battery not found")
	}
	return render(c, http.StatusOK, partials.BatteryForm(&b))
}

func CreateBattery(c echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	desc := c.FormValue("description")
	usedIn := c.FormValue("used_in")
	interval, _ := strconv.Atoi(c.FormValue("charge_interval_days"))
	database.DB.Exec(`INSERT INTO batteries (name, description, used_in, charge_interval_days)
		VALUES (?, NULLIF(?,''), NULLIF(?,''), ?)`, name, desc, usedIn, interval)
	return BatteriesPage(c)
}

func UpdateBattery(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := strings.TrimSpace(c.FormValue("name"))
	desc := c.FormValue("description")
	usedIn := c.FormValue("used_in")
	interval, _ := strconv.Atoi(c.FormValue("charge_interval_days"))
	database.DB.Exec(`UPDATE batteries SET name=?, description=NULLIF(?,''), used_in=NULLIF(?,''), charge_interval_days=? WHERE id=?`,
		name, desc, usedIn, interval, id)
	return BatteriesPage(c)
}

func DeleteBattery(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM battery_charge_cycles WHERE battery_id=?", id)
	database.DB.Exec("DELETE FROM batteries WHERE id=?", id)
	return BatteriesPage(c)
}

func LogCharge(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	trackedTime := c.FormValue("tracked_time")
	if trackedTime == "" {
		trackedTime = time.Now().Format("2006-01-02 15:04:05")
	}
	database.DB.Exec(`INSERT INTO battery_charge_cycles (battery_id, tracked_time) VALUES (?, ?)`, id, trackedTime)
	return BatteriesOverview(c)
}

func UndoCharge(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec(`UPDATE battery_charge_cycles SET undone=1, undone_timestamp=datetime('now','localtime')
		WHERE id = (SELECT id FROM battery_charge_cycles WHERE battery_id=? AND undone=0 ORDER BY tracked_time DESC LIMIT 1)`, id)
	return BatteriesOverview(c)
}

func BatteryHistory(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var b models.Battery
	database.DB.Get(&b, `SELECT id, name, IFNULL(description,'') as description,
		IFNULL(used_in,'') as used_in, charge_interval_days, active FROM batteries WHERE id=?`, id)
	type Cycle struct {
		ID          int    `db:"id"`
		TrackedTime string `db:"tracked_time"`
		Undone      int    `db:"undone"`
	}
	var cycles []Cycle
	database.DB.Select(&cycles, `SELECT id, IFNULL(tracked_time,'') as tracked_time, undone
		FROM battery_charge_cycles WHERE battery_id=? ORDER BY tracked_time DESC LIMIT 50`, id)
	// Convert to the exported type for the partial.
	var out []models.BatteryChargeCycle
	for _, cy := range cycles {
		out = append(out, models.BatteryChargeCycle{
			ID:          cy.ID,
			TrackedTime: cy.TrackedTime,
			Undone:      cy.Undone,
		})
	}
	return render(c, http.StatusOK, partials.BatteryHistoryPanel(b, out))
}
