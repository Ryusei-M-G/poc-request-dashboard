package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var (
	clientID     string
	clientSecret string
	redirectURL  string
	issuerURL    string
	provider     *oidc.Provider
	oauth2Config oauth2.Config
)

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Load configuration from environment variables
	clientID = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	redirectURL = os.Getenv("REDIRECT_URL")
	issuerURL = os.Getenv("ISSUER_URL")

	var err error
	// Initialize OIDC provider
	provider, err = oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		log.Fatalf("Failed to create OIDC provider: %v", err)
	}

	// Set up OAuth2 config
	oauth2Config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "phone", "openid", "email"},
	}
}

func main() {
	r := gin.Default()

	// CORS設定
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Routes
	r.GET("/login", handleLogin)
	r.GET("/callback", handleCallback)
	r.GET("/logout", handleLogout)

	log.Println("Server is running on http://localhost:8080")
	r.Run(":8080")
}

func handleLogin(c *gin.Context) {
	state := "state" // Replace with a secure random string in production
	url := oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

func handleCallback(c *gin.Context) {
	ctx := context.Background()
	code := c.Query("code")

	// Exchange the authorization code for a token
	rawToken, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to exchange token: "+err.Error())
		return
	}

	// TODO: Store refresh_token in Redis with sessionID
	// sessionID := generateSessionID()
	// rdb.Set(ctx, sessionID, rawToken.RefreshToken, 7*24*time.Hour)
	// c.SetCookie("session_id", sessionID, 7*24*3600, "/", "", true, true)

	// Redirect to frontend with access_token in fragment
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	redirectURL := frontendURL + "/#access_token=" + rawToken.AccessToken
	c.Redirect(http.StatusFound, redirectURL)
}

func handleLogout(c *gin.Context) {
	// Here, you would clear the session or cookie if stored.
	c.Redirect(http.StatusFound, "/")
}
