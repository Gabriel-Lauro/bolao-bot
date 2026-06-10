package db

import (
	"bolao-bot/models"
	"database/sql"
	_ "embed"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

type DB struct {
	conn *sql.DB
}

func New(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path+"?_journal=WAL&_timeout=5000&_fk=true")
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	if _, err := conn.Exec(schema); err != nil {
		return nil, err
	}
	log.Println("Banco conectado:", path)
	return &DB{conn: conn}, nil
}

// ── Users ──────────────────────────────────────────────────────────────────

func (d *DB) UpsertUser(u *models.User) error {
	_, err := d.conn.Exec(`
		INSERT INTO users (id, username) VALUES (?, ?)
		ON CONFLICT(id) DO UPDATE SET username = ?
	`, u.ID, u.Username, u.Username)
	return err
}

func (d *DB) GetUser(id string) (*models.User, error) {
	row := d.conn.QueryRow(`SELECT id, username FROM users WHERE id = ?`, id)
	u := &models.User{}
	err := row.Scan(&u.ID, &u.Username)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// ── Matches ────────────────────────────────────────────────────────────────

func (d *DB) UpsertMatch(m *models.Match) error {
	_, err := d.conn.Exec(`
		INSERT INTO matches (id, home_team, away_team, stage, starts_at, status, home_score, away_score)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			home_score = excluded.home_score,
			away_score = excluded.away_score,
			status     = excluded.status
	`, m.ID, m.HomeTeam, m.AwayTeam, m.Stage, m.StartsAt, m.Status, m.HomeScore, m.AwayScore)
	return err
}

func (d *DB) UpdateMatchScore(id, home, away int, status string) error {
	_, err := d.conn.Exec(`
		UPDATE matches SET home_score=?, away_score=?, status=? WHERE id=?
	`, home, away, status, id)
	return err
}

func (d *DB) GetMatch(id int) (*models.Match, error) {
	row := d.conn.QueryRow(`
		SELECT id, home_team, away_team, home_score, away_score, stage, starts_at, status
		FROM matches WHERE id = ?
	`, id)
	return scanMatch(row)
}

func (d *DB) GetOpenMatches() ([]*models.Match, error) {
	rows, err := d.conn.Query(`
		SELECT id, home_team, away_team, home_score, away_score, stage, starts_at, status
		FROM matches WHERE status = 'scheduled' AND starts_at > datetime('now')
		ORDER BY starts_at ASC LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (d *DB) GetTodayMatches() ([]*models.Match, error) {
	rows, err := d.conn.Query(`
		SELECT id, home_team, away_team, home_score, away_score, stage, starts_at, status
		FROM matches
		WHERE date(starts_at) = date('now') AND status != 'finished'
		ORDER BY starts_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

// ── Guesses ────────────────────────────────────────────────────────────────

func (d *DB) UpsertGuess(g *models.Guess) error {
	_, err := d.conn.Exec(`
		INSERT INTO guesses (user_id, match_id, home_score, away_score)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, match_id) DO UPDATE SET
			home_score = excluded.home_score,
			away_score = excluded.away_score,
			points     = 0,
			calculated = 0
	`, g.UserID, g.MatchID, g.HomeScore, g.AwayScore)
	return err
}

func (d *DB) GetUserGuesses(userID string) ([]*models.Guess, error) {
	rows, err := d.conn.Query(`
		SELECT g.id, g.user_id, g.match_id, g.home_score, g.away_score, g.points, g.calculated,
		       m.home_team, m.away_team, m.home_score, m.away_score, m.status
		FROM guesses g
		JOIN matches m ON m.id = g.match_id
		WHERE g.user_id = ?
		ORDER BY m.starts_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guesses []*models.Guess
	for rows.Next() {
		g := &models.Guess{}
		var calc int
		var homeTeam, awayTeam, status string
		var mHome, mAway *int
		if err := rows.Scan(&g.ID, &g.UserID, &g.MatchID, &g.HomeScore, &g.AwayScore,
			&g.Points, &calc, &homeTeam, &awayTeam, &mHome, &mAway, &status); err != nil {
			return nil, err
		}
		g.Calculated = calc == 1
		guesses = append(guesses, g)
	}
	return guesses, nil
}

func (d *DB) GetUserGuessesWithMatches(userID string) ([]GuessWithMatch, error) {
	rows, err := d.conn.Query(`
		SELECT g.id, g.user_id, g.match_id, g.home_score, g.away_score, g.points, g.calculated,
		       m.home_team, m.away_team, m.home_score, m.away_score, m.status, m.starts_at
		FROM guesses g
		JOIN matches m ON m.id = g.match_id
		WHERE g.user_id = ?
		ORDER BY m.starts_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []GuessWithMatch
	for rows.Next() {
		var gm GuessWithMatch
		var calc int
		if err := rows.Scan(
			&gm.Guess.ID, &gm.Guess.UserID, &gm.Guess.MatchID,
			&gm.Guess.HomeScore, &gm.Guess.AwayScore, &gm.Guess.Points, &calc,
			&gm.HomeTeam, &gm.AwayTeam, &gm.MatchHomeScore, &gm.MatchAwayScore,
			&gm.Status, &gm.StartsAt,
		); err != nil {
			return nil, err
		}
		gm.Guess.Calculated = calc == 1
		result = append(result, gm)
	}
	return result, nil
}

type GuessWithMatch struct {
	Guess          models.Guess
	HomeTeam       string
	AwayTeam       string
	MatchHomeScore *int
	MatchAwayScore *int
	Status         string
	StartsAt       time.Time
}

func (d *DB) GetMatchGuesses(matchID int) ([]*models.Guess, error) {
	rows, err := d.conn.Query(`
		SELECT id, user_id, match_id, home_score, away_score, points, calculated
		FROM guesses WHERE match_id = ? AND calculated = 0
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guesses []*models.Guess
	for rows.Next() {
		g := &models.Guess{}
		var calc int
		if err := rows.Scan(&g.ID, &g.UserID, &g.MatchID, &g.HomeScore, &g.AwayScore, &g.Points, &calc); err != nil {
			return nil, err
		}
		g.Calculated = calc == 1
		guesses = append(guesses, g)
	}
	return guesses, nil
}

func (d *DB) SetGuessPoints(id, points int) error {
	_, err := d.conn.Exec(`UPDATE guesses SET points=?, calculated=1 WHERE id=?`, points, id)
	return err
}

// ── Ranking ────────────────────────────────────────────────────────────────

func (d *DB) GetRanking() ([]*models.RankingEntry, error) {
	rows, err := d.conn.Query(`
		SELECT
			u.id, u.username,
			COALESCE(SUM(g.points), 0),
			COALESCE(SUM(CASE WHEN g.points=3 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN g.points=1 THEN 1 ELSE 0 END), 0),
			COUNT(g.id)
		FROM users u
		LEFT JOIN guesses g ON g.user_id = u.id
		GROUP BY u.id
		ORDER BY SUM(g.points) DESC, SUM(CASE WHEN g.points=3 THEN 1 ELSE 0 END) DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranking []*models.RankingEntry
	pos := 1
	for rows.Next() {
		e := &models.RankingEntry{Position: pos}
		if err := rows.Scan(&e.UserID, &e.Username, &e.Total, &e.Exact, &e.Winner, &e.Guesses); err != nil {
			return nil, err
		}
		ranking = append(ranking, e)
		pos++
	}
	return ranking, nil
}

// ── Sync log ───────────────────────────────────────────────────────────────

func (d *DB) LogSync(t string) {
	d.conn.Exec(`INSERT INTO sync_log (type) VALUES (?)`, t)
}

func (d *DB) TodaySyncCount() int {
	row := d.conn.QueryRow(`SELECT COUNT(*) FROM sync_log WHERE date(synced_at)=date('now')`)
	var n int
	row.Scan(&n)
	return n
}

// ── Helpers ────────────────────────────────────────────────────────────────

func scanMatch(row *sql.Row) (*models.Match, error) {
	m := &models.Match{}
	err := row.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.HomeScore, &m.AwayScore, &m.Stage, &m.StartsAt, &m.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

func scanMatches(rows *sql.Rows) ([]*models.Match, error) {
	var list []*models.Match
	for rows.Next() {
		m := &models.Match{}
		if err := rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.HomeScore, &m.AwayScore, &m.Stage, &m.StartsAt, &m.Status); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}
