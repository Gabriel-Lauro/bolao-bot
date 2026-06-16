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
	totalPts, totalExatos, totalVenc, total := 0, 0, 0, len(guesses)

	for _, gm := range guesses {
		g := gm.Guess
		total = len(guesses)

		if g.Calculated {
			totalPts += g.Points
			if g.Points == 3 {
				totalExatos++
			} else if g.Points == 1 {
				totalVenc++
			}

			var pts string
			switch g.Points {
			case 3:
				pts = "🎯 +3pts"
			case 1:
				pts = "✅ +1pt"
			default:
				pts = "❌ 0pts"
			}
			line := fmt.Sprintf("**%s × %s** — palpite: %d×%d %s",
				gm.HomeTeam, gm.AwayTeam, g.HomeScore, g.AwayScore, pts)
			done = append(done, line)
		} else {
			extra := ""
			if gm.Status == "live" {
				extra = " 🔴"
			}
			line := fmt.Sprintf("**%s × %s** — palpite: %d×%d%s",
				gm.HomeTeam, gm.AwayTeam, g.HomeScore, g.AwayScore, extra)
			pending = append(pending, line)
		}
	}

	resumo := fmt.Sprintf("📊 %d apostas · %dpts · ✅ %d · 🎯 %d",
		total, totalPts, totalVenc, totalExatos)

	var fields []*discordgo.MessageEmbedField

	// Pendentes primeiro (todos)
	if len(pending) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "⏳ Pendente:",
			Value: strings.Join(pending, "\n"),
		})
	}

	// Últimos resultados — prioriza os mais recentes, máx 15 - len(pending)
	maxDone := 15 - len(pending)
	if maxDone < 5 {
		maxDone = 5
	}
	if len(done) > maxDone {
		done = done[len(done)-maxDone:]
	}

	if len(done) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "📊 Últimos resultados:",
			Value: strings.Join(done, "\n"),
		})
	}

	return &discordgo.MessageEmbed{
		Title:       "🎯 Seus palpites",
		Description: resumo,
		Color:       0x5865F2,
		Fields:      fields,
		Footer:      &discordgo.MessageEmbedFooter{Text: "Placar exato = 3pts · Vencedor certo = 1pt"},
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
			"%s **%s** — %d pts (🎯 %d · ✅ %d)",
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
