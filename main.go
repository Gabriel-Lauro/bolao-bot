package main

import (
	"bolao-bot/bot"
	"bolao-bot/config"
	"bolao-bot/db"
	"bolao-bot/football"
	"bolao-bot/models"
	"bolao-bot/scheduler"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.Load()

	database, err := db.New(config.C.DBPath)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}

	fc := football.New(config.C.FootballAPIKey, config.C.FootballCompetition)

	// Callback chamado quando um jogo termina — publica resultado no canal
	var b *bot.Bot
	onFinished := func(match *models.Match, guesses []*models.Guess, usernames map[string]string) {
		// Busca usernames do banco pra montar o embed
		names := map[string]string{}
		for _, g := range guesses {
			u, err := database.GetUser(g.UserID)
			if err == nil && u != nil {
				names[g.UserID] = u.Username
			}
		}

		if b != nil {
			embed := bot.EmbedResultado(match, guesses, names)
			b.Session().ChannelMessageSendEmbed(config.C.DiscordResultsChID, embed)
		}
	}

	sched := scheduler.New(database, fc, onFinished)
	sched.Start()

	b, err = bot.New(database, sched)
	if err != nil {
		log.Fatal("Erro ao criar bot:", err)
	}

	if err := b.Start(); err != nil {
		log.Fatal("Erro ao iniciar bot:", err)
	}
	defer b.Stop()

	log.Println("✅ Bolão Bot rodando! Pressione Ctrl+C para parar.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Encerrando...")
}
