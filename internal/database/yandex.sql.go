// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: yandex.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type BatchCreateYandexDishesParams struct {
	ID                 pgtype.UUID
	Name               string
	Description        pgtype.Text
	Price              pgtype.Numeric
	DiscountedPrice    pgtype.Numeric
	YandexRestaurantID pgtype.UUID
	YandexApiID        int32
	CreatedAt          pgtype.Timestamp
	UpdatedAt          pgtype.Timestamp
}

type BatchCreateYandexRestaurantsParams struct {
	ID            pgtype.UUID
	Name          string
	Address       pgtype.Text
	DeliveryFee   pgtype.Numeric
	PhoneNumber   pgtype.Text
	YandexApiSlug string
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
}

const getAllYandexDishes = `-- name: GetAllYandexDishes :many
select id, name, description, price, discounted_price, yandex_restaurant_id, yandex_api_id, created_at, updated_at from yandex_dish
`

func (q *Queries) GetAllYandexDishes(ctx context.Context) ([]YandexDish, error) {
	rows, err := q.db.Query(ctx, getAllYandexDishes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []YandexDish
	for rows.Next() {
		var i YandexDish
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Price,
			&i.DiscountedPrice,
			&i.YandexRestaurantID,
			&i.YandexApiID,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllYandexRestaurants = `-- name: GetAllYandexRestaurants :many
select id, name, address, delivery_fee, phone_number, yandex_api_slug, created_at, updated_at from yandex_restaurant
`

func (q *Queries) GetAllYandexRestaurants(ctx context.Context) ([]YandexRestaurant, error) {
	rows, err := q.db.Query(ctx, getAllYandexRestaurants)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []YandexRestaurant
	for rows.Next() {
		var i YandexRestaurant
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Address,
			&i.DeliveryFee,
			&i.PhoneNumber,
			&i.YandexApiSlug,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getYandexDishApiIDS = `-- name: GetYandexDishApiIDS :many
select yandex_api_id from yandex_dish
`

func (q *Queries) GetYandexDishApiIDS(ctx context.Context) ([]int32, error) {
	rows, err := q.db.Query(ctx, getYandexDishApiIDS)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var yandex_api_id int32
		if err := rows.Scan(&yandex_api_id); err != nil {
			return nil, err
		}
		items = append(items, yandex_api_id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getYandexDishesForRestaurant = `-- name: GetYandexDishesForRestaurant :many
select id, name, description, price, discounted_price, yandex_restaurant_id, yandex_api_id, created_at, updated_at from yandex_dish
where yandex_restaurant_id = $1
`

func (q *Queries) GetYandexDishesForRestaurant(ctx context.Context, yandexRestaurantID pgtype.UUID) ([]YandexDish, error) {
	rows, err := q.db.Query(ctx, getYandexDishesForRestaurant, yandexRestaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []YandexDish
	for rows.Next() {
		var i YandexDish
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Price,
			&i.DiscountedPrice,
			&i.YandexRestaurantID,
			&i.YandexApiID,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getYandexFilters = `-- name: GetYandexFilters :many
SELECT name FROM yandex_filters
`

func (q *Queries) GetYandexFilters(ctx context.Context) ([]string, error) {
	rows, err := q.db.Query(ctx, getYandexFilters)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		items = append(items, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getYandexRestaurant = `-- name: GetYandexRestaurant :one
select id, name, address, delivery_fee, phone_number, yandex_api_slug, created_at, updated_at from yandex_restaurant limit 1
`

func (q *Queries) GetYandexRestaurant(ctx context.Context) (YandexRestaurant, error) {
	row := q.db.QueryRow(ctx, getYandexRestaurant)
	var i YandexRestaurant
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Address,
		&i.DeliveryFee,
		&i.PhoneNumber,
		&i.YandexApiSlug,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getYandexRestaurantSlugs = `-- name: GetYandexRestaurantSlugs :many
SELECT yandex_api_slug FROM yandex_restaurant
`

func (q *Queries) GetYandexRestaurantSlugs(ctx context.Context) ([]string, error) {
	rows, err := q.db.Query(ctx, getYandexRestaurantSlugs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var yandex_api_slug string
		if err := rows.Scan(&yandex_api_slug); err != nil {
			return nil, err
		}
		items = append(items, yandex_api_slug)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
