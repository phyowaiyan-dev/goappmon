package utils

import "github.com/gin-gonic/gin"

type ErrorResponse struct {
	Error string `json:"error"`
}

func JSON(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}

func JSONError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, ErrorResponse{Error: message})
}
