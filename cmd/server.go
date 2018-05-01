package main

import (
	"github.com/gin-gonic/gin"
	"github.com/goglue/docker-registry-oauth/auth"
)

func main() {
	//gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/token", auth.New().Authorize)
	router.Run(":4444")
}
