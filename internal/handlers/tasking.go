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

// ── helpers ──────────────────────────────────────────────────────────────────

func getUsers() []models.User {
	var users []models.User
	database.DB.Select(&users, "SELECT id, username, first_name, last_name FROM users ORDER BY username")
	return users
}

func getTaskLists(userID int) []models.TaskList {
	var lists []models.TaskList
	database.DB.Select(&lists, `
		SELECT tl.id, tl.name, IFNULL(tl.description,'') as description, tl.user_id,
			COUNT(t.id) as task_count,
			SUM(CASE WHEN t.done=1 THEN 1 ELSE 0 END) as done_count
		FROM task_lists tl
		LEFT JOIN tasks t ON t.task_list_id = tl.id
		WHERE tl.user_id = ?
		GROUP BY tl.id ORDER BY tl.name`, userID)
	return lists
}

func getChoreLists(userID int) []models.ChoreList {
	var lists []models.ChoreList
	database.DB.Select(&lists, `
		SELECT cl.id, cl.name, IFNULL(cl.description,'') as description, cl.user_id,
			COUNT(c.id) as chore_count,
			SUM(CASE WHEN c.done=1 THEN 1 ELSE 0 END) as done_count
		FROM chore_lists cl
		LEFT JOIN chores c ON c.chore_list_id = cl.id
		WHERE cl.user_id = ?
		GROUP BY cl.id ORDER BY cl.name`, userID)
	return lists
}

func getTasksForList(listID int) []models.Task {
	var tasks []models.Task
	database.DB.Select(&tasks, `
		SELECT t.id, t.task_list_id, t.name, IFNULL(t.description,'') as description,
			IFNULL(t.due_date,'') as due_date, t.done,
			IFNULL(t.done_timestamp,'') as done_timestamp,
			t.assigned_to_user_id,
			IFNULL(u.username,'') as assigned_username
		FROM tasks t
		LEFT JOIN users u ON t.assigned_to_user_id = u.id
		WHERE t.task_list_id = ?
		ORDER BY t.done ASC, t.due_date ASC, t.name ASC`, listID)
	return tasks
}

func getChoresForList(listID int) []models.Chore {
	var chores []models.Chore
	database.DB.Select(&chores, `
		SELECT c.id, c.chore_list_id, c.name, IFNULL(c.description,'') as description,
			IFNULL(c.due_date,'') as due_date, c.done,
			IFNULL(c.done_timestamp,'') as done_timestamp,
			c.assigned_to_user_id,
			IFNULL(u.username,'') as assigned_username
		FROM chores c
		LEFT JOIN users u ON c.assigned_to_user_id = u.id
		WHERE c.chore_list_id = ?
		ORDER BY c.done ASC, c.due_date ASC, c.name ASC`, listID)
	return chores
}

// ── Task Lists ────────────────────────────────────────────────────────────────

func TasksPage(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	lists := getTaskLists(userID)
	// Auto-select first list if only one, or use query param.
	activeListID, _ := strconv.Atoi(c.QueryParam("list_id"))
	var tasks []models.Task
	if activeListID > 0 {
		tasks = getTasksForList(activeListID)
	} else if len(lists) > 0 {
		activeListID = lists[0].ID
		tasks = getTasksForList(activeListID)
	}
	users := getUsers()
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.TasksContent(lists, tasks, activeListID, users))
	}
	return render(c, http.StatusOK, templates.TasksPage(lists, tasks, activeListID, users))
}

func CreateTaskList(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	name := strings.TrimSpace(c.FormValue("name"))
	desc := c.FormValue("description")
	if name == "" {
		return c.String(http.StatusBadRequest, "Name required")
	}
	res, _ := database.DB.Exec(
		"INSERT INTO task_lists (name, description, user_id) VALUES (?, ?, ?)", name, desc, userID)
	newID, _ := res.LastInsertId()
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(int(newID))
	return TasksPage(c)
}

func UpdateTaskList(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := strings.TrimSpace(c.FormValue("name"))
	desc := c.FormValue("description")
	database.DB.Exec("UPDATE task_lists SET name=?, description=? WHERE id=?", name, desc, id)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(id)
	return TasksPage(c)
}

func DeleteTaskList(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM tasks WHERE task_list_id=?", id)
	database.DB.Exec("DELETE FROM task_lists WHERE id=?", id)
	return TasksPage(c)
}

func ClearTaskList(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("UPDATE tasks SET done=0, done_timestamp=NULL WHERE task_list_id=?", id)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(id)
	return TasksPage(c)
}

func ClearDoneTaskList(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM tasks WHERE task_list_id=? AND done=1", id)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(id)
	return TasksPage(c)
}

// ── Tasks ─────────────────────────────────────────────────────────────────────

func CreateTask(c echo.Context) error {
	listID, _ := strconv.Atoi(c.FormValue("task_list_id"))
	name := strings.TrimSpace(c.FormValue("name"))
	desc := c.FormValue("description")
	dueDate := c.FormValue("due_date")
	assignedStr := c.FormValue("assigned_to_user_id")
	var assignedID *int
	if assignedStr != "" && assignedStr != "0" {
		aid, _ := strconv.Atoi(assignedStr)
		assignedID = &aid
	}
	database.DB.Exec(
		`INSERT INTO tasks (task_list_id, name, description, due_date, assigned_to_user_id)
		 VALUES (?, ?, NULLIF(?,''), NULLIF(?,''), ?)`,
		listID, name, desc, dueDate, assignedID)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(listID)
	return TasksPage(c)
}

func UpdateTask(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := strings.TrimSpace(c.FormValue("name"))
	desc := c.FormValue("description")
	dueDate := c.FormValue("due_date")
	assignedStr := c.FormValue("assigned_to_user_id")
	var assignedID *int
	if assignedStr != "" && assignedStr != "0" {
		aid, _ := strconv.Atoi(assignedStr)
		assignedID = &aid
	}
	var listID int
	database.DB.Get(&listID, "SELECT task_list_id FROM tasks WHERE id=?", id)
	database.DB.Exec(
		`UPDATE tasks SET name=?, description=NULLIF(?,''), due_date=NULLIF(?,''), assigned_to_user_id=? WHERE id=?`,
		name, desc, dueDate, assignedID, id)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(listID)
	return TasksPage(c)
}

func DeleteTask(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var listID int
	database.DB.Get(&listID, "SELECT task_list_id FROM tasks WHERE id=?", id)
	database.DB.Exec("DELETE FROM tasks WHERE id=?", id)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(listID)
	return TasksPage(c)
}

func ToggleTask(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var done int
	database.DB.Get(&done, "SELECT done FROM tasks WHERE id=?", id)
	if done == 0 {
		database.DB.Exec("UPDATE tasks SET done=1, done_timestamp=? WHERE id=?", time.Now().Format("2006-01-02 15:04:05"), id)
	} else {
		database.DB.Exec("UPDATE tasks SET done=0, done_timestamp=NULL WHERE id=?", id)
	}
	var listID int
	database.DB.Get(&listID, "SELECT task_list_id FROM tasks WHERE id=?", id)
	tasks := getTasksForList(listID)
	return render(c, http.StatusOK, partials.TaskRows(tasks, listID))
}

func TaskEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var task models.Task
	database.DB.Get(&task, `
		SELECT t.id, t.task_list_id, t.name, IFNULL(t.description,'') as description,
			IFNULL(t.due_date,'') as due_date, t.done,
			IFNULL(t.done_timestamp,'') as done_timestamp,
			t.assigned_to_user_id, IFNULL(u.username,'') as assigned_username
		FROM tasks t LEFT JOIN users u ON t.assigned_to_user_id = u.id
		WHERE t.id=?`, id)
	users := getUsers()
	return render(c, http.StatusOK, partials.TaskForm(&task, task.TaskListID, users))
}

// ── Chore Lists ───────────────────────────────────────────────────────────────

func ChoresPage(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	lists := getChoreLists(userID)
	activeListID, _ := strconv.Atoi(c.QueryParam("list_id"))
	var chores []models.Chore
	if activeListID > 0 {
		chores = getChoresForList(activeListID)
	} else if len(lists) > 0 {
		activeListID = lists[0].ID
		chores = getChoresForList(activeListID)
	}
	users := getUsers()
	if isHTMX(c) {
		return render(c, http.StatusOK, templates.ChoresContent(lists, chores, activeListID, users))
	}
	return render(c, http.StatusOK, templates.ChoresPage(lists, chores, activeListID, users))
}

func CreateChoreList(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	name := strings.TrimSpace(c.FormValue("name"))
	desc := c.FormValue("description")
	if name == "" {
		return c.String(http.StatusBadRequest, "Name required")
	}
	res, _ := database.DB.Exec(
		"INSERT INTO chore_lists (name, description, user_id) VALUES (?, ?, ?)", name, desc, userID)
	newID, _ := res.LastInsertId()
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(int(newID))
	return ChoresPage(c)
}

func UpdateChoreList(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := strings.TrimSpace(c.FormValue("name"))
	desc := c.FormValue("description")
	database.DB.Exec("UPDATE chore_lists SET name=?, description=? WHERE id=?", name, desc, id)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(id)
	return ChoresPage(c)
}

func DeleteChoreList(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM chores WHERE chore_list_id=?", id)
	database.DB.Exec("DELETE FROM chore_lists WHERE id=?", id)
	return ChoresPage(c)
}

func ClearChoreList(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("UPDATE chores SET done=0, done_timestamp=NULL WHERE chore_list_id=?", id)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(id)
	return ChoresPage(c)
}

func ClearDoneChoreList(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Exec("DELETE FROM chores WHERE chore_list_id=? AND done=1", id)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(id)
	return ChoresPage(c)
}

// ── Chores ────────────────────────────────────────────────────────────────────

func CreateChore(c echo.Context) error {
	listID, _ := strconv.Atoi(c.FormValue("chore_list_id"))
	name := strings.TrimSpace(c.FormValue("name"))
	desc := c.FormValue("description")
	dueDate := c.FormValue("due_date")
	assignedStr := c.FormValue("assigned_to_user_id")
	var assignedID *int
	if assignedStr != "" && assignedStr != "0" {
		aid, _ := strconv.Atoi(assignedStr)
		assignedID = &aid
	}
	database.DB.Exec(
		`INSERT INTO chores (chore_list_id, name, description, due_date, assigned_to_user_id)
		 VALUES (?, ?, NULLIF(?,''), NULLIF(?,''), ?)`,
		listID, name, desc, dueDate, assignedID)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(listID)
	return ChoresPage(c)
}

func UpdateChore(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	name := strings.TrimSpace(c.FormValue("name"))
	desc := c.FormValue("description")
	dueDate := c.FormValue("due_date")
	assignedStr := c.FormValue("assigned_to_user_id")
	var assignedID *int
	if assignedStr != "" && assignedStr != "0" {
		aid, _ := strconv.Atoi(assignedStr)
		assignedID = &aid
	}
	var listID int
	database.DB.Get(&listID, "SELECT chore_list_id FROM chores WHERE id=?", id)
	database.DB.Exec(
		`UPDATE chores SET name=?, description=NULLIF(?,''), due_date=NULLIF(?,''), assigned_to_user_id=? WHERE id=?`,
		name, desc, dueDate, assignedID, id)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(listID)
	return ChoresPage(c)
}

func DeleteChore(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var listID int
	database.DB.Get(&listID, "SELECT chore_list_id FROM chores WHERE id=?", id)
	database.DB.Exec("DELETE FROM chores WHERE id=?", id)
	c.Request().URL.RawQuery = "list_id=" + strconv.Itoa(listID)
	return ChoresPage(c)
}

func ToggleChore(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var done int
	database.DB.Get(&done, "SELECT done FROM chores WHERE id=?", id)
	if done == 0 {
		database.DB.Exec("UPDATE chores SET done=1, done_timestamp=? WHERE id=?", time.Now().Format("2006-01-02 15:04:05"), id)
	} else {
		database.DB.Exec("UPDATE chores SET done=0, done_timestamp=NULL WHERE id=?", id)
	}
	var listID int
	database.DB.Get(&listID, "SELECT chore_list_id FROM chores WHERE id=?", id)
	chores := getChoresForList(listID)
	return render(c, http.StatusOK, partials.ChoreRows(chores, listID))
}

func ChoreEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var chore models.Chore
	database.DB.Get(&chore, `
		SELECT c.id, c.chore_list_id, c.name, IFNULL(c.description,'') as description,
			IFNULL(c.due_date,'') as due_date, c.done,
			IFNULL(c.done_timestamp,'') as done_timestamp,
			c.assigned_to_user_id, IFNULL(u.username,'') as assigned_username
		FROM chores c LEFT JOIN users u ON c.assigned_to_user_id = u.id
		WHERE c.id=?`, id)
	users := getUsers()
	return render(c, http.StatusOK, partials.ChoreForm(&chore, chore.ChoreListID, users))
}
