package internal

import (
	"Ezidal/YandexOSUploaderPhotos/internal/storage"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func MainPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func Upload(c *gin.Context) {
	s3 := storage.LoadStorage()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Ошибка загрузки файла")
		return
	}
	defer file.Close()

	bucket := "golang-uploader-photos"
	key := fileHeader.Filename

	err = s3.UploadFile(ctx, file, fileHeader, bucket, fileHeader.Filename, "img/"+key)
	if err != nil {
		c.String(http.StatusInternalServerError, "Ошибка загрузки файла в облако")
		return
	}
	url := fmt.Sprintf("https://%s.storage.yandexcloud.net/%s", bucket, key)

	c.HTML(http.StatusOK, "index.html", gin.H{"url": url})

}
