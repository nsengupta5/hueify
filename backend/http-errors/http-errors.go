package http_errors

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AlbumInfoFailure(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"message": "Failed to retrieve album info",
		"error":   err,
	})
}

func JSONSerializeFailure(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"message": "Failed to write to album",
		"error":   err,
	})
}

func JSONDeserializeFailure(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"message": "Failed to read album info",
		"error":   err,
	})
}

func GetRelatedArtistsFailure(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"message": "Failed to get related artists",
		"error":   err,
	})
}

func GetCachedRelatedArtistsFailure(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"message": "Failed to get related artists from cache",
		"error":   err,
	})
}

func GetRecommendedAlbumsFailure(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"message": "Failed to get recommended albums",
		"error":   err,
	})
}
