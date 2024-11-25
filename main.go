package main

import (
	// "fmt"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/thrashdev/foodsearch/internal/config"
	"github.com/thrashdev/foodsearch/internal/database"
	"github.com/thrashdev/foodsearch/internal/fetcher"
	"github.com/thrashdev/foodsearch/internal/utils"

	// "github.com/thrashdev/foodsearch/internal/fetcher"
	// "github.com/thrashdev/foodsearch/internal/models"
	"net/http"
	// "github.com/thrashdev/foodsearch/internal/models"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	type responseReady struct {
		Status string `json:"status"`
	}

	resp := responseReady{"OK"}
	err := respondWithJSON(w, 200, resp)
	if err != nil {
		log.Println("Server failed on checking readiness")
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	glovoSearchURL := os.Getenv("glovo_url")
	glovoFiltersURL := os.Getenv("glovo_filters_url")
	glovoDishURL := os.Getenv("glovo_dishes_url")
	yandexSearchURL := os.Getenv("yandex_search_url")
	yandexLatitudeStr := os.Getenv("yandex_latitude")
	yandexLongitudeStr := os.Getenv("yandex_longitude")
	yandexRestaurantMenuURL := os.Getenv("yandex_restaurant_menu_url")
	yandexLatitude, err := strconv.ParseFloat(yandexLatitudeStr, 64)
	if err != nil {
		log.Fatalf("Couldn't parse latitude: %v", err)
	}
	yandexLongitude, err := strconv.ParseFloat(yandexLongitudeStr, 64)
	if err != nil {
		log.Fatalf("Couldn't parse latitude: %v", err)
	}
	port := os.Getenv("PORT")
	connection_string := os.Getenv("connection_string")
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connection_string)
	if err != nil {
		log.Fatalf("Couldn't connect to database :%v", err)
	}
	defer conn.Close(ctx)
	db := database.New(conn)
	cfg := &config.Config{
		Glovo: config.GlovoConfig{
			SearchURL:  glovoSearchURL,
			FiltersURL: glovoFiltersURL,
			DishURL:    glovoDishURL,
		},
		Yandex: config.YandexConfig{
			SearchURL:         yandexSearchURL,
			RestaurantMenuURL: yandexRestaurantMenuURL,
			Loc:               config.YandexLocation{Longitude: yandexLongitude, Latitude: yandexLatitude},
		},
		DB:              *db,
		UpdateBatchSize: 5,
	}
	fmt.Println("Started fetching new restaurants")
	// err = fetcher.CreateNewGlovoRestaurants(cfg)
	// err = fetcher.CreateNewDishesForRestaurants(cfg)
	rowsAffected := fetcher.CreateNewYandexRestaurants(cfg)
	fmt.Printf("Created %v restaurants\n", rowsAffected)
	if err != nil {
		log.Fatalf("Error fetching yandex restaurants: %v", err)
	}
	yandexRest, _ := cfg.DB.GetYandexRestaurant(context.Background())
	dishes := fetcher.FetchYandexDishes(cfg, utils.DatabaseYandexRestaurantToModel(yandexRest))
	for _, d := range dishes {
		fmt.Println(d)
	}
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("GET /v1/healthz", handlerReadiness)

	server := http.Server{Handler: serveMux, Addr: ":" + port}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("Couldn't start the server: ", err)
	}

}
