package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitalis-virtus/ecommerce-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")
		if userID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "userId is empty"})
			c.Abort()
			return
		}

		address, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Internal server error")
		}

		var addresses models.Address

		addresses.Address_ID = primitive.NewObjectID()

		if err = c.BindJSON(&addresses); err != nil {
			c.IndentedJSON(http.StatusNotAcceptable, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// we create agregation query
		matchFilter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: address}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$address"}}}}
		group := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$address_id"}, {Key: "count", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}}}}}

		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{matchFilter, unwind, group})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "internal server error")
		}

		var addressInfo []bson.M
		if err = pointCursor.All(ctx, &addressInfo); err != nil {
			panic(err)
		}

		var size int32

		for _, addressNo := range addressInfo {

			count := addressNo["count"]
			size = count.(int32)
		}
		if size < 2 {
			filter := bson.D{primitive.E{Key: "_id", Value: address}}
			update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
			_, err := UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			c.IndentedJSON(http.StatusBadRequest, "not allowed")
		}
		defer cancel()
		ctx.Done()
	}
}

func EditAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")

		if user_id == "" {
			c.Header("Content-type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid user_id"})
			c.Abort()
			return
		}

		usert_id, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Internal server error")
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var editaddress models.Address
		if err := c.BindJSON(&editaddress); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}
		filter := bson.D{primitive.E{Key: "_id", Value: usert_id}}
		// home address is with a "0" index
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.0.house_name", Value: editaddress.House}, {Key: "address.0.street_name", Value: editaddress.Street}, {Key: "address.0.city_name", Value: editaddress.City}, {Key: "address.0.pin_code", Value: editaddress.Pincode}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "something went wrong")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(http.StatusOK, "successfully updated the home address")
	}
}

func EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")

		if user_id == "" {
			c.Header("Content-type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{
				"error": "invalid user_id",
			})
			c.Abort()
			return
		}

		usert_id, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "internal server error")
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var editaddress models.Address
		if err := c.BindJSON(&editaddress); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
		}

		filter := bson.D{primitive.E{Key: "_id", Value: usert_id}}

		// work address is with a "1" index
		update := bson.D{primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "address.1.house_name", Value: editaddress.House}, {Key: "address.1.street_name", Value: editaddress.Street}, {Key: "address.1.city_name", Value: editaddress.City}, {Key: "address.1.pin_code", Value: editaddress.Pincode},
		}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "something wrong")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(http.StatusOK, "successfully updated the work address")

	}
}

func DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")

		if user_id == "" {
			c.Header("Content-type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"erorr": "invalid search index"})
			c.Abort()
			return
		}

		addresses := make([]models.Address, 0)

		usert_id, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Internal server error")
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		filter := bson.D{primitive.E{Key: "_id", Value: usert_id}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, "Wrong command")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(http.StatusOK, "Successfully deleted")
	}
}
