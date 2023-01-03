package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/vitalis-virtus/ecommerce-go/controllers"
	"github.com/vitalis-virtus/ecommerce-go/db"
	"github.com/vitalis-virtus/ecommerce-go/middleware"
	"github.com/vitalis-virtus/ecommerce-go/routes"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
		DisableColors:   false,
	})

	log.SetLevel(log.DebugLevel)
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	app := controllers.NewApplication(db.ProductData(db.Client, "Products"), db.UserData(db.Client, "Users"))

	router := gin.New()
	router.Use(gin.Logger())

	routes.UserRoutes(router)
	router.Use(middleware.Authentication())

	router.GET("/addtocart", app.AddToCart())
	router.GET("/removeitem", app.RemoveItem())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())

	log.Fatal(router.Run(":" + port))
}
