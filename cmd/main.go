package main

import (
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/jollej/go-api/pgm/db"
	"github.com/jollej/go-api/pgm/repository"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Env struct {
	items      repository.ItemModel
	categories repository.CategoryModel
}

func main() {

	os.Setenv("DATABASE_URL", "postgres://postgres:docker@localhost:5432/postgres")
	db := db.Connect()

	env := &Env{
		items:      repository.ItemModel{Conn: db},
		categories: repository.CategoryModel{Conn: db},
	}
	t := &Template{
		templates: template.Must(template.ParseGlob("../pgm/public/views/*.html")),
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Renderer = t

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/", func(c echo.Context) error {
		itemViewData := &repository.ItemViewData{
			Items: nil,
		}

		return c.Render(http.StatusOK, "index", itemViewData)
	})
	e.GET("/items/:id", env.items.GetItem)
	//e.POST("/items", env.categories.ListCategories)
	e.GET("/categories", func(c echo.Context) error {
		categories, err := env.categories.ListCategories()
		if err != nil {
			return err
		}

		categoriesPage := &repository.CategoriesListPage{
			Categories: categories,
		}

		return c.Render(http.StatusOK, "categories-list", categoriesPage)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
