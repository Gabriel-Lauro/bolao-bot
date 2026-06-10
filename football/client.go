package football

import (
	"bolao-bot/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://api.football-data.org/v4"

type Client struct {
	apiKey      string
	competition string
	http        *http.Client
}

func New(apiKey, competition string) *Client {
	return &Client{
		apiKey:      apiKey,
		competition: competition,
		http:        &http.Client{Timeout: 15 * time.Second},
	}
}

type apiResponse struct {
	Matches []apiMatch `json:"matches"`
}

type apiMatch struct {
	ID      int    `json:"id"`
	Status  string `json:"status"`
	Stage   string `json:"stage"`
	UTCDate string `json:"utcDate"`
	HomeTeam struct {
		Name string `json:"name"`
	} `json:"homeTeam"`
	AwayTeam struct {
		Name string `json:"name"`
	} `json:"awayTeam"`
	Score struct {
		FullTime struct {
			Home *int `json:"home"`
			Away *int `json:"away"`
		} `json:"fullTime"`
	} `json:"score"`
}

func (c *Client) FetchCalendar() ([]*models.Match, error) {
	url := fmt.Sprintf("%s/competitions/%s/matches", baseURL, c.competition)
	return c.fetch(url)
}

func (c *Client) FetchMatch(id int) (*models.Match, error) {
	url := fmt.Sprintf("%s/matches/%d", baseURL, id)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Auth-Token", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API status %d", resp.StatusCode)
	}

	var m apiMatch
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return convert(m), nil
}

func (c *Client) fetch(url string) ([]*models.Match, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Auth-Token", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API status %d", resp.StatusCode)
	}

	var data apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var matches []*models.Match
	for _, m := range data.Matches {
		matches = append(matches, convert(m))
	}
	return matches, nil
}

func convert(m apiMatch) *models.Match {
	t, _ := time.Parse(time.RFC3339, m.UTCDate)
	match := &models.Match{
		ID:        m.ID,
		HomeTeam:  Translate(m.HomeTeam.Name),
		AwayTeam:  Translate(m.AwayTeam.Name),
		Stage:     m.Stage,
		StartsAt:  t,
		Status:    mapStatus(m.Status),
		HomeScore: m.Score.FullTime.Home,
		AwayScore: m.Score.FullTime.Away,
	}
	return match
}

func mapStatus(s string) string {
	switch s {
	case "FINISHED":
		return "finished"
	case "IN_PLAY", "PAUSED":
		return "live"
	default:
		return "scheduled"
	}
}
