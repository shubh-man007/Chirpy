package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

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

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

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

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

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

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

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

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

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

// sort = {asc, desc}
func (c *Chirpy) GetMyChirps(sort string) ([]models.Chirp, error) {
	var URL string
	switch sort {
	case "desc":
		URL = c.BaseURL + "/api/me/chirps?sort=desc"
	default:
		URL = c.BaseURL + "/api/me/chirps"
	}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

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

	chirps := []models.Chirp{}
	err = json.NewDecoder(res.Body).Decode(&chirps)
	if err != nil {
		log.Printf("Something went wrong: %v", err)
		return nil, err
	}

	return chirps, nil
}

func (c *Chirpy) GetChirpsByUser(userID, sort string) ([]models.Chirp, error) {
	var URL string
	switch sort {
	case "desc":
		URL = c.BaseURL + "/api/users/" + userID + "/chirps?sort=desc"
	default:
		URL = c.BaseURL + "/api/users/" + userID + "/chirps"
	}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

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

	chirps := []models.Chirp{}
	err = json.NewDecoder(res.Body).Decode(&chirps)
	if err != nil {
		log.Printf("Something went wrong: %v", err)
		return nil, err
	}

	return chirps, nil
}

// limit = 20 ; offset = 0
func (c *Chirpy) GetFeed(limit, offset int) ([]models.Chirp, error) {
	u, err := url.Parse(c.BaseURL + "/api/feed")
	if err != nil {
		return nil, err
	}

	q := u.Query()

	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	var chirps []models.Chirp
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&chirps); err != nil {
		if errors.Is(err, io.EOF) {
			return []models.Chirp{}, nil
		}
		return nil, err
	}

	return chirps, nil
}

func (c *Chirpy) FollowUser(userID string) error {
	followee := models.Follow{FolloweeID: userID}

	followeeJSON, _ := json.Marshal(followee)

	req, err := http.NewRequest("POST", c.BaseURL+"/api/follow", bytes.NewBuffer(followeeJSON))
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

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

func (c *Chirpy) UnfollowUser(userID string) error {
	req, err := http.NewRequest("DELETE", c.BaseURL+"/api/follow/"+userID, nil)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

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

func (c *Chirpy) GetFollowers() (*models.FollowResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/followers", nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	res, err := c.Client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	var response models.FollowResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		log.Printf("Error decoding followers response: %v", err)
		return nil, err
	}

	return &response, nil
}

func (c *Chirpy) GetFollowing() (*models.FollowResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/following", nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	res, err := c.Client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	var response models.FollowResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		log.Printf("Error decoding following response: %v", err)
		return nil, err
	}

	return &response, nil
}
