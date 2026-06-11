package bot

import (
	"bolao-bot/db"
	"bolao-bot/models"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func embedJogos(matches []*models.Match) *discordgo.MessageEmbed {
	var lines []string
	for _, m := range matches {
		t := m.StartsAt.In(time.FixedZone("BRT", -3*3600))
		lines = append(lines, fmt.Sprintf("⚽ **%s** — %s", m.Label(), t.Format("02/01 às 15h04")))
	}
	return &discordgo.MessageEmbed{
		Title:       "🗓️ Jogos abertos para palpite",
		Description: strings.Join(lines, "\n"),
		Color:       0x5865F2,
		Footer:      &discordgo.MessageEmbedFooter{Text: "Use /palpite para enviar seu palpite"},
	}
}

func embedMeusPalpites(guesses []db.GuessWithMatch) *discordgo.MessageEmbed {
	var pending, done []string
	totalPts := 0
	for _, gm := range guesses {
		g := gm.Guess
		label := fmt.Sprintf("**%s** — seu palpite: %d×%d", gm.HomeTeam+" × "+gm.AwayTeam, g.HomeScore, g.AwayScore)
		if g.Calculated {
			totalPts += g.Points
			var status string
			switch g.Points {
			case 3:
				status = "🎯 Placar exato! +3pts"
			case 1:
				status = "✅ Vencedor certo! +1pt"
			default:
				status = "❌ Sem pontos"
			}
			result := ""
			if gm.MatchHomeScore != nil && gm.MatchAwayScore != nil {
				result = fmt.Sprintf(" _(resultado: %d×%d)_", *gm.MatchHomeScore, *gm.MatchAwayScore)
			}
			done = append(done, label+result+"\n└ "+status)
		} else {
			extra := ""
			if gm.Status == "live" {
				extra = " 🔴 AO VIVO"
			}
			pending = append(pending, label+extra)
		}
	}
	var fields []*discordgo.MessageEmbedField
	if len(pending) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "⏳ Aguardando resultado",
			Value: strings.Join(pending, "\n"),
		})
	}
	if len(done) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "📊 Finalizados",
			Value: strings.Join(done, "\n"),
		})
	}
	return &discordgo.MessageEmbed{
		Title:  "🎯 Seus palpites",
		Color:  0x5865F2,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Total: %d pontos", totalPts)},
	}
}

func embedRanking(ranking []*models.RankingEntry) *discordgo.MessageEmbed {
	medals := []string{"🥇", "🥈", "🥉"}
	var lines []string
	for _, e := range ranking {
		var medal string
		if e.Position <= 3 {
			medal = medals[e.Position-1]
		} else {
			medal = fmt.Sprintf("%d.", e.Position)
		}
		lines = append(lines, fmt.Sprintf(
			"%s **%s** — %d pts _(🎯 %d exatos · ✅ %d venc.)_",
			medal, e.Username, e.Total, e.Exact, e.Winner,
		))
	}
	return &discordgo.MessageEmbed{
		Title:       "🏆 Ranking do Bolão Copa 2026",
		Description: strings.Join(lines, "\n"),
		Color:       0xFFD700,
		Footer:      &discordgo.MessageEmbedFooter{Text: "Placar exato = 3pts · Vencedor certo = 1pt"},
	}
}

func EmbedResultado(match *models.Match, guesses []*models.Guess, usernames map[string]string) *discordgo.MessageEmbed {
	home, away := 0, 0
	if match.HomeScore != nil {
		home = *match.HomeScore
	}
	if match.AwayScore != nil {
		away = *match.AwayScore
	}
	var exatos, vencedores, erros []string
	for _, g := range guesses {
		name := usernames[g.UserID]
		if name == "" {
			name = g.UserID
		}
		switch g.Points {
		case 3:
			exatos = append(exatos, "🎯 "+name)
		case 1:
			vencedores = append(vencedores, "✅ "+name)
		default:
			erros = append(erros, "❌ "+name)
		}
	}
	var fields []*discordgo.MessageEmbedField
	if len(exatos) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("🎯 Placar exato (+3pts) — %d pessoa(s)", len(exatos)),
			Value: strings.Join(exatos, "  "),
		})
	}
	if len(vencedores) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("✅ Vencedor certo (+1pt) — %d pessoa(s)", len(vencedores)),
			Value: strings.Join(vencedores, "  "),
		})
	}
	if len(erros) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("❌ Sem pontos — %d pessoa(s)", len(erros)),
			Value: strings.Join(erros, "  "),
		})
	}
	if len(fields) == 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "Nenhum palpite",
			Value: "Ninguém apostou nesse jogo.",
		})
	}
	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🏁 %s %d × %d %s", match.HomeTeam, home, away, match.AwayTeam),
		Description: "Jogo encerrado! Confira os pontos:",
		Color:       0x23A55A,
		Fields:      fields,
		Footer:      &discordgo.MessageEmbedFooter{Text: "Use /ranking para ver a classificação"},
	}
}
