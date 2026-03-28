package models

import (
	"strconv"
	"time"
)

type Product struct {
	ID             int       `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Location       string    `json:"location" db:"location"`
	Quantity       int       `json:"quantity" db:"quantity"`
	MinStockAmount int       `json:"min_stock_amount" db:"min_stock_amount"`
	ExpiresAt      time.Time `json:"expires_at,omitempty" db:"-"`
	ExpiresAtStr   string    `json:"-" db:"expires_at"`
}

func (p Product) IDStr() string {
	return strconv.Itoa(p.ID)
}

func (p Product) QuantityStr() string {
	return strconv.Itoa(p.Quantity)
}

func (p Product) ExpiresDisplay() string {
	// Try parsing from the DB string first
	if p.ExpiresAtStr != "" {
		t, err := time.Parse("2006-01-02", p.ExpiresAtStr)
		if err == nil {
			return t.Format("Jan 2, 2006")
		}
	}
	if !p.ExpiresAt.IsZero() {
		return p.ExpiresAt.Format("Jan 2, 2006")
	}
	return "—"
}

func (p Product) IsExpiringSoon() bool {
	var t time.Time
	if p.ExpiresAtStr != "" {
		t, _ = time.Parse("2006-01-02", p.ExpiresAtStr)
	} else {
		t = p.ExpiresAt
	}
	if t.IsZero() {
		return false
	}
	return time.Until(t) < 7*24*time.Hour && time.Until(t) > 0
}

type ShoppingItem struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Quantity int    `json:"quantity" db:"quantity"`
	Done     bool   `json:"done" db:"done"`
}

func (s ShoppingItem) IDStr() string {
	return strconv.Itoa(s.ID)
}

func (s ShoppingItem) QuantityStr() string {
	return strconv.Itoa(s.Quantity)
}

type DashboardStats struct {
	ProductCount  int
	ExpiringSoon  int
	BelowMinStock int
	ShoppingItems int
}

func (d DashboardStats) ProductCountStr() string {
	return strconv.Itoa(d.ProductCount)
}

func (d DashboardStats) ExpiringSoonStr() string {
	return strconv.Itoa(d.ExpiringSoon)
}

func (d DashboardStats) BelowMinStockStr() string {
	return strconv.Itoa(d.BelowMinStock)
}

func (d DashboardStats) ShoppingItemsStr() string {
	return strconv.Itoa(d.ShoppingItems)
}
