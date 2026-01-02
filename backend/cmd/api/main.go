package main

import (
	"log"
	"net/http"
	"os"
	"time"

	_ "example.com/webwhatsapp/backend/docs"
	"example.com/webwhatsapp/backend/internal/bootstrap"
	// Eğer docs.SwaggerInfo'yu runtime set etmek istersen:
	// "example.com/webwhatsapp/backend/docs"
)

// Swagger annotations
// @title           Vue-WhatsApp API
// @version         1.0
// @description     API documentation
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	app, err := bootstrap.Build()
	if err != nil {
		log.Fatal(err)
	}

	// (Opsiyonel) Reverse proxy arkasında basePath/host set etmek istersen:
	// swaggerHost := os.Getenv("SWAGGER_HOST")         // ör: localhost:8088
	// swaggerBase := os.Getenv("SWAGGER_BASE_PATH")    // ör: /api
	// if swaggerHost != "" {
	// 	docs.SwaggerInfo.Host = swaggerHost
	// }
	// if swaggerBase != "" {
	// 	docs.SwaggerInfo.BasePath = swaggerBase
	// }

	port := app.Port
	if port == "" {
		port = os.Getenv("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: app.Router,

		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("backend listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}
