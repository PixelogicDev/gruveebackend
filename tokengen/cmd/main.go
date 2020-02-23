package main

import (
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	generator "github.com/PixelogicDev/Gruvee-Backend/tokengen/tokengen"
)

func main() {
	funcframework.RegisterHTTPFunction("/generateToken", generator.GenerateToken)
	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}