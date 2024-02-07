package repository

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CategoryModel struct {
	Conn *pgx.Conn
}

type CategoriesListPage struct {
	Categories []Category
}

type Category struct {
	Id          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	CategoryId  string    `json:"CategoryId"`
}

func (i CategoryModel) ListCategories() ([]Category, error) {

	var categories []Category

	rows, err := i.Conn.Query(context.Background(), "select id, description, category from categories")
	if err != nil {
		log.Fatal("Select query failed", err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category Category
		rows.Scan(&category.Id, &category.Description, &category.CategoryId)
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
