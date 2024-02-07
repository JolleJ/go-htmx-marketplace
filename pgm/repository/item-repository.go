package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

type ItemModel struct {
	Conn *pgx.Conn
}

type ItemViewData struct {
	Items []Item
}

type Item struct {
	Id             uuid.UUID      `form:"id" json:"id"`
	Name           string         `form:"name" json:"name"`
	Description    sql.NullString `form:"description" json:"description"`
	Price          float64        `json:"price"`
	CurrentBid     float64        `json:"bid"`
	CurrentBidder  float64        `json:"bidder"`
	BiddingEndDate time.Time      `json:"biddingEndDate"`
	Category       string         `json:"category"`
}

type ItemBidData struct {
	Id     uuid.UUID `form:"id" json:"id"`
	Bid    float64   `json:"bid"`
	Bidder float64   `json:"bidder"`
}

func (i ItemModel) BidOnItem(c echo.Context) error {

	var itemBidDate ItemBidData
	err := c.Bind(&itemBidDate)
	if err != nil {
		log.Fatal("Could not bind request data")
		return c.JSON(http.StatusBadRequest, c.Request().Body)
	}

	err = i.Conn.QueryRow(context.Background(), "update items set current_bid = $1, current_bidder = $2 where id = $3").Scan("", "", "")
	if err != nil {
		log.Fatal("Error updating bidding data")
	}
	return c.JSON(http.StatusCreated, "")
}

func (i ItemModel) GetItem(c echo.Context) error {

	var item Item
	var uuid string

	err := (&echo.DefaultBinder{}).BindPathParams(c, &uuid)
	if err != nil {
		log.Fatal("Could not bind parameters", err.Error())
	}

	fmt.Println(uuid)
	err = i.Conn.QueryRow(context.Background(), "select id, item_name, description from items where id = $1", &uuid).Scan(&item.Id, &item.Name, &item.Description)
	if err != nil {
		log.Fatal("Could not fetch item", err.Error())
	}

	return c.JSON(http.StatusOK, item)
}

func (i ItemModel) CreateItem(c echo.Context) error {

	var item Item
	err := c.Bind(&item)
	if err != nil {
		log.Fatal("Could not bind item")
		return c.JSON(http.StatusBadRequest, err)
	}

	_, err = i.Conn.Exec(context.Background(), "insert into items (item_name, description, category) values ($1, $2, $3)", &item.Name, &item.Description, &item.Category)
	if err != nil {
		log.Fatal("Could not insert item")
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, item)
}

func (i ItemModel) InitializeItems(c echo.Context) error {

	now := time.Now()
	var items []Item

	for i := 0; i < 10000; i++ {
		item := Item{Name: fmt.Sprint("Item ", i), Price: 100, CurrentBid: 0, BiddingEndDate: now.Add(10 * 24 * time.Hour)}
		items = append(items, item)
	}

	for _, item := range items {
		_, err := i.Conn.Exec(context.Background(), "insert into items (item_name, description) values ($1, $2)", &item.Name, &item.Description)
		if err != nil {
			log.Fatal("Could not fetch item", err.Error())
		}
	}

	return c.JSON(http.StatusOK, "item")
}

func (i ItemModel) ListItems() ([]Item, error) {

	var items []Item

	rows, err := i.Conn.Query(context.Background(), "select id, item_name, description from items")
	if err != nil {
		log.Fatal("Select query failed", err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		rows.Scan(&item.Id, &item.Name, &item.Description)
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
