package main

import (
	"Domicile/internal/database"
	"Domicile/internal/handlers"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize database
	if err := database.Init("data"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Static files — registered before auth middleware so they're always accessible
	e.Static("/static", "static")
	e.Static("/sml", "static/js/sml")

	// Auth routes (no auth needed)
	e.GET("/login", handlers.LoginPage)
	e.POST("/login", handlers.LoginSubmit)
	e.GET("/setup", handlers.SetupPage)
	e.POST("/setup", handlers.SetupCreate)
	e.GET("/logout", handlers.Logout)

	// Auth middleware for everything else
	e.Use(handlers.AuthMiddleware)

	// Admin routes
	admin := e.Group("/admin", handlers.AdminMiddleware)
	admin.GET("", handlers.AdminPage)
	admin.GET("/users/new", handlers.AdminUserNewForm)
	admin.POST("/users", handlers.AdminCreateUser)
	admin.GET("/users/:id/edit", handlers.AdminUserEditForm)
	admin.POST("/users/:id/update", handlers.AdminUpdateUser)
	admin.DELETE("/users/:id", handlers.AdminDeleteUser)

	// Image serving
	e.GET("/images/:entity/:filename", handlers.ServeImage)

	// Pages
	e.GET("/", handlers.HomePage)
	e.GET("/products", handlers.ProductsPage)

	// Product CRUDL
	e.GET("/products/new", handlers.ProductNewForm)
	e.POST("/products", handlers.CreateProduct)
	e.GET("/products/:id", handlers.ProductView)
	e.GET("/products/:id/edit", handlers.ProductEditForm)
	e.PUT("/products/:id", handlers.UpdateProduct)
	e.POST("/products/:id/update", handlers.UpdateProduct)
	e.DELETE("/products/:id", handlers.DeleteProduct)
	e.GET("/products/:id/merge", handlers.MergeProductForm)
	e.POST("/products/:id/merge", handlers.MergeProduct)
	e.POST("/products/:id/consume", handlers.ConsumeProduct)
	e.POST("/products/:id/add", handlers.AddProduct)
	e.POST("/products/:id/add-to-cart", handlers.AddToCart)

	// Locations CRUDL
	e.GET("/locations", handlers.LocationsPage)
	e.GET("/locations/new", handlers.LocationNewForm)
	e.POST("/locations", handlers.CreateLocation)
	e.GET("/locations/:id/edit", handlers.LocationEditForm)
	e.PUT("/locations/:id", handlers.UpdateLocation)
	e.POST("/locations/:id/update", handlers.UpdateLocation)
	e.DELETE("/locations/:id", handlers.DeleteLocation)

	// Shopping List CRUDL
	e.GET("/shopping", handlers.ShoppingPage)
	e.GET("/shopping/new", handlers.ShoppingNewForm)
	e.POST("/shopping", handlers.CreateShoppingItem)
	e.POST("/shopping/add-continue", handlers.CreateShoppingItemContinue)
	e.GET("/shopping/:id/edit", handlers.ShoppingEditForm)
	e.PUT("/shopping/:id", handlers.UpdateShoppingItem)
	e.DELETE("/shopping/:id", handlers.DeleteShoppingItem)
	e.POST("/shopping/:id/toggle", handlers.ToggleShoppingDone)
	e.POST("/shopping/clear-done", handlers.ClearDoneItems)
	e.POST("/shopping/clear-list", handlers.ClearList)
	e.POST("/shopping/add-to-stock", handlers.AddDoneToStock)
	e.POST("/shopping/add-missing", handlers.AddMissingStock)
	e.GET("/shopping/shop", handlers.ShoppingShopMode)
	e.POST("/shopping/:id/shop-toggle", handlers.ShopModeToggle)
	e.GET("/shopping/print", handlers.ShoppingPrint)
	e.POST("/shopping/:id/assign-group", handlers.AssignProductGroup)
	e.GET("/api/product-groups", handlers.ProductGroupsJSON)
	e.GET("/api/navbar-user", handlers.NavbarUser)

	// Settings
	e.GET("/settings", handlers.SettingsPage)
	e.GET("/settings/field-visibility", handlers.SettingsFieldVisibility)
	e.GET("/settings/fields", handlers.SettingsFieldList)
	e.POST("/settings/toggle-field", handlers.ToggleFieldVisibility)
	e.GET("/settings/backup", handlers.BackupSettingsPage)
	e.GET("/settings/shopping", handlers.ShoppingSettingsPage)
	e.POST("/settings/user-setting", handlers.SaveUserSetting)
	e.GET("/settings/backup/download", handlers.BackupDownload)
	e.POST("/settings/backup/restore", handlers.BackupRestore)

	// Equipment CRUDL
	e.GET("/inventory", handlers.EquipmentPage)
	e.GET("/inventory/new", handlers.EquipmentNewForm)
	e.POST("/inventory", handlers.CreateEquipment)
	e.GET("/inventory/:id/edit", handlers.EquipmentEditForm)
	e.PUT("/inventory/:id", handlers.UpdateEquipment)
	e.POST("/inventory/:id/update", handlers.UpdateEquipment)
	e.DELETE("/inventory/:id", handlers.DeleteEquipment)
	e.POST("/inventory/bulk-upload", handlers.BulkUploadEquipment)

	// Product Groups CRUDL
	e.GET("/product-groups", handlers.ProductGroupsPage)

	// Stock
	e.GET("/stock", handlers.StockOverview)

	// Recipes CRUDL
	e.GET("/recipes", handlers.RecipesPage)
	e.GET("/recipes/new", handlers.RecipeNewForm)
	e.POST("/recipes", handlers.CreateRecipe)
	e.POST("/recipes/scrape", handlers.ScrapeRecipe)
	e.GET("/recipes/:id", handlers.RecipeView)
	e.GET("/recipes/:id/edit", handlers.RecipeEditForm)
	e.PUT("/recipes/:id", handlers.UpdateRecipe)
	e.POST("/recipes/:id/update", handlers.UpdateRecipe)
	e.DELETE("/recipes/:id", handlers.DeleteRecipe)
	e.GET("/recipes/:id/image", handlers.RecipeImage)
	e.POST("/recipes/:id/image", handlers.RecipeImageUpload)
	e.POST("/recipes/:id/fetch-image", handlers.RecipeFetchImage)

	// Meal Plan
	e.GET("/meal-plan", handlers.MealPlanPage)
	e.GET("/meal-plan/week", handlers.MealPlanWeek)
	e.GET("/meal-plan/entries/new", handlers.MealPlanEntryForm)
	e.POST("/meal-plan/entries", handlers.CreateMealPlanEntry)
	e.GET("/meal-plan/entries/:id/edit", handlers.MealPlanEntryEditForm)
	e.POST("/meal-plan/entries/:id/update", handlers.UpdateMealPlanEntry)
	e.DELETE("/meal-plan/entries/:id", handlers.DeleteMealPlanEntry)
	e.POST("/meal-plan/entries/:id/toggle", handlers.ToggleMealPlanDone)
	e.GET("/meal-plan/ingredients", handlers.MealPlanIngredientsForm)
	e.POST("/meal-plan/ingredients", handlers.MealPlanExportIngredients)
	e.POST("/meal-plan/ingredients/add-to-shopping", handlers.MealPlanAddToShopping)
	e.GET("/meal-plan/slots", handlers.MealPlanSlotsPage)
	e.POST("/meal-plan/slots", handlers.CreateMealSlot)
	e.POST("/meal-plan/slots/:id/update", handlers.UpdateMealSlot)
	e.DELETE("/meal-plan/slots/:id", handlers.DeleteMealSlot)

	// Batteries
	e.GET("/batteries", handlers.BatteriesPage)
	e.GET("/batteries/new", handlers.BatteryNewForm)
	e.POST("/batteries", handlers.CreateBattery)
	e.GET("/batteries/overview", handlers.BatteriesOverview)
	e.GET("/batteries/:id/edit", handlers.BatteryEditForm)
	e.POST("/batteries/:id/update", handlers.UpdateBattery)
	e.DELETE("/batteries/:id", handlers.DeleteBattery)
	e.POST("/batteries/:id/charge", handlers.LogCharge)
	e.POST("/batteries/:id/undo", handlers.UndoCharge)
	e.GET("/batteries/:id/history", handlers.BatteryHistory)

	// Tasks
	e.GET("/tasks", handlers.TasksPage)
	e.POST("/tasks/lists", handlers.CreateTaskList)
	e.POST("/tasks/lists/:id/update", handlers.UpdateTaskList)
	e.DELETE("/tasks/lists/:id", handlers.DeleteTaskList)
	e.POST("/tasks/lists/:id/reset", handlers.ClearTaskList)
	e.POST("/tasks/lists/:id/clear-done", handlers.ClearDoneTaskList)
	e.POST("/tasks", handlers.CreateTask)
	e.GET("/tasks/:id/edit", handlers.TaskEditForm)
	e.POST("/tasks/:id/update", handlers.UpdateTask)
	e.DELETE("/tasks/:id", handlers.DeleteTask)
	e.POST("/tasks/:id/toggle", handlers.ToggleTask)

	// Chores
	e.GET("/chores", handlers.ChoresPage)
	e.POST("/chores/lists", handlers.CreateChoreList)
	e.POST("/chores/lists/:id/update", handlers.UpdateChoreList)
	e.DELETE("/chores/lists/:id", handlers.DeleteChoreList)
	e.POST("/chores/lists/:id/reset", handlers.ClearChoreList)
	e.POST("/chores/lists/:id/clear-done", handlers.ClearDoneChoreList)
	e.POST("/chores", handlers.CreateChore)
	e.GET("/chores/:id/edit", handlers.ChoreEditForm)
	e.POST("/chores/:id/update", handlers.UpdateChore)
	e.DELETE("/chores/:id", handlers.DeleteChore)
	e.POST("/chores/:id/toggle", handlers.ToggleChore)

	// Stock continued
	e.GET("/stock/purchase", handlers.PurchaseForm)
	e.POST("/stock/purchase", handlers.CreatePurchase)
	e.GET("/stock/consume", handlers.ConsumeForm)
	e.POST("/stock/consume", handlers.CreateConsume)
	e.GET("/stock/transfer", handlers.TransferForm)
	e.POST("/stock/transfer", handlers.CreateTransfer)

	// Product Groups CRUDL continued
	e.GET("/product-groups/new", handlers.ProductGroupNewForm)
	e.POST("/product-groups", handlers.CreateProductGroup)
	e.GET("/product-groups/:id/edit", handlers.ProductGroupEditForm)
	e.PUT("/product-groups/:id", handlers.UpdateProductGroup)
	e.DELETE("/product-groups/:id", handlers.DeleteProductGroup)

	// Quantity Units CRUDL
	e.GET("/quantity-units", handlers.QuantityUnitsPage)
	e.GET("/quantity-units/new", handlers.QuantityUnitNewForm)
	e.POST("/quantity-units", handlers.CreateQuantityUnit)
	e.GET("/quantity-units/:id/edit", handlers.QuantityUnitEditForm)
	e.PUT("/quantity-units/:id", handlers.UpdateQuantityUnit)
	e.DELETE("/quantity-units/:id", handlers.DeleteQuantityUnit)

	// Stores CRUDL
	e.GET("/stores", handlers.StoresPage)
	e.GET("/stores/new", handlers.StoreNewForm)
	e.POST("/stores", handlers.CreateStore)
	e.GET("/stores/:id/edit", handlers.StoreEditForm)
	e.PUT("/stores/:id", handlers.UpdateStore)
	e.DELETE("/stores/:id", handlers.DeleteStore)

	// Image uploads
	e.POST("/products/:id/image", handlers.UploadProductImage)
	e.POST("/locations/:id/image", handlers.UploadLocationImage)
	e.POST("/inventory/:id/image", handlers.UploadEquipmentImage)
	e.POST("/inventory/:id/manual", handlers.UploadEquipmentManual)

	// Dashboard partials (htmx)
	e.GET("/partials/expiring-soon", handlers.ExpiringSoonPartial)
	e.GET("/partials/shopping-preview", handlers.ShoppingPreviewPartial)

	// Find available port starting at 1323
	port := 1323
	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", &port)
	} else {
		port = findAvailablePort(1323, 1423)
	}
	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost:%d", port)

	log.Printf("Domicile running at %s", url)
	if os.Getenv("DOMICILE_NO_BROWSER") == "" {
		openBrowser(url)
	}

	e.Logger.Fatal(e.Start(addr))
}

func findAvailablePort(start, max int) int {
	for port := start; port <= max; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port
		}
	}
	log.Fatalf("No available port found in range %d-%d", start, max)
	return 0
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		log.Printf("Open %s in your browser", url)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Could not open browser: %v — open %s manually", err, url)
	}
}
