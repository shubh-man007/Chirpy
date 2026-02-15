package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/shubh-man007/Chirpy/tui/internal/models"
)

type Chirpy struct {
	BaseURL      string
	Client       *http.Client
	AccessToken  string
	RefreshToken string
}

func NewChirpy(client *http.Client) *Chirpy {
	return &Chirpy{
		Client: client,
	}
}

func (c *Chirpy) Login(email, password string) (*models.LoginResponse, error) {
	payload := models.LoginCreds{
		Password: password,
		Email:    email,
	}

	payloadJSON, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", c.BaseURL+"/api/login", bytes.NewBuffer(payloadJSON))
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}

	res, err := c.Client.Do(req)
	if err != nil {
		log.Printf("Error getting response: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	loginRes := models.LoginResponse{}
	err = json.NewDecoder(res.Body).Decode(&loginRes)
	if err != nil {
		return nil, err
	}

	c.AccessToken = loginRes.Token
	c.RefreshToken = loginRes.RefreshToken

	return &loginRes, nil
}

func (c *Chirpy) Register(email, password string) (*models.User, error) {
	payload := models.LoginCreds{
		Password: password,
		Email:    email,
	}

	payloadJSON, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", c.BaseURL+"/api/users", bytes.NewBuffer(payloadJSON))
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}

	res, err := c.Client.Do(req)
	if err != nil {
		log.Printf("Error getting response: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	userRes := models.User{}
	err = json.NewDecoder(res.Body).Decode(&userRes)
	if err != nil {
		return nil, err
	}

	return &userRes, nil
}

func (c *Chirpy) UpdateUserCredentials(email, password string) (*models.User, error) {
	credUp := models.LoginCreds{
		Password: password,
		Email:    email,
	}

	credUpJSON, _ := json.Marshal(credUp)

	req, err := http.NewRequest("PUT", c.BaseURL+"/api/users", bytes.NewBuffer(credUpJSON))
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}

	res, err := c.Client.Do(req)
	if err != nil {
		log.Printf("Error getting response: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	userRes := models.User{}
	err = json.NewDecoder(res.Body).Decode(&userRes)
	if err != nil {
		return nil, err
	}

	return &userRes, nil
}

func (c *Chirpy) DeleteUser(userID string) error {
	req, err := http.NewRequest("DELETE", c.BaseURL+"/api/users/"+userID, nil)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer"+c.AccessToken)

	res, err := c.Client.Do(req)
	if err != nil {
		log.Printf("Error getting response: %v", err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	return nil
}

func (c *Chirpy) CreateChirp(body string) (*models.Chirp, error) {
	chirpBody := models.ChirpBody{Body: body}

	chirpBodyJSON, _ := json.Marshal(chirpBody)

	req, err := http.NewRequest("POST", c.BaseURL+"/api/chirps", bytes.NewBuffer(chirpBodyJSON))
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer"+c.AccessToken)

	res, err := c.Client.Do(req)
	if err != nil {
		log.Printf("Error getting response: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	chirp := models.Chirp{}
	err = json.NewDecoder(res.Body).Decode(&chirp)
	if err != nil {
		return nil, err
	}

	return &chirp, nil
}

func (c *Chirpy) UpdateChirp(chirpID, body string) (*models.Chirp, error) {
	chirpBody := models.ChirpBody{Body: body}

	chirpBodyJSON, _ := json.Marshal(chirpBody)

	req, err := http.NewRequest("PATCH", c.BaseURL+"/api/chirps/"+chirpID, bytes.NewBuffer(chirpBodyJSON))
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer"+c.AccessToken)

	res, err := c.Client.Do(req)
	if err != nil {
		log.Printf("Error getting response: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	chirp := models.Chirp{}
	err = json.NewDecoder(res.Body).Decode(&chirp)
	if err != nil {
		return nil, err
	}

	return &chirp, nil
}

func (c *Chirpy) DeleteChirp(chirpID string) error {
	req, err := http.NewRequest("DELETE", c.BaseURL+"/api/chirps/"+chirpID, nil)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer"+c.AccessToken)

	res, err := c.Client.Do(req)
	if err != nil {
		log.Printf("Error getting response: %v", err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	return nil
}
