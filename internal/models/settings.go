package models

// FieldVisibility represents a single field's visibility setting.
type FieldVisibility struct {
	ID        int    `db:"id" json:"id"`
	TableName string `db:"table_name" json:"table_name"`
	FieldName string `db:"field_name" json:"field_name"`
	Hidden    int    `db:"hidden" json:"hidden"`
}

// TableInfo describes a user-facing table in the settings UI.
type TableInfo struct {
	Name        string
	DisplayName string
	Description string
}

// FieldInfo describes a single field for the settings UI.
type FieldInfo struct {
	Name        string
	DisplayName string
	Description string
	Hidden      bool
}

// ProductFormField represents a field to render in the product form.
type ProductFormField struct {
	Name        string
	DisplayName string
	Description string
}

// ProductFormData bundles all dropdown data needed by the product form.
type ProductFormData struct {
	Locations     []Location
	ProductGroups []Location // Reusing Location (ID+Name) for simple dropdowns.
	Fields        []ProductFormField
}
