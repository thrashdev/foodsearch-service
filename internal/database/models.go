// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package database

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type DishBinding struct {
	ID          pgtype.UUID
	GlovoDishID pgtype.UUID
}

type GlovoDish struct {
	ID                pgtype.UUID
	Name              string
	Description       string
	Price             pgtype.Numeric
	DiscountedPrice   pgtype.Numeric
	GlovoApiDishID    int32
	GlovoRestaurantID pgtype.UUID
	CreatedAt         pgtype.Timestamp
	UpdatedAt         pgtype.Timestamp
}

type GlovoRestaurant struct {
	ID                pgtype.UUID
	Name              string
	Address           string
	DeliveryFee       pgtype.Numeric
	PhoneNumber       pgtype.Text
	GlovoApiStoreID   int32
	GlovoApiAddressID int32
	GlovoApiSlug      string
	CreatedAt         pgtype.Timestamp
	UpdatedAt         pgtype.Timestamp
}

type RestaurantBinding struct {
	ID                 pgtype.UUID
	GlovoRestaurantID  pgtype.UUID
	YandexRestaurantID pgtype.UUID
}

type YandexDish struct {
	ID                 pgtype.UUID
	Name               string
	Description        pgtype.Text
	Price              pgtype.Numeric
	DiscountedPrice    pgtype.Numeric
	YandexRestaurantID pgtype.UUID
	CreatedAt          pgtype.Timestamp
	UpdatedAt          pgtype.Timestamp
}

type YandexFilter struct {
	ID   pgtype.UUID
	Name string
}

type YandexRestaurant struct {
	ID            pgtype.UUID
	Name          string
	Address       pgtype.Text
	DeliveryFee   pgtype.Numeric
	PhoneNumber   pgtype.Text
	YandexApiSlug string
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
}
