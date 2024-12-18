package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thrashdev/foodsearch/internal/config"
	"github.com/thrashdev/foodsearch/internal/database"
	"github.com/thrashdev/foodsearch/internal/models"
	"github.com/thrashdev/foodsearch/internal/utils"
)

type dishResponse struct {
	Payload      []byte
	RestaurantID uuid.UUID
}

func InitGlovo(cfg *config.ServiceConfig) {
	startupCommands := []startupCommand{
		CreateNewGlovoRestaurants,
		CreateNewDishesForGlovoRestaurants,
	}
	for _, cmd := range startupCommands {
		res, err := cmd(cfg)
		if err != nil {
			cfg.Logger.DPanic(err)
		}
		res.print()
	}
}

func restaurantDifferenceGlovo(restaurants []models.GlovoRestaurant, dbrNames []string) []models.GlovoRestaurant {
	mb := make(map[string]struct{}, len(dbrNames))
	for _, name := range dbrNames {
		mb[name] = struct{}{}
	}
	var diff []models.GlovoRestaurant
	for _, rest := range restaurants {
		if _, found := mb[rest.Name]; !found {
			diff = append(diff, rest)
		}
	}
	return diff

}

func dishNameDifference(dishes []models.GlovoDish, dbDishNames []string) []models.GlovoDish {
	mb := make(map[string]struct{}, len(dbDishNames))
	for _, name := range dbDishNames {
		mb[name] = struct{}{}
	}
	var diff []models.GlovoDish
	for _, dish := range dishes {
		if _, found := mb[dish.Name]; !found {
			diff = append(diff, dish)
		}
	}
	return diff
}

func dishGlovoApiIdDifference(dishes []models.GlovoDish, ids []int32) []models.GlovoDish {
	mb := make(map[int32]struct{}, len(ids))
	for _, id := range ids {
		mb[id] = struct{}{}
	}
	var diff []models.GlovoDish
	for _, dish := range dishes {
		if _, found := mb[int32(dish.GlovoAPIDishID)]; !found {
			diff = append(diff, dish)
		}
	}
	return diff

}

func findMaxDiscount(promos []glovoPromotion) float64 {
	max := 0.0
	for _, promo := range promos {
		if promo.Price > max {
			max = promo.Price
		}
	}
	// fmt.Printf("Returning %v\n", max)
	return max
}

func removeDuplicateRestaurants(rests []models.GlovoRestaurant) []models.GlovoRestaurant {
	extracted := make(map[string]struct{})
	result := []models.GlovoRestaurant{}
	for _, r := range rests {
		_, ok := extracted[r.Name]
		if ok {
			continue
		}
		result = append(result, r)
		extracted[r.Name] = struct{}{}
	}
	return result
}

func removeDuplicateDishes(dishesD []models.GlovoDish) []models.GlovoDish {
	extracted := make(map[int]struct{})
	result := []models.GlovoDish{}
	for _, d := range dishesD {
		_, ok := extracted[d.GlovoAPIDishID]
		if ok {
			continue
		}
		result = append(result, d)
		extracted[d.GlovoAPIDishID] = struct{}{}
	}
	return result
}

func fetchByGlovoUrl(url string) (payload []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("Couldn't create request, err: %v", err)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Glovo-API-Version", "14")
	req.Header.Set("Glovo-App-Platform", "web")
	req.Header.Set("Glovo-App-Type", "customer")
	req.Header.Set("Glovo-App-Version", "7")
	req.Header.Set("Glovo-Location-City-Code", "BSK")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("Couldn't post request, err: %v", err)
	}
	defer resp.Body.Close()
	fmt.Println("Response Code: ", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Couldn't read response body, err: %v", err)
	}

	return respBody, nil
}

func fetchGlovoFilters(filtersURL string) (filters []string, err error) {
	payload, err := fetchByGlovoUrl(filtersURL)
	if err != nil {
		return []string{}, err
	}
	var glovoFilters glovoFiltersResponse
	err = json.Unmarshal(payload, &glovoFilters)
	if err != nil {
		return []string{}, fmt.Errorf("Couldn't unmarshal json, err: %v", err)
	}

	for _, item := range glovoFilters.TopFilters {
		filters = append(filters, item.FilterName)
	}

	return filters, nil
}

func fetchGlovoRestaurantsByFilter(baseURL string, filter string) (restaurants []models.GlovoRestaurant, err error) {
	fullURL := baseURL + "&filter=" + filter
	respBody, err := fetchByGlovoUrl(fullURL)

	var glovoResp glovoRestaurantsResponse
	err = json.Unmarshal(respBody, &glovoResp)
	if err != nil {
		log.Println(fullURL)
		return []models.GlovoRestaurant{}, err
	}

	for _, item := range glovoResp.Elements {
		glovoRestaurant := models.GlovoRestaurant{
			Restaurant: models.Restaurant{
				ID:          uuid.New(),
				Name:        item.SingleData.StoreData.Store.Name,
				Address:     &item.SingleData.StoreData.Store.Address,
				DeliveryFee: &item.SingleData.StoreData.Store.DeliveryFeeInfo.Fee,
				PhoneNumber: &item.SingleData.StoreData.Store.PhoneNumber,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			GlovoApiStoreID:   item.SingleData.StoreData.Store.ID,
			GlovoApiAddressID: item.SingleData.StoreData.Store.AddressID,
			GlovoApiSlug:      item.SingleData.StoreData.Store.Slug,
		}

		restaurants = append(restaurants, glovoRestaurant)
	}

	return restaurants, nil

}

// TODO: implement proper error-handling with an error channel
func CreateNewDishesForGlovoRestaurants(cfg *config.ServiceConfig) (DBActionResult, error) {
	fmt.Println("Creating new dishes")
	ctx := context.Background()
	maxConcurrency := 2
	limiter := make(chan struct{}, maxConcurrency)
	errCh := make(chan error)
	go utils.PrintErrors(errCh)
	dbRestaurants, err := cfg.DB.GetAllGlovoRestaurants(ctx)
	if err != nil {
		wrapped := fmt.Errorf("Couldn't get glovo restaurants: %w", err)
		return DBActionResult{}, wrapped
	}
	payloadsCh := make(chan dishResponse, len(dbRestaurants))
	for _, dbRest := range dbRestaurants {
		limiter <- struct{}{}
		go func() {
			defer func() { <-limiter }()
			// defer fmt.Printf("Fetched dishes for %v\n", dbRest.Name)
			rest := utils.GlovoRestDBtoModel(dbRest)
			payload := FetchGlovoDishes(rest, cfg.Glovo.DishURL, errCh)
			payloadsCh <- payload
		}()
	}

	// waiting for goroutines to finish
	for i := 0; i < cap(limiter); i++ {
		limiter <- struct{}{}
	}
	close(payloadsCh)

	fmt.Println("Collected all responses")
	dishesD := []models.GlovoDish{}
	for p := range payloadsCh {
		dishesPerRestaurant, err := serializeGlovoDishes(p.Payload, p.RestaurantID)
		if err != nil {
			log.Printf("Error when serializing glovo dishes: %v", err)
			continue
		}
		dishesD = append(dishesD, dishesPerRestaurant...)
	}
	fmt.Println("Prepped all dishes")
	fmt.Printf("Total dishes fetched: %v\n", len(dishesD))
	dishes := removeDuplicateDishes(dishesD)
	dbApiDish_IDS, err := cfg.DB.GetGlovoDishAPI_ID(ctx)
	if err != nil {
		wrapped := fmt.Errorf("Couldn't get glovo dish api IDs: %w", err)
		return DBActionResult{}, wrapped
	}
	dishesToAdd := dishGlovoApiIdDifference(dishes, dbApiDish_IDS)
	fmt.Printf("New dishes: %v", len(dishesToAdd))
	totalDishesCreated := createNewDishesForGlovoRestaurant(cfg, dishesToAdd, errCh)

	result := DBActionResult{}
	result.records = []DBActionResultRecord{
		makeDBActionResultRecord("Updated %v restaurants", int64(len(dbRestaurants))),
		makeDBActionResultRecord("Created %v total dishes", int64(totalDishesCreated)),
	}
	return result, nil
}

// TODO: implement proper error-handling with an error channel
func createNewDishesForGlovoRestaurant(cfg *config.ServiceConfig, dishes []models.GlovoDish, errCh chan error) (dishesCreated int) {
	ctx := context.Background()
	args := []database.BatchCreateGlovoDishesParams{}
	for _, dish := range dishes {
		arg := utils.GlovoDishModelToDB(dish)
		args = append(args, arg)
	}

	rowsAffected, err := cfg.DB.BatchCreateGlovoDishes(ctx, args)
	if err != nil {
		errCh <- fmt.Errorf("Couldn't create dishes: %v", err)
		return 0
	}
	log.Printf("Created %v dishes", rowsAffected)
	return int(rowsAffected)

}

func CreateNewGlovoRestaurants(cfg *config.ServiceConfig) (DBActionResult, error) {
	newRestaurants, err := fetchGlovoRestaurants(cfg.Glovo.SearchURL, cfg.Glovo.FiltersURL)
	if err != nil {
		wrapped := fmt.Errorf("Error while fetching glovo restaurants: %w", err)
		return DBActionResult{}, wrapped
	}
	log.Printf("Fetched %v restaurants from API\n", len(newRestaurants))
	ctx := context.Background()
	dbRestaurantNames, err := cfg.DB.GetGlovoRestaurantNames(ctx)
	if err != nil {
		wrapped := fmt.Errorf("Error while fetching glovo restaurant names: %w", err)
		return DBActionResult{}, wrapped
	}
	log.Printf("Fetched %v restaurants from DB\n", len(dbRestaurantNames))
	restaurantsToAdd := restaurantDifferenceGlovo(newRestaurants, dbRestaurantNames)
	log.Printf("Restaurants to add: %v\n", len(restaurantsToAdd))
	args := []database.BatchCreateGlovoRestaurantsParams{}
	for _, rest := range restaurantsToAdd {
		arg := utils.GlovoRestModelToDB(rest)
		args = append(args, arg)
	}
	log.Printf("Prepared to create %v restaurants\n", len(args))
	rowsAffected, err := cfg.DB.BatchCreateGlovoRestaurants(ctx, args)
	if err != nil {
		wrapped := fmt.Errorf("Error while posting glovo restaurants to DB: %w", err)
		return DBActionResult{}, wrapped
	}

	result := DBActionResult{}
	result.records = append(result.records, makeDBActionResultRecord("Created %v glovo restaurants", rowsAffected))
	return result, nil
}

func fetchGlovoRestaurants(searchURL string, filtersURL string) (allRestaurants []models.GlovoRestaurant, err error) {
	filters, err := fetchGlovoFilters(filtersURL)
	if err != nil {
		return []models.GlovoRestaurant{}, fmt.Errorf("Couldn't get filters, err: %v", err)
	}

	for _, f := range filters {
		restaurantsByFilter, err := fetchGlovoRestaurantsByFilter(searchURL, url.QueryEscape(f))
		if err != nil {
			return []models.GlovoRestaurant{}, fmt.Errorf("Couldn't fetch by filter: %s. Error :%v", f, err)
		}

		allRestaurants = append(allRestaurants, restaurantsByFilter...)
	}
	result := removeDuplicateRestaurants(allRestaurants)

	return result, nil
}

func FetchGlovoDishes(rest models.GlovoRestaurant, dishURL string, errCh chan error) (payload dishResponse) {
	targetURL := strings.Replace(dishURL, "{glovo_store_id}", strconv.Itoa(rest.GlovoApiStoreID), 1)
	targetURL = strings.Replace(targetURL, "{glovo_address_id}", strconv.Itoa(rest.GlovoApiAddressID), 1)
	responsePayload, err := fetchByGlovoUrl(targetURL)
	if err != nil {
		errCh <- fmt.Errorf("Error encountered while fetching dishes for %v: %v\n", rest.Name, err)
		return dishResponse{}
	}
	dishResp := dishResponse{Payload: responsePayload, RestaurantID: rest.ID}

	return dishResp

}

func serializeGlovoDishes(responsePayload []byte, restID uuid.UUID) ([]models.GlovoDish, error) {
	var dishesResponse glovoDishesResponse
	err := json.Unmarshal(responsePayload, &dishesResponse)
	if err != nil {
		return []models.GlovoDish{}, fmt.Errorf("Error encountered while fetching glovo dishes: %v\n", err)
	}

	dishes := []models.GlovoDish{}
	for _, elem := range dishesResponse.Data.Body {
		if strings.ToLower(elem.Data.Title) == "напитки" {
			continue
		}
		for _, dishItem := range elem.Data.Elements {
			discount := findMaxDiscount(dishItem.Data.Promotions)
			dishes = append(dishes, models.GlovoDish{
				GlovoAPIDishID: int(dishItem.Data.ID),
				Dish: models.Dish{
					ID:              uuid.New(),
					Name:            dishItem.Data.Name,
					Description:     dishItem.Data.Description,
					Price:           dishItem.Data.Price,
					DiscountedPrice: discount,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				},
				GlovoRestaurantID: restID,
			})

		}
	}

	return dishes, nil
}
