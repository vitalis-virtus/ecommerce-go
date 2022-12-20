package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vitalis-virtus/ecommerce-go/controllers"
)

func UserRoutes(incommingRoutes *gin.Engine) {
	incommingRoutes.POST("/users/signup", controllers.Signup())
	incommingRoutes.POST("/users/login", controllers.Login())
	incommingRoutes.POST("/admin/addproduct", controllers.ProductViewerAdmin())
	incommingRoutes.GET("/users/productview", controllers.SearchProduct())
	incommingRoutes.GET("/users/search", controllers.SearchProductByQuery())
}
