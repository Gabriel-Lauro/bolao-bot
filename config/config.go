package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken        string
	DiscordGuildID      string
	DiscordAdminID      string
	DiscordResultsChID  string
	FootballAPIKey      string
	FootballCompetition string
	DBPath              string
}

var C Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("Sem .env, lendo do ambiente")
	}

	C = Config{
		DiscordToken:        mustGet("DISCORD_TOKEN"),
		DiscordGuildID:      mustGet("DISCORD_GUILD_ID"),
		DiscordAdminID:      mustGet("DISCORD_ADMIN_ID"),
		DiscordResultsChID:  mustGet("DISCORD_RESULTS_CHANNEL_ID"),
		FootballAPIKey:      mustGet("FOOTBALL_API_KEY"),
		FootballCompetition: getOrDefault("FOOTBALL_COMPETITION", "2000"),
		DBPath:              getOrDefault("DB_PATH", "./bolao.db"),
	}
}

func mustGet(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("FATAL: %s não definido no .env", key)
	}
	return v
}

func getOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
