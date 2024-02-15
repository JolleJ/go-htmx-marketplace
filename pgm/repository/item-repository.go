package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	CurrentBidder  string         `json:"bidder"`
	BiddingEndDate time.Time      `json:"biddingEndDate"`
	Category       string         `json:"category"`
}

type ItemBidData struct {
	Id     uuid.UUID `json:"id"`
	Bid    float64   `json:"bid"`
	Bidder string    `json:"bidder"`
}

func (i ItemModel) BidOnItem(c echo.Context) error {

	var itemBidData ItemBidData
	err := c.Bind(&itemBidData)
	if err != nil {
		log.Fatal("Could not bind request data")
		return c.JSON(http.StatusBadRequest, c.Request().Body)
	}

	uuid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Fatal("Could not bind parameters", err.Error())
	}

	if err != nil {
		log.Fatal("Could not bind parameters", err.Error())
	}

	_, err = i.Conn.Exec(context.Background(), "update items set current_bid = $1, current_bidder = $2 where id = $3", itemBidData.Bid, itemBidData.Bidder, uuid)
	if err != nil {
		log.Fatal("Error updating bidding data: ", err)
	}

	itemBidData.Id = uuid
	return c.JSON(http.StatusOK, itemBidData)
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

// Returns all items
func (i ItemModel) ListItems() ([]Item, error) {

	var items []Item

	rows, err := i.Conn.Query(context.Background(), "select id, item_name, description, current_bid, current_bidder from items")
	if err != nil {
		log.Fatal("Select query failed", err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		rows.Scan(&item.Id, &item.Name, &item.Description, &item.CurrentBid, &item.CurrentBidder)
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// Returns all items given the pagination settings
// input: page int
// input: numberOfItems int, How many items to return
func (i ItemModel) ListItemsPagination(c echo.Context) error {

	var acceptedParams = map[string]string{
		"page":          "Current request page",
		"numberOfItems": "How many items the page should include in the output",
	}

	for param := range c.QueryParams() {
		if _, ok := acceptedParams[param]; !ok {
			return c.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid query parameter: %s", param))
		}
	}

	// Retrieve query parameters
	page := c.QueryParam("page")
	numberOfItems := c.QueryParam("numberOfItems")

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "page has to be a number")
	}

	numberOfItemsInt, err := strconv.Atoi(numberOfItems)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "numberOfItems has to be a number")
	}

	var items []Item
	offset := (pageInt - 1) * numberOfItemsInt
	rows, err := i.Conn.Query(context.Background(), fmt.Sprintf("select id, item_name, description, current_bid, current_bidder from items LIMIT %d OFFSET %d", numberOfItemsInt, offset))
	if err != nil {
		log.Fatal("Select query failed", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		rows.Scan(&item.Id, &item.Name, &item.Description, &item.CurrentBid, &item.CurrentBidder)
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, items)
}
