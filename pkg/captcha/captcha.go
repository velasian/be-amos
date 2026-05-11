package captcha

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type TurnstileResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func Verify(token string) error {
	secretKey := os.Getenv("TURNSTILE_SECRET_KEY")
	if secretKey == "" {
		return fmt.Errorf("TURNSTILE_SECRET_KEY is not set")
	}

	if token == "" {
		return fmt.Errorf("captcha token is empty")
	}

	formData := url.Values{}
	formData.Set("secret", secretKey)
	formData.Set("response", token)

	resp, err := http.PostForm(turnstileVerifyURL, formData)
	if err != nil {
		return fmt.Errorf("failed to verify captcha: %v", err)
	}
	defer resp.Body.Close()

	var result TurnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode captcha response: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("captcha verification failed: %v", result.ErrorCodes)
	}

	return nil
}
