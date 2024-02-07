package services

import (
	"net/http"

	"github.com/jollej/go-api/pgm/repository"
	"github.com/labstack/echo/v4"
)

func CreateItem(c echo.Context) error {

	var item repository.Item
	err := c.Bind(&item)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, item)
}
