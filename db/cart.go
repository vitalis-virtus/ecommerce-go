package db

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/vitalis-virtus/ecommerce-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindProduct    = errors.New("can't find the product")
	ErrCantDecodeProducts = errors.New("can't find the product")
	ErrUserIdIsInvalid    = errors.New("this user is not valid")
	ErrCantUpdateUser     = errors.New("can't add this product to the cart")
	ErrCantRemoveItemCart = errors.New("can't remove this item from the cart")
	ErrCantGetItem        = errors.New("was unable to get the item from the cart")
	ErrCantBuyCartItem    = errors.New("can't update the purchase")
)

func AddProductToCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	searchFromdb, err := prodCollection.Find(ctx, bson.M{"_id": productID})
	if err != nil {
		log.Println(err)
		return ErrCantFindProduct
	}

	var productCart []models.ProductUser
	err = searchFromdb.All(ctx, &productCart)
	if err != nil {
		log.Println(err)
		return ErrCantDecodeProducts
	}

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsInvalid
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "usercart", Value: bson.D{{Key: "$each", Value: productCart}}}}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
		return ErrCantUpdateUser
	}
	return nil
}

func RemoveItemFromCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsInvalid
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.M{"$pull": bson.M{"usercart": bson.M{"_id": (productID)}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return ErrCantRemoveItemCart
	}
	return nil
}

func BuyItemFromCart(ctx context.Context, userCollection *mongo.Collection, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsInvalid
	}

	var getCartItems models.User
	var orderCart models.Order

	orderCart.Order_ID = primitive.NewObjectID()
	orderCart.Ordered_At = time.Now()
	orderCart.Order_Cart = make([]models.ProductUser, 0)
	orderCart.Payment_Method.COD = true

	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
	grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}

	currentResults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
	if err != nil {
		panic(err)
	}

	var getUserCart []bson.M
	if err = currentResults.All(ctx, &getUserCart); err != nil {
		panic(err)
	}
	var totalPrice int32

	for _, user_item := range getUserCart {
		price := user_item["total"]
		totalPrice = price.(int32)
	}

	orderCart.Price = int(totalPrice)

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: orderCart}}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}

	err = userCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&getCartItems)
	if err != nil {
		log.Println(err)
	}

	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": bson.M{"$each": getCartItems.User_Cart}}}
	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err != nil {
		log.Println(err)
	}

	userCartEmpty := make([]models.ProductUser, 0)
	filter3 := bson.D{primitive.E{Key: "_id", Value: id}}
	update3 := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "usercart", Value: userCartEmpty}}}}
	_, err = userCollection.UpdateOne(ctx, filter3, update3)
	if err != nil {
		log.Println(err)
		return ErrCantBuyCartItem
	}

}

func InstantBuyer() {}
