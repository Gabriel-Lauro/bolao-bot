package bot

import (
	"bolao-bot/models"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Name
	switch name {
	case "jogos":
		b.cmdJogos(s, i)
	case "palpite":
		b.cmdPalpite(s, i)
	case "meus-palpites":
		b.cmdMeusPalpites(s, i)
	case "ranking":
		b.cmdRanking(s, i)
	case "sync":
		b.cmdSync(s, i)
	}
}

func (b *Bot) handleModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	if strings.HasPrefix(data.CustomID, "palpite_modal_") {
		b.submitPalpite(s, i, data)
	}
}

func (b *Bot) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	if strings.HasPrefix(data.CustomID, "select_jogo_") {
		b.openPalpiteModal(s, i, data.Values[0])
	}
}

// ── /jogos ─────────────────────────────────────────────────────────────────

func (b *Bot) cmdJogos(s *discordgo.Session, i *discordgo.InteractionCreate) {
	matches, err := b.db.GetOpenMatches()
	if err != nil || len(matches) == 0 {
		respond(s, i, "Nenhum jogo aberto para palpite no momento.", true)
		return
	}

	respondEmbed(s, i, embedJogos(matches), false)
}

// ── /palpite ───────────────────────────────────────────────────────────────

func (b *Bot) cmdPalpite(s *discordgo.Session, i *discordgo.InteractionCreate) {
	matches, err := b.db.GetOpenMatches()
	if err != nil || len(matches) == 0 {
		respond(s, i, "Nenhum jogo aberto para palpite no momento.", true)
		return
	}

	// Garante usuário no banco
	uid := userID(i)
	uname := username(i)
	b.db.UpsertUser(&models.User{ID: uid, Username: uname})

	// Monta select menu com os jogos disponíveis
	options := []discordgo.SelectMenuOption{}
	for _, m := range matches {
		t := m.StartsAt.In(time.FixedZone("BRT", -3*3600))
		label := m.Label()
		desc := t.Format("02/01 às 15h04")
		options = append(options, discordgo.SelectMenuOption{
			Label:       label,
			Value:       strconv.Itoa(m.ID),
			Description: desc,
			Emoji:       &discordgo.ComponentEmoji{Name: "⚽"},
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "**Escolha o jogo para palpitar:**",
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "select_jogo_palpite",
							Placeholder: "Selecione o jogo...",
							Options:     options,
						},
					},
				},
			},
		},
	})
}

func (b *Bot) openPalpiteModal(s *discordgo.Session, i *discordgo.InteractionCreate, matchIDStr string) {
	matchID, err := strconv.Atoi(matchIDStr)
	if err != nil {
		respond(s, i, "Jogo inválido.", true)
		return
	}

	match, err := b.db.GetMatch(matchID)
	if err != nil || match == nil {
		respond(s, i, "Jogo não encontrado.", true)
		return
	}

	if !match.IsOpen() {
		respond(s, i, "❌ Prazo encerrado para este jogo.", true)
		return
	}

	// Pega palpite existente se tiver
	uid := userID(i)
	existing, _ := b.db.GetUserGuessesWithMatches(uid)
	homeVal, awayVal := "0", "0"
	for _, gm := range existing {
		if gm.Guess.MatchID == matchID {
			homeVal = strconv.Itoa(gm.Guess.HomeScore)
			awayVal = strconv.Itoa(gm.Guess.AwayScore)
			break
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("palpite_modal_%d", matchID),
			Title:    "🎯 " + match.Label(),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "home_score",
							Label:       "Gols — " + match.HomeTeam,
							Style:       discordgo.TextInputShort,
							Placeholder: "0",
							Value:       homeVal,
							Required:    true,
							MinLength:   1,
							MaxLength:   2,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "away_score",
							Label:       "Gols — " + match.AwayTeam,
							Style:       discordgo.TextInputShort,
							Placeholder: "0",
							Value:       awayVal,
							Required:    true,
							MinLength:   1,
							MaxLength:   2,
						},
					},
				},
			},
		},
	})
}

func (b *Bot) submitPalpite(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ModalSubmitInteractionData) {
	parts := strings.Split(data.CustomID, "_")
	matchID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		respond(s, i, "Erro interno.", true)
		return
	}

	match, err := b.db.GetMatch(matchID)
	if err != nil || match == nil || !match.IsOpen() {
		respond(s, i, "❌ Prazo encerrado para este jogo.", true)
		return
	}

	homeStr := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	awayStr := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	home, err1 := strconv.Atoi(strings.TrimSpace(homeStr))
	away, err2 := strconv.Atoi(strings.TrimSpace(awayStr))

	if err1 != nil || err2 != nil || home < 0 || away < 0 || home > 99 || away > 99 {
		respond(s, i, "❌ Placar inválido. Use números de 0 a 99.", true)
		return
	}

	uid := userID(i)
	err = b.db.UpsertGuess(&models.Guess{
		UserID:    uid,
		MatchID:   matchID,
		HomeScore: home,
		AwayScore: away,
	})
	if err != nil {
		respond(s, i, "❌ Erro ao salvar palpite.", true)
		return
	}

	respond(s, i, fmt.Sprintf("✅ Palpite salvo!\n**%s %d × %d %s**",
		match.HomeTeam, home, away, match.AwayTeam), true)
}

// ── /meus-palpites ─────────────────────────────────────────────────────────

func (b *Bot) cmdMeusPalpites(s *discordgo.Session, i *discordgo.InteractionCreate) {
	uid := userID(i)
	guesses, err := b.db.GetUserGuessesWithMatches(uid)
	if err != nil || len(guesses) == 0 {
		respond(s, i, "Você ainda não fez nenhum palpite. Use **/palpite** para começar!", true)
		return
	}

	respondEmbed(s, i, embedMeusPalpites(guesses), true)
}

// ── /ranking ───────────────────────────────────────────────────────────────

func (b *Bot) cmdRanking(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ranking, err := b.db.GetRanking()
	if err != nil || len(ranking) == 0 {
		respond(s, i, "Nenhum palpite registrado ainda.", false)
		return
	}

	respondEmbed(s, i, embedRanking(ranking), false)
}

// ── /sync ──────────────────────────────────────────────────────────────────

func (b *Bot) cmdSync(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !b.isAdmin(userID(i)) {
		respond(s, i, "❌ Sem permissão.", true)
		return
	}

	b.scheduler.ForceSync()
	count := b.db.TodaySyncCount()
	respond(s, i, fmt.Sprintf("✅ Sync iniciado! Requests hoje: **%d**", count), true)
}
