// creativenobu - "compiled but feels interpreted (02/26/20)
// pheonix_d123 - "Felt Compiled. Might interpret later" (02/26/20)
// sillyonly - "YOU ALWAYS CLEAN MASTER BY FORCE PUSHING THE PERFECT CODE AND NOT THE CODE YOU WROTE" (02/23/20)
// sillyonly - "OR PUSH AFTER AN APPROVED PR" (02/23/20)
// no_neon_one - "have you tried Flutter?" (02/26/20)
package main

import (
	"log"
	"os"

	"gruvee.com/auth"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

// sillyonly - "FINALLLY!! WHO WE ARE CHAT! WHAT DO WE WANT! A DISCOUNT! (I will start a riot)" (02/26/20)
func main() {
	funcframework.RegisterHTTPFunction("/generateCustomToken", auth.GenerateCustomToken)
	funcframework.RegisterHTTPFunction("/authorizeWithSpotify", auth.AuthorizeWithSpotify)
	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
