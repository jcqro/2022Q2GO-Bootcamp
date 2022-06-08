package main

import (
	"2022Q2GO-BOOTCAMP/controllers"

	"github.com/gin-gonic/gin"
)

//App startpoint
func main() {
	router := gin.Default()
	router.GET("/loadbeers", controllers.RunClient)
	router.GET("/beers", controllers.GetBeers)
	router.GET("/beers/:id", controllers.GetBeerById)
	router.GET("/fasterbeers", controllers.GetBeersConcurrently)

	router.Run("localhost:8080")
}
