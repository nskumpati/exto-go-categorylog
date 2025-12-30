package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetUserFromMicrosoftToken(c *gin.Context) (*GoogleTokenUser, error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		return nil, errors.New("authorization token is required")
	}
	// Remove "Bearer " prefix if it exists
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	claims, err := ValidateMicrosoftToken(token)
	if err != nil {
		return nil, errors.New("invalid Microsoft access token")
	}
	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return nil, errors.New("email not found in Microsoft token")
	}
	exp, _ := claims["exp"].(int64)
	return &GoogleTokenUser{
		Email:         email,
		EmailVerified: true,
		Exp:           exp,
		Hd:            "",
	}, nil
}

func ValidateMicrosoftToken(token string) (map[string]interface{}, error) {
	if token == "" {
		return nil, errors.New("authorization token is required")
	}
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}
	accessToken := strings.TrimPrefix(token, "Bearer ")
	req, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Graph API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graph API returned %d", resp.StatusCode)
	}

	var user struct {
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	email := user.Mail
	if email == "" {
		email = user.UserPrincipalName
	}
	if email == "" {
		return nil, errors.New("email not found in Microsoft token")
	}

	claims := map[string]interface{}{
		"email": email,
		"exp":   time.Now().Add(1 * time.Hour).Unix(), // Graph doesn't return exp
	}
	return claims, nil
}
