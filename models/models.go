package models

import "time"

type User struct {
	ID       string
	Username string
}

type Match struct {
	ID        int
	HomeTeam  string
	AwayTeam  string
	HomeScore *int
	AwayScore *int
	Stage     string
	StartsAt  time.Time
	Status    string
}

func (m *Match) IsOpen() bool {
	return time.Now().Before(m.StartsAt)
}

func (m *Match) Winner() string {
	if m.HomeScore == nil || m.AwayScore == nil {
		return ""
	}
	if *m.HomeScore > *m.AwayScore {
		return "home"
	}
	if *m.AwayScore > *m.HomeScore {
		return "away"
	}
	return "draw"
}

func (m *Match) Label() string {
	return m.HomeTeam + " × " + m.AwayTeam
}

type Guess struct {
	ID         int
	UserID     string
	MatchID    int
	HomeScore  int
	AwayScore  int
	Points     int
	Calculated bool
}

func (g *Guess) Winner() string {
	if g.HomeScore > g.AwayScore {
		return "home"
	}
	if g.AwayScore > g.HomeScore {
		return "away"
	}
	return "draw"
}

func CalcPoints(g *Guess, m *Match) int {
	if m.HomeScore == nil || m.AwayScore == nil {
		return 0
	}
	if g.HomeScore == *m.HomeScore && g.AwayScore == *m.AwayScore {
		return 3
	}
	if g.Winner() == m.Winner() {
		return 1
	}
	return 0
}

type RankingEntry struct {
	Position int
	UserID   string
	Username string
	Total    int
	Exact    int
	Winner   int
	Guesses  int
}
