package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitalis-virtus/ecommerce-go/db"
	"github.com/vitalis-virtus/ecommerce-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewApplication(prodCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("Product id is empty")

			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := c.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")

			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)

		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		err = db.AddProductToCart(ctx, app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(http.StatusOK, "successfully added to cart")
	}
}

func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("Product id is empty")

			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := c.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")

			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)

		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		err = db.RemoveItemFromCart(ctx, app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(http.StatusOK, "successfully removed item from cart")

	}
}

func GetItemFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")
		if userID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid id"})
			c.Abort()
			return
		}

		usertID, _ := primitive.ObjectIDFromHex(userID)

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var filledCart models.User
		err := UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: usertID}}).Decode(&filledCart)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, "not found")
		}

		filterMatch := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: usertID}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.E{Key: "$sum", Value: "$usercart.price"}}}}}
		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filterMatch, unwind, grouping})
		if err != nil {
			log.Println(err)
		}

		var listing []bson.M
		if err = pointCursor.All(ctx, &listing); err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		for _, json := range listing {
			c.IndentedJSON(http.StatusOK, json["total"])
			c.IndentedJSON(http.StatusOK, filledCart.User_Cart)
		}

		ctx.Done()
	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.Query("id")

		if userQueryID == "" {
			log.Panicln("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		err := db.BuyItemFromCart(ctx, app.userCollection, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(http.StatusOK, "Successfully placed the order")
	}
}

func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("Product id is empty")

			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}

		userQueryID := c.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")

			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)

		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		err = db.InstantBuyer(ctx, app.prodCollection, app.userCollection, productID, userQueryID)

		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		c.IndentedJSON(http.StatusOK, "successfully placed the order")
	}
}
