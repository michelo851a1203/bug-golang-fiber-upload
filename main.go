package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func UploadHandler(ctx *fiber.Ctx) error {
	logger, err := zap.NewProduction()
	defer logger.Sync()
	if err != nil {
		panic("zap log error")
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		logger.Error("file header error",
			zap.String("error", err.Error()),
		)
		ctx.Status(fiber.StatusInternalServerError)
		return nil
	}
	// multipart fileheader --> file
	fileMultiPart, err := fileHeader.Open()
	defer fileMultiPart.Close()
	if err != nil {
		logger.Error("file multipart error",
			zap.String("error", err.Error()),
		)
		ctx.Status(fiber.StatusInternalServerError)
		return nil
	}

	fileByte, err := io.ReadAll(fileMultiPart)

	if err != nil {
		logger.Error("file byte error",
			zap.String("error", err.Error()),
		)
		ctx.Status(fiber.StatusInternalServerError)
		return nil
	}

	baseStr := base64.StdEncoding.EncodeToString(fileByte)

	detectContentType := http.DetectContentType(fileByte)
	switch detectContentType {
	case "image/jpeg", "image/png", "application/pdf":
		return ctx.JSON(map[string]interface{}{
			"msg": fmt.Sprintf("data:%s;base64,%s", detectContentType, baseStr),
		})
	}

	ctx.Status(fiber.StatusMethodNotAllowed)
	return nil
}

var port string

func CreateFiberApp() *fiber.App {
	app := fiber.New()
	app.Post("/upload", UploadHandler)
	return app
}

func main() {
	flag.StringVar(&port, "p", "8080", "default port")
	flag.Parse()
	logger, err := zap.NewProduction()
	defer logger.Sync()
	if err != nil {
		panic("zap error")
	}

	app := CreateFiberApp()

	quitChan := make(chan os.Signal, 1)
	finishShutdownChan := make(chan struct{})
	signal.Notify(quitChan, os.Interrupt)

	go func() {
		<-quitChan
		if err := app.Shutdown(); err != nil {
			logger.Error("shutdown error",
				zap.String("error", err.Error()),
			)
			return
		}
		finishShutdownChan <- struct{}{}
	}()

	if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		logger.Error("start server error ...",
			zap.String("error", err.Error()),
		)
		return
	}

	<-finishShutdownChan
	logger.Info("clean task")
}
