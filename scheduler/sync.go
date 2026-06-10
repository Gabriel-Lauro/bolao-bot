package scheduler

import (
	"bolao-bot/db"
	"bolao-bot/football"
	"bolao-bot/models"
	"log"
	"sync"
	"time"
)

type OnMatchFinished func(match *models.Match, guesses []*models.Guess, usernames map[string]string)

type Scheduler struct {
	db          *db.DB
	football    *football.Client
	onFinished  OnMatchFinished
	polling     map[int]bool
	mu          sync.Mutex
}

func New(database *db.DB, fc *football.Client, onFinished OnMatchFinished) *Scheduler {
	return &Scheduler{
		db:         database,
		football:   fc,
		onFinished: onFinished,
		polling:    make(map[int]bool),
	}
}

func (s *Scheduler) Start() {
	log.Println("Scheduler iniciado")
	go s.syncCalendar()
	go s.runDaily()
	go s.watchMatches()
}

func (s *Scheduler) ForceSync() {
	go s.syncCalendar()
}

func (s *Scheduler) runDaily() {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, now.Location())
		time.Sleep(time.Until(next))
		s.syncCalendar()
	}
}

func (s *Scheduler) syncCalendar() {
	log.Println("Sincronizando calendário...")
	matches, err := s.football.FetchCalendar()
	if err != nil {
		log.Println("Erro no sync:", err)
		return
	}
	for _, m := range matches {
		if err := s.db.UpsertMatch(m); err != nil {
			log.Println("Erro ao salvar jogo:", err)
		}
	}
	s.db.LogSync("calendar")
	log.Printf("Calendário sync: %d jogos\n", len(matches))
}

func (s *Scheduler) watchMatches() {
	for {
		time.Sleep(1 * time.Minute)
		matches, err := s.db.GetTodayMatches()
		if err != nil {
			continue
		}
		for _, m := range matches {
			if time.Now().After(m.StartsAt) && m.Status != "finished" {
				s.mu.Lock()
				already := s.polling[m.ID]
				s.mu.Unlock()
				if !already {
					go s.pollMatch(m)
				}
			}
		}
	}
}

func (s *Scheduler) pollMatch(m *models.Match) {
	s.mu.Lock()
	s.polling[m.ID] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.polling, m.ID)
		s.mu.Unlock()
	}()

	log.Printf("Polling jogo %d: %s\n", m.ID, m.Label())

	for {
		result, err := s.football.FetchMatch(m.ID)
		if err != nil {
			log.Printf("Erro ao buscar jogo %d: %v\n", m.ID, err)
			time.Sleep(5 * time.Minute)
			continue
		}

		home, away := 0, 0
		if result.HomeScore != nil {
			home = *result.HomeScore
		}
		if result.AwayScore != nil {
			away = *result.AwayScore
		}

		s.db.UpdateMatchScore(m.ID, home, away, result.Status)
		s.db.LogSync("scores")

		if result.Status == "finished" {
			log.Printf("Jogo %d finalizado: %d×%d\n", m.ID, home, away)
			s.calcPoints(result)
			return
		}

		time.Sleep(5 * time.Minute)
	}
}

func (s *Scheduler) calcPoints(match *models.Match) {
	guesses, err := s.db.GetMatchGuesses(match.ID)
	if err != nil {
		log.Println("Erro ao buscar palpites:", err)
		return
	}

	usernames := map[string]string{}
	for _, g := range guesses {
		pts := models.CalcPoints(g, match)
		s.db.SetGuessPoints(g.ID, pts)
		g.Points = pts
	}

	if s.onFinished != nil {
		s.onFinished(match, guesses, usernames)
	}
}
