package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

type ClaimsPage struct {
	AccessToken string
	Claims      jwt.MapClaims
}

var (
	clientID     string
	clientSecret string
	redirectURL  string
	issuerURL    string
	provider     *oidc.Provider
	oauth2Config oauth2.Config
)

func handleHome(c *gin.Context) {
	html := `
        <html>
        <body>
            <h1>Welcome to Cognito OIDC Go App</h1>
            <a href="/login">Login with Cognito</a>
        </body>
        </html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
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
	tokenString := rawToken.AccessToken

	// Parse the token (do signature verification for your use case in production)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		c.String(http.StatusInternalServerError, "Error parsing token: "+err.Error())
		return
	}

	// Check if the token is valid and extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.String(http.StatusBadRequest, "Invalid claims")
		return
	}

	// Render HTML template
	c.HTML(http.StatusOK, "claims.html", gin.H{
		"AccessToken": tokenString,
		"Claims":      claims,
	})
}

func handleLogout(c *gin.Context) {
	// Here, you would clear the session or cookie if stored.
	c.Redirect(http.StatusFound, "/")
}

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

	// Load HTML templates
	r.SetHTMLTemplate(claimsTemplate())

	// Routes
	r.GET("/", handleHome)
	r.GET("/login", handleLogin)
	r.GET("/logout", handleLogout)
	r.GET("/callback", handleCallback)

	log.Println("Server is running on http://localhost:8080")
	r.Run(":8080")
}

func claimsTemplate() *template.Template {
	tmpl := `
    <html>
        <body>
            <h1>User Information</h1>
            <h1>JWT Claims</h1>
            <p><strong>Access Token:</strong> {{.AccessToken}}</p>
            <ul>
                {{range $key, $value := .Claims}}
                    <li><strong>{{$key}}:</strong> {{$value}}</li>
                {{end}}
            </ul>
            <a href="/logout">Logout</a>
        </body>
    </html>`
	return template.Must(template.New("claims.html").Parse(tmpl))
}
