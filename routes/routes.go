package routes

import (
	"Pint/controllers"
	"Pint/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")

	// User routes
	user := api.Group("/user")
	{
		user.POST("/register", controllers.RegisterUser)
		user.POST("/login", controllers.LoginUser)
		user.GET("/logout", middleware.AuthMiddleware(), controllers.LogoutUser)
		user.GET("/me", middleware.AuthMiddleware(), controllers.MyProfile)
		user.GET("/:id", middleware.AuthMiddleware(), controllers.UserProfile)
		user.POST("/follow/:id", middleware.AuthMiddleware(), controllers.FollowUser)
	}

	// Pin routes
	pin := api.Group("/pin")
	{
		pin.POST("/new", middleware.AuthMiddleware(), controllers.CreatePin)
		pin.GET("/all", middleware.AuthMiddleware(), controllers.GetAllPins)
		pin.GET("/:id", middleware.AuthMiddleware(), controllers.GetSinglePin)
		pin.PUT("/:id", middleware.AuthMiddleware(), controllers.UpdatePin)
		pin.DELETE("/:id", middleware.AuthMiddleware(), controllers.DeletePin)
		pin.POST("/comment/:id", middleware.AuthMiddleware(), controllers.CommentOnPin)
		pin.DELETE("/comment/:id", middleware.AuthMiddleware(), controllers.DeleteComment)
	}
}
