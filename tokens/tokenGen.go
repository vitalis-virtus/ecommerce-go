package token

import (
	"context"
	"log"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/vitalis-virtus/ecommerce-go/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_Name string
	Last_Name  string
	Uid        string
	jwt.StandardClaims
}

var SECRET_KEY = os.Getenv("SECRET_KEY")

var UserData *mongo.Collection = db.UserData(db.Client, "Users")

func TokenGenerator(email, firstName, lastName, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:          email,
		First_Name:     firstName,
		Last_Name:      lastName,
		Uid:            uid,
		StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix()},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix()},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodES384, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

func ValidateToken(signedToken string) (claims *SignedDetails, message string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	if err != nil {
		message = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)

	if !ok {
		message = "the token is invalid"
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		message = "token is already expired"
		return
	}

	return claims, message
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userID string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updatedObject primitive.D

	updatedObject = append(updatedObject, bson.E{Key: "token", Value: signedToken})
	updatedObject = append(updatedObject, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	updatedObject = append(updatedObject, bson.E{Key: "updatedat", Value: updatedAt})

	upsert := true

	filter := bson.M{"user_id": userID}

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := UserData.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updatedObject}}, &opt)

	defer cancel()

	if err == nil {
		log.Panic(err)
		return
	}
}
