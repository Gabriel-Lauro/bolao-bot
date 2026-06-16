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
	db         *db.DB
	football   *football.Client
	onFinished OnMatchFinished
	polling    map[int]bool
	mu         sync.Mutex
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

// ── Job diário meia-noite ──────────────────────────────────────────────────

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

// ── Watcher: detecta jogos que começaram ──────────────────────────────────

func (s *Scheduler) watchMatches() {
	for {
		time.Sleep(1 * time.Minute)

		matches, err := s.db.GetTodayMatches()
		if err != nil {
			continue
		}

		for _, m := range matches {
			if time.Now().Before(m.StartsAt) {
				continue
			}

			s.mu.Lock()
			already := s.polling[m.ID]
			s.mu.Unlock()

			if !already {
				go s.waitAndPoll(m)
			}
		}
	}
}

// ── Espera 95min e começa a pollar ────────────────────────────────────────

func (s *Scheduler) waitAndPoll(m *models.Match) {
	s.mu.Lock()
	s.polling[m.ID] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.polling, m.ID)
		s.mu.Unlock()
	}()

	// Quanto tempo falta pra completar 95min desde o início
	elapsed := time.Since(m.StartsAt)
	waitFor := 95*time.Minute - elapsed
	if waitFor > 0 {
		log.Printf("Jogo %d (%s) começou. Aguardando %v para iniciar polling...\n",
			m.ID, m.Label(), waitFor.Round(time.Minute))
		time.Sleep(waitFor)
	}

	log.Printf("Iniciando polling do jogo %d: %s\n", m.ID, m.Label())
	s.pollMatch(m)
}

// ── Polling a cada 5min até finished ─────────────────────────────────────

func (s *Scheduler) pollMatch(m *models.Match) {
	for {
		result, err := s.football.FetchMatch(m.ID)
		if err != nil {
			log.Printf("Erro ao buscar jogo %d: %v\n", m.ID, err)
			time.Sleep(5 * time.Minute)
			continue
		}

		// Só atualiza se tiver placar real
		if result.HomeScore != nil && result.AwayScore != nil {
			home := *result.HomeScore
			away := *result.AwayScore
			s.db.UpdateMatchScore(m.ID, home, away, result.Status)
			s.db.LogSync("scores")
			log.Printf("Jogo %d: %s %d×%d [%s]\n", m.ID, m.Label(), home, away, result.Status)
		}

		if result.Status == "finished" {
			// Busca o jogo atualizado do banco pra calcular pontos
			updated, err := s.db.GetMatch(m.ID)
			if err != nil || updated == nil {
				log.Printf("Erro ao buscar jogo %d do banco\n", m.ID)
				return
			}
			log.Printf("Jogo %d finalizado: %d×%d\n", m.ID, *updated.HomeScore, *updated.AwayScore)
			s.calcPoints(updated)
			return
		}

		time.Sleep(5 * time.Minute)
	}
}

// ── Calcula pontos e notifica ─────────────────────────────────────────────

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

	log.Printf("Pontos calculados: %d palpites do jogo %d\n", len(guesses), match.ID)

	if s.onFinished != nil {
		s.onFinished(match, guesses, usernames)
	}
}
