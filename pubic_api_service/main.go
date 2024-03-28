package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ListingsResponse struct {
	Result   bool `json:"result"`
	Listings []Listing
}

type Listing struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	ListingType string `json:"listing_type"`
	Price       int    `json:"price"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	User        User   `json:"user"`
}

type ListingCreateResponse struct {
	Result  bool `json:"result"`
	Listing ListingCreate
}

type ListingCreate struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	ListingType string `json:"listing_type"`
	Price       int    `json:"price"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type UserResponse struct {
	Result bool `json:"result"`
	User   User
}

type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// INTERFACE LAYER, FACILITATING COMMUNICATION BETWEEN DIFFERENT COMPONENTS IN THE SYSTEM
func routeRest(router *gin.Engine) {
	router.GET("/public-api/listings", getListingsHandler)
	router.POST("/public-api/listings", createListingHandler)
	router.POST("/public-api/users", createUserHandler)
}

func main() {
	router := gin.Default()

	// set rest route
	routeRest(router)

	port := ":6002"
	log.Printf("Starting public API layer. PORT: %s\n", port)
	router.Run(port)
}

// =========== INTERFACE HANDLER, HANDLING REQUEST RESPONSE API DEPEND INTERFACE ===========

func getListingsHandler(c *gin.Context) {
	pageNum, err := strconv.Atoi(c.DefaultQuery("page_num", "1"))
	if err != nil {
		log.Println("error handler: code error 020, ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page_num param"})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		log.Println("error handler: code error 019, ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page_size param"})
		return
	}

	userID := c.Query("user_id")
	res, err := getListingsUsecase(userID, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": true, "listings": res})
}

func createListingHandler(c *gin.Context) {
	var body Listing
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("error handler: code error 018, ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := createListingUsecase(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"listing": res})
}

func createUserHandler(c *gin.Context) {
	var body User
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("error handler: code error 017, ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := createUserUsecase(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": res})
}

// =========== USECASE LAYER, SERVES AS AN INTERMEDIARY BETWEEN THE PRESENTATION LAYER AND THE DATA LAYER ===========

func getListingsUsecase(userId string, pageNum, pageSize int) ([]Listing, error) {
	res, err := findListingsService(userId, pageNum, pageSize)
	if err != nil {
		return nil, errors.New("api call error: get listings error")
	}

	if !res.Result {
		log.Println("error usecase: code error 016, ", "api result failed: failed to get listings")
		return nil, errors.New("api result failed: failed to get listings")
	}

	var listings []Listing
	for _, val := range res.Listings {
		userRes, err := findUserByIDService(val.UserID)
		if err != nil {
			return nil, errors.New("api call error: get user error")
		}

		if !userRes.Result {
			log.Println("error usecase: code error 016, ", "api result failed: failed to get user")
			return nil, errors.New("api result failed: failed to get user")
		}

		listings = append(listings, Listing{
			ID:          val.ID,
			UserID:      val.UserID,
			ListingType: val.ListingType,
			Price:       val.Price,
			CreatedAt:   val.CreatedAt,
			UpdatedAt:   val.UpdatedAt,
			User: User{
				ID:        userRes.User.ID,
				Name:      userRes.User.Name,
				CreatedAt: userRes.User.CreatedAt,
				UpdatedAt: userRes.User.UpdatedAt,
			},
		})
	}

	return listings, nil
}

func createListingUsecase(listing Listing) (*ListingCreate, error) {
	listingJSON, err := json.Marshal(listing)
	if err != nil {
		log.Println("error usecase: code error 015, ", err)
		return nil, err
	}

	res, err := createListingService(listingJSON)
	if err != nil {
		return nil, errors.New("api call error: create listing error")
	}

	if !res.Result {
		log.Println("error usecase: code error 014, ", "api result failed: failed to create listings")
		return nil, errors.New("api result failed: failed to create listings")
	}

	return &res.Listing, nil
}

func createUserUsecase(user User) (*User, error) {
	userJSON, err := json.Marshal(user)
	if err != nil {
		log.Println("error usecase: code error 013, ", err)
		return nil, err
	}

	res, err := createUserService(userJSON)
	if err != nil {
		return nil, errors.New("api call error: create user error")
	}

	return &res.User, nil
}

// =========== REPOSITORY LAYER, ABSTRACTION OVER THE DATA PERSISTENCE (databases, file systems, or external APIs) ===========

var (
	// listing service api path
	apiPathListingGetList = "http://localhost:6000/listings?page_num=%d&page_size=%d&user_id=%s"
	apiPathListingCreate  = "http://localhost:6000/listings"

	// user service api path
	apiPathUserGetDetail = "http://localhost:6001/users/%d"
	apiPathUserCreate    = "http://localhost:6001/users"
)

func findListingsService(userID string, pageNum, pageSize int) (*ListingsResponse, error) {
	// Call Listing Service to get listings
	resp, err := http.Get(fmt.Sprintf(apiPathListingGetList, pageNum, pageSize, userID))
	if err != nil {
		log.Println("error service: code error 001, ", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("error service: code error 002, ", "error fetching listings from listing service")
		return nil, errors.New("error fetching listings from listing service")
	}

	var listings ListingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listings); err != nil {
		log.Println("error service: code error 003, ", err)
		return nil, err
	}

	return &listings, err
}

func createListingService(listingByte []byte) (*ListingCreateResponse, error) {
	resp, err := http.Post(apiPathListingCreate, "application/json", bytes.NewBuffer(listingByte))
	if err != nil {
		log.Println("error service: code error 004, ", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Println("error service: code error 005, ", "error creating listing from listing service")
		return nil, errors.New("error creating listing from listing service")
	}

	var listing ListingCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		log.Println("error service: code error 006, ", err)
		return nil, err
	}

	return &listing, nil
}

func findUserByIDService(userID int) (*UserResponse, error) {
	// Call User Service to get user
	res, err := http.Get(fmt.Sprintf(apiPathUserGetDetail, userID))
	if err != nil {
		log.Println("error service: code error 007, ", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Println("error service: code error 008, ", "error fetching user from user service")
		return nil, errors.New("error fetching user from user service")
	}

	var user UserResponse
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		log.Println("error service: code error 009, ", err)
		log.Println("error service: ", err)
		return nil, err
	}

	return &user, nil
}

func createUserService(userByte []byte) (*UserResponse, error) {
	resp, err := http.Post(apiPathUserCreate, "application/json", bytes.NewBuffer(userByte))
	if err != nil {
		log.Println("error service: code error 010, ", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Println("error service: code error 011, ", "error creating user from user service")
		return nil, errors.New("error creating user from user service")
	}

	var user UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Println("error service: code error 012, ", err)
		return nil, err
	}

	return &user, nil
}
