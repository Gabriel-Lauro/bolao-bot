package bot

import (
	"bolao-bot/config"
	"log"

	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "jogos",
		Description: "Lista os próximos jogos abertos para palpite",
	},
	{
		Name:        "palpite",
		Description: "Envia ou edita seu palpite para um jogo",
	},
	{
		Name:        "meus-palpites",
		Description: "Veja todos os seus palpites",
	},
	{
		Name:        "ranking",
		Description: "Ranking do bolão",
	},
	{
		Name:        "sync",
		Description: "[Admin] Força sincronização dos jogos com a API",
	},
}

func (b *Bot) registerCommands() {
	existing, err := b.session.ApplicationCommands(b.session.State.User.ID, config.C.DiscordGuildID)
	if err != nil {
		log.Println("Erro ao buscar comandos existentes:", err)
	}

	// Remove comandos antigos que não existem mais
	existingNames := map[string]string{}
	for _, cmd := range existing {
		existingNames[cmd.Name] = cmd.ID
	}

	wantedNames := map[string]bool{}
	for _, cmd := range commands {
		wantedNames[cmd.Name] = true
	}

	for name, id := range existingNames {
		if !wantedNames[name] {
			b.session.ApplicationCommandDelete(b.session.State.User.ID, config.C.DiscordGuildID, id)
		}
	}

	// Registra/atualiza comandos
	for _, cmd := range commands {
		if _, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, config.C.DiscordGuildID, cmd); err != nil {
			log.Printf("Erro ao registrar comando /%s: %v\n", cmd.Name, err)
		} else {
			log.Printf("Comando /%s registrado\n", cmd.Name)
		}
	}
}
