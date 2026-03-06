package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	clerkSignInURL  = "https://clerk.isometricon.com/v1/client/sign_ins"
	iconGenerateURL = "https://www.isometricon.com/api/icons/text"
	creditsURL      = "https://www.isometricon.com/api/user/credits"
	tokenFileName   = ".isometricon-token.json"
)

type Client struct {
	sessionToken string
	httpClient   *http.Client
}

type storedToken struct {
	JWT       string `json:"jwt"`
	ExpiresAt int64  `json:"expires_at"`
}

func New() *Client {
	c := &Client{
		httpClient: &http.Client{},
	}
	c.loadToken()
	return c
}

func (c *Client) IsLoggedIn() bool {
	return c.sessionToken != ""
}

func tokenPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return tokenFileName
	}
	dir := filepath.Join(home, ".config", "isometricon")
	os.MkdirAll(dir, 0700)
	return filepath.Join(dir, tokenFileName)
}

func (c *Client) saveToken() {
	t := storedToken{
		JWT:       c.sessionToken,
		ExpiresAt: time.Now().Add(55 * time.Second).Unix(),
	}
	data, _ := json.Marshal(t)
	os.WriteFile(tokenPath(), data, 0600)
}

func (c *Client) loadToken() {
	data, err := os.ReadFile(tokenPath())
	if err != nil {
		return
	}
	var t storedToken
	if err := json.Unmarshal(data, &t); err != nil {
		return
	}
	if time.Now().Unix() < t.ExpiresAt {
		c.sessionToken = t.JWT
	}
}

// Login authenticates via Clerk and stores the JWT session token.
func (c *Client) Login(email, password string) (string, error) {
	form := url.Values{}
	form.Set("identifier", email)
	form.Set("password", password)
	form.Set("strategy", "password")

	req, err := http.NewRequest("POST", clerkSignInURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://www.isometricon.com")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Client struct {
			Sessions []struct {
				User struct {
					FirstName string `json:"first_name"`
					LastName  string `json:"last_name"`
					Username  string `json:"username"`
				} `json:"user"`
				LastActiveToken struct {
					JWT string `json:"jwt"`
				} `json:"last_active_token"`
			} `json:"sessions"`
		} `json:"client"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parsing login response: %w", err)
	}

	if len(result.Client.Sessions) == 0 {
		return "", fmt.Errorf("no active session returned from login")
	}

	session := result.Client.Sessions[0]
	c.sessionToken = session.LastActiveToken.JWT

	if c.sessionToken == "" {
		return "", fmt.Errorf("no JWT token in login response")
	}

	c.saveToken()

	name := session.User.FirstName
	if session.User.Username != "" {
		name = session.User.Username
	}

	return name, nil
}

type GenerateResult struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	FilePath string `json:"-"`
}

// GenerateIcon generates an isometric icon from a text prompt.
func (c *Client) GenerateIcon(prompt, outputPath string) (*GenerateResult, error) {
	if !c.IsLoggedIn() {
		return nil, fmt.Errorf("not logged in — call the login tool first, or set ISOMETRICON_EMAIL and ISOMETRICON_PASSWORD env vars")
	}

	payload := fmt.Sprintf(`{"prompt":%q}`, prompt)
	req, err := http.NewRequest("POST", iconGenerateURL, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating generate request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.isometricon.com")
	req.Header.Set("Referer", "https://www.isometricon.com/app")
	req.Header.Set("Authorization", "Bearer "+c.sessionToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("generate request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("generate failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		ID       string `json:"id"`
		ImageB64 string `json:"imageB64"`
		URL      string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("parsing generate response: %w", err)
	}

	// Decode base64 and save to file
	imgData, err := base64.StdEncoding.DecodeString(apiResp.ImageB64)
	if err != nil {
		return nil, fmt.Errorf("decoding image data: %w", err)
	}

	if outputPath == "" {
		outputPath = "./icon.png"
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, imgData, 0644); err != nil {
		return nil, fmt.Errorf("writing image file: %w", err)
	}

	absPath, _ := filepath.Abs(outputPath)

	return &GenerateResult{
		ID:       apiResp.ID,
		URL:      apiResp.URL,
		FilePath: absPath,
	}, nil
}

type Credits struct {
	Balance          int     `json:"balance"`
	Role             int     `json:"role"`
	Unlimited        bool    `json:"unlimited"`
	SubscriptionTier *string `json:"subscription_tier"`
}

// CheckCredits returns the current credit balance.
func (c *Client) CheckCredits() (*Credits, error) {
	if !c.IsLoggedIn() {
		return nil, fmt.Errorf("not logged in — call the login tool first, or set ISOMETRICON_EMAIL and ISOMETRICON_PASSWORD env vars")
	}

	req, err := http.NewRequest("GET", creditsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating credits request: %w", err)
	}
	req.Header.Set("Referer", "https://www.isometricon.com/app")
	req.Header.Set("Authorization", "Bearer "+c.sessionToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("credits request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("credits check failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var credits Credits
	if err := json.NewDecoder(resp.Body).Decode(&credits); err != nil {
		return nil, fmt.Errorf("parsing credits response: %w", err)
	}

	return &credits, nil
}
