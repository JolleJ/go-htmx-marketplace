package main

import (
	"net/http"
	"os"

	"github.com/jollej/go-api/pgm/db"
	"github.com/jollej/go-api/pgm/repository"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

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
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/", func(c echo.Context) error {
		itemViewData := &repository.ItemViewData{
			Items: nil,
		}

		return c.Render(http.StatusOK, "index", itemViewData)
	})

	// /items handles GET requests to /items endpoint
	// It accepts the following optinoal query parameters:
	// Both query parameters must be included in order for pagination to work.
	// If none is provided, all items will be returned (Not prefered)
	//   - page: Current request page
	//   - numberOfItems: How many items the page should include in the output
	e.GET("/items", env.items.ListItemsPagination)
	e.GET("/items/:id", env.items.GetItem)
	e.POST("/items/:id/bid", env.items.BidOnItem)

	e.Logger.Fatal(e.Start(":1323"))

}
