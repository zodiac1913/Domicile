-- Domicile Database Schema
-- Based on Grocy's data model (MIT licensed, github.com/grocy/grocy)
-- Consolidated into a single schema for fresh installs

-- ============================================================
-- Core lookup tables
-- ============================================================

CREATE TABLE IF NOT EXISTS locations (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    is_freezer TINYINT NOT NULL DEFAULT 0,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS quantity_units (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    name_plural TEXT,
    plural_forms TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS quantity_unit_conversions (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    from_qu_id INTEGER NOT NULL,
    to_qu_id INTEGER NOT NULL,
    factor REAL NOT NULL,
    product_id INTEGER,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS product_groups (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS shopping_locations (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

-- ============================================================
-- Products
-- ============================================================

CREATE TABLE IF NOT EXISTS products (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    location_id INTEGER NOT NULL,
    product_group_id INTEGER,
    qu_id_purchase INTEGER NOT NULL,
    qu_id_stock INTEGER NOT NULL,
    qu_id_consume INTEGER,
    qu_id_price INTEGER,
    qu_factor_purchase_to_stock REAL NOT NULL DEFAULT 1.0,
    barcode TEXT,
    min_stock_amount INTEGER NOT NULL DEFAULT 0,
    default_best_before_days INTEGER NOT NULL DEFAULT 0,
    default_best_before_days_after_open INTEGER NOT NULL DEFAULT 0,
    default_best_before_days_after_freezing INTEGER NOT NULL DEFAULT 0,
    default_best_before_days_after_thawing INTEGER NOT NULL DEFAULT 0,
    picture_file_name TEXT,
    enable_tare_weight_handling TINYINT NOT NULL DEFAULT 0,
    tare_weight REAL NOT NULL DEFAULT 0.0,
    not_check_stock_fulfillment_for_recipes TINYINT NOT NULL DEFAULT 0,
    parent_product_id INTEGER,
    calories INTEGER,
    cumulate_min_stock_amount_of_sub_products TINYINT NOT NULL DEFAULT 0,
    due_type TINYINT NOT NULL DEFAULT 1,
    quick_consume_amount REAL NOT NULL DEFAULT 1.0,
    quick_open_amount REAL NOT NULL DEFAULT 1.0,
    hide_on_stock_overview TINYINT NOT NULL DEFAULT 0,
    default_stock_label_type INTEGER NOT NULL DEFAULT 0,
    should_not_be_frozen TINYINT NOT NULL DEFAULT 0,
    treat_opened_as_out_of_stock TINYINT NOT NULL DEFAULT 1,
    no_own_stock TINYINT NOT NULL DEFAULT 0,
    default_consume_location_id INTEGER,
    move_on_open TINYINT NOT NULL DEFAULT 0,
    shopping_location_id INTEGER,
    default_purchase_price_type INTEGER,
    active TINYINT NOT NULL DEFAULT 1,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS product_barcodes (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL,
    barcode TEXT NOT NULL UNIQUE,
    qu_id INTEGER,
    amount REAL,
    shopping_location_id INTEGER,
    last_price DECIMAL(15, 2),
    note TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

-- ============================================================
-- Stock tracking
-- ============================================================

CREATE TABLE IF NOT EXISTS stock (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL,
    amount REAL NOT NULL,
    best_before_date DATE,
    purchased_date DATE DEFAULT (datetime('now', 'localtime')),
    stock_id TEXT NOT NULL,
    price DECIMAL(15, 2),
    open TINYINT NOT NULL DEFAULT 0,
    opened_date DATETIME,
    location_id INTEGER,
    shopping_location_id INTEGER,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS stock_log (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL,
    amount REAL NOT NULL,
    best_before_date DATE,
    purchased_date DATE,
    used_date DATE,
    spoiled TINYINT NOT NULL DEFAULT 0,
    stock_id TEXT NOT NULL,
    transaction_type TEXT NOT NULL,
    price DECIMAL(15, 2),
    undone TINYINT NOT NULL DEFAULT 0,
    undone_timestamp DATETIME,
    opened_date DATETIME,
    location_id INTEGER,
    shopping_location_id INTEGER,
    correlation_id TEXT,
    transaction_id TEXT,
    stock_row_id INTEGER,
    recipe_id INTEGER,
    note TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

-- ============================================================
-- Stock views
-- ============================================================

CREATE VIEW IF NOT EXISTS stock_current AS
SELECT
    s.product_id,
    SUM(s.amount) AS amount,
    MIN(s.best_before_date) AS best_before_date,
    SUM(CASE WHEN s.open = 1 THEN s.amount ELSE 0 END) AS amount_opened,
    SUM(s.amount) - SUM(CASE WHEN s.open = 1 THEN s.amount ELSE 0 END) AS amount_unopened,
    IFNULL((SELECT SUM(sc2.amount) FROM stock sc2 WHERE sc2.product_id = s.product_id OR sc2.product_id IN (SELECT id FROM products WHERE parent_product_id = s.product_id)), 0) AS amount_aggregated,
    0 AS is_aggregated_amount
FROM stock s
GROUP BY s.product_id;

CREATE VIEW IF NOT EXISTS stock_missing_products AS
SELECT
    p.id,
    MAX(p.name) AS name,
    p.min_stock_amount - IFNULL(SUM(s.amount), 0) AS amount_missing
FROM products p
LEFT JOIN stock s ON p.id = s.product_id
WHERE p.min_stock_amount != 0
    AND p.active = 1
GROUP BY p.id
HAVING IFNULL(SUM(s.amount), 0) < p.min_stock_amount;

CREATE VIEW IF NOT EXISTS stock_current_locations AS
SELECT
    1 AS id,
    s.product_id,
    SUM(s.amount) AS amount,
    s.location_id AS location_id,
    l.name AS location_name,
    l.is_freezer AS location_is_freezer
FROM stock s
JOIN locations l ON s.location_id = l.id
GROUP BY s.product_id, s.location_id, l.name;

-- ============================================================
-- Shopping
-- ============================================================

CREATE TABLE IF NOT EXISTS shopping_lists (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS shopping_list (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    shopping_list_id INTEGER NOT NULL DEFAULT 1,
    product_id INTEGER,
    note TEXT,
    amount REAL NOT NULL DEFAULT 0,
    amount_autoadded REAL NOT NULL DEFAULT 0,
    done TINYINT NOT NULL DEFAULT 0,
    qu_id INTEGER,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

-- ============================================================
-- Recipes & Meal Planning
-- ============================================================

CREATE TABLE IF NOT EXISTS recipes (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    picture_file_name TEXT,
    base_servings INTEGER NOT NULL DEFAULT 1,
    desired_servings INTEGER NOT NULL DEFAULT 1,
    not_check_shoppinglist TINYINT NOT NULL DEFAULT 0,
    product_id INTEGER,
    type TEXT NOT NULL DEFAULT 'normal',
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS recipes_pos (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    recipe_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    amount REAL NOT NULL DEFAULT 0,
    note TEXT,
    qu_id INTEGER,
    only_check_single_unit_in_stock TINYINT NOT NULL DEFAULT 0,
    not_check_stock_fulfillment TINYINT NOT NULL DEFAULT 0,
    ingredient_group TEXT,
    variable_amount TEXT,
    price_factor REAL NOT NULL DEFAULT 1.0,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS recipes_nestings (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    recipe_id INTEGER NOT NULL,
    includes_recipe_id INTEGER NOT NULL,
    includes_servings INTEGER NOT NULL DEFAULT 1,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS meal_plan (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    day DATE NOT NULL,
    recipe_id INTEGER,
    product_id INTEGER,
    product_amount REAL,
    product_qu_id INTEGER,
    note TEXT,
    type TEXT NOT NULL DEFAULT 'recipe',
    section_id INTEGER,
    done TINYINT NOT NULL DEFAULT 0,
    done_timestamp DATETIME,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS meal_plan_sections (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    sort_number INTEGER,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

-- ============================================================
-- Chores
-- ============================================================

CREATE TABLE IF NOT EXISTS chores (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    period_type TEXT NOT NULL,
    period_days INTEGER,
    period_config TEXT,
    track_date_only TINYINT NOT NULL DEFAULT 0,
    rollover TINYINT NOT NULL DEFAULT 0,
    assignment_type TEXT,
    assignment_config TEXT,
    next_execution_assigned_to_user_id INTEGER,
    consume_product_on_execution TINYINT NOT NULL DEFAULT 0,
    product_id INTEGER,
    product_amount REAL,
    period_interval INTEGER NOT NULL DEFAULT 1,
    active TINYINT NOT NULL DEFAULT 1,
    start_date DATE,
    rescheduled_date DATETIME,
    rescheduled_next_execution_assigned_to_user_id INTEGER,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS chores_log (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    chore_id INTEGER NOT NULL,
    tracked_time DATETIME,
    done_by_user_id INTEGER,
    undone TINYINT NOT NULL DEFAULT 0,
    undone_timestamp DATETIME,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

-- ============================================================
-- Tasks
-- ============================================================

CREATE TABLE IF NOT EXISTS task_categories (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    due_date DATETIME,
    done TINYINT NOT NULL DEFAULT 0,
    done_timestamp DATETIME,
    category_id INTEGER,
    assigned_to_user_id INTEGER,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);


-- ============================================================
-- Batteries
-- ============================================================

CREATE TABLE IF NOT EXISTS batteries (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    used_in TEXT,
    charge_interval_days INTEGER NOT NULL DEFAULT 0,
    active TINYINT NOT NULL DEFAULT 1,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS battery_charge_cycles (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    battery_id INTEGER NOT NULL,
    tracked_time DATETIME,
    undone TINYINT NOT NULL DEFAULT 0,
    undone_timestamp DATETIME,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

-- ============================================================
-- Equipment
-- ============================================================

CREATE TABLE IF NOT EXISTS equipment (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    instruction_manual_file_name TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

-- ============================================================
-- Users & Auth
-- ============================================================

CREATE TABLE IF NOT EXISTS users (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    first_name TEXT,
    last_name TEXT,
    password TEXT NOT NULL,
    picture_file_name TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    session_key TEXT NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    expires DATETIME,
    last_used DATETIME,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS api_keys (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    api_key TEXT NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    expires DATETIME,
    last_used DATETIME,
    key_type TEXT NOT NULL DEFAULT 'default',
    description TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS user_settings (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    key TEXT NOT NULL,
    value TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime')),
    row_updated_timestamp DATETIME DEFAULT (datetime('now', 'localtime')),
    UNIQUE(user_id, key)
);

-- ============================================================
-- User-defined fields & entities
-- ============================================================

CREATE TABLE IF NOT EXISTS userfields (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    entity TEXT NOT NULL,
    name TEXT NOT NULL,
    caption TEXT NOT NULL,
    type TEXT NOT NULL,
    show_as_column_in_tables TINYINT NOT NULL DEFAULT 0,
    sort_number INTEGER,
    config TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime')),
    UNIQUE(entity, name)
);

CREATE TABLE IF NOT EXISTS userfield_values (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    field_id INTEGER NOT NULL,
    object_id INTEGER NOT NULL,
    value TEXT NOT NULL,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime')),
    UNIQUE(field_id, object_id)
);

CREATE TABLE IF NOT EXISTS userentities (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    caption TEXT NOT NULL,
    description TEXT,
    show_in_sidebar_menu TINYINT NOT NULL DEFAULT 1,
    icon_css_class TEXT,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS userobjects (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    userentity_id INTEGER NOT NULL,
    row_created_timestamp DATETIME DEFAULT (datetime('now', 'localtime'))
);

-- ============================================================
-- Seed data
-- ============================================================

INSERT OR IGNORE INTO locations (name, description) VALUES ('Fridge', 'Main refrigerator');
INSERT OR IGNORE INTO locations (name, description, is_freezer) VALUES ('Freezer', 'Main freezer', 1);
INSERT OR IGNORE INTO locations (name, description) VALUES ('Pantry', 'Dry goods storage');
INSERT OR IGNORE INTO locations (name, description) VALUES ('Cabinet', 'Kitchen cabinet');

INSERT OR IGNORE INTO quantity_units (name, description, name_plural) VALUES ('Piece', 'Single item', 'Pieces');
INSERT OR IGNORE INTO quantity_units (name, description, name_plural) VALUES ('Pack', 'Package', 'Packs');
INSERT OR IGNORE INTO quantity_units (name, description, name_plural) VALUES ('Bottle', 'Bottle', 'Bottles');
INSERT OR IGNORE INTO quantity_units (name, description, name_plural) VALUES ('Can', 'Can', 'Cans');
INSERT OR IGNORE INTO quantity_units (name, description, name_plural) VALUES ('lb', 'Pound', 'lbs');
INSERT OR IGNORE INTO quantity_units (name, description, name_plural) VALUES ('oz', 'Ounce', 'oz');
INSERT OR IGNORE INTO quantity_units (name, description, name_plural) VALUES ('gal', 'Gallon', 'gal');

INSERT OR IGNORE INTO shopping_lists (name) VALUES ('Default');
