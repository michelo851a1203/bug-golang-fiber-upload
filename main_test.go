package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func createImage() *image.RGBA {
	width := 100
	height := 100
	fromPoint := image.Point{0, 0}
	toPoint := image.Point{width, height}
	img := image.NewRGBA(image.Rectangle{fromPoint, toPoint})
	cyan := color.RGBA{100, 100, 200, 0xff}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2, x > width/2 && y > height/2:
				img.Set(x, y, cyan)
			}
		}
	}

	return img
}

func TestUploadHandler(t *testing.T) {
	assert := assert.New(t)
	pr, pw := io.Pipe()
	imgChan := make(chan string)

	writer := multipart.NewWriter(pw)

	go func() {
		defer writer.Close()
		imageWriter, err := writer.CreateFormFile("file", "test.png")
		assert.Nil(err)

		img := createImage()
		err = png.Encode(imageWriter, img)
		assert.Nil(err)
		// okay here need to remember to writer to pipe

		buf := new(bytes.Buffer)
		err = png.Encode(buf, img)
		assert.Nil(err)
		imgByteContainer := buf.Bytes()

		baseStr := base64.StdEncoding.EncodeToString(imgByteContainer)

		jsonByte, err := json.Marshal(map[string]interface{}{
			"msg": fmt.Sprintf("data:image/png;base64,%s", baseStr),
		})

		assert.Nil(err)

		imgChan <- string(jsonByte)
	}()

	req, err := http.NewRequest("POST", "/upload", pr)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	assert.Nil(err)

	app := CreateFiberApp()
	res, err := app.Test(req, 1)
	assert.Nil(err)
	assert.Equal(res.StatusCode, fiber.StatusOK)
	bodyReader, err := io.ReadAll(res.Body)
	assert.Nil(err)
	assert.Contains(string(bodyReader), <-imgChan)
}

func TestUploadSolvedWithBufferHandler(t *testing.T) {

	assert := assert.New(t)
	body := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(body)

	imgChan := make(chan string, 1)

	go func() {
		defer writer.Close()
		ioWriter, err := writer.CreateFormFile("file", "test.png")

		assert.Nil(err)
		img := createImage()
		err = png.Encode(ioWriter, img)
		assert.Nil(err)

		buf := new(bytes.Buffer)

		err = png.Encode(buf, img)
		assert.Nil(err)
		byteContainer := buf.Bytes()

		baseStr := base64.StdEncoding.EncodeToString(byteContainer)

		jsonByte, err := json.Marshal(map[string]interface{}{
			"msg": fmt.Sprintf("data:image/png;base64,%s", baseStr),
		})

		assert.Nil(err)
		imgChan <- string(jsonByte)
	}()

	imgStr := <-imgChan

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	app := CreateFiberApp()
	res, err := app.Test(req)
	assert.Nil(err)

	assert.Equal(res.StatusCode, fiber.StatusOK)
	bodyByte, err := io.ReadAll(res.Body)
	assert.Nil(err)

	assert.Contains(string(bodyByte), imgStr)

}
