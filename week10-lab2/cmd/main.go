package main

import (
    _ "week10-lab2/docs"
    "week10-lab2/internal/handler"

    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Simple API Example
// @version         1.0
// @description     This is a simple example of using Gin with Swagger.
// @host            localhost:9999
// @BasePath        /api/v1
func main() {
	r := gin.Default()

	config := cors.DefaultConfig()
    config.AllowAllOrigins = true
    r.Use(cors.New(config))
	// Swagger endpoint
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// User API routes
	api := r.Group("/api/v1")
	{
		api.GET("books/:id", handler.GetBookByID) // ใช้ Handler จากไฟล์ book_handler.go
	}

	// Start server
	r.Run(":9999")
}