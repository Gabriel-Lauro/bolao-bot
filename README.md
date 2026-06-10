# ⚽ bolao-bot

Bot do Discord para bolão da Copa do Mundo. Login via Discord, palpites de placar, ranking automático e resultado postado no canal quando o jogo termina.

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go)
![Discord](https://img.shields.io/badge/Discord-Bot-5865F2?style=flat&logo=discord)
![SQLite](https://img.shields.io/badge/SQLite-3-003B57?style=flat&logo=sqlite)
![License](https://img.shields.io/badge/license-MIT-green?style=flat)

---

## Como funciona

- Jogos sincronizados automaticamente via [football-data.org](https://football-data.org)
- Palpites aceitos até o horário de início do jogo
- Scheduler monitora jogos ao vivo e calcula pontos automaticamente quando termina
- Resultado postado em tempo real no canal configurado

```
Placar exato  → 3 pts
Vencedor certo → 1 pt
Errou          → 0 pts
```

---

## Comandos

| Comando | Descrição | Visível |
|---|---|---|
| `/jogos` | Lista os próximos jogos abertos para palpite | Público |
| `/palpite` | Envia ou edita seu palpite | Privado |
| `/meus-palpites` | Seus palpites e pontuação | Privado |
| `/ranking` | Ranking geral do bolão | Público |
| `/sync` | Força sync com a API *(admin only)* | Privado |

---

## Stack

- **[Go 1.22](https://go.dev)** — binário único, baixo consumo de memória
- **[discordgo](https://github.com/bwmarrin/discordgo)** — WebSocket com o Discord
- **[SQLite](https://modernc.org/sqlite)** — banco local sem dependências externas
- **[football-data.org](https://football-data.org)** — API de jogos (~30 requests/dia no pico)

---

## Requisitos

- Go 1.22+
- Conta no [Discord Developer Portal](https://discord.com/developers/applications)
- API key gratuita em [football-data.org](https://www.football-data.org/client/register)

---

## Instalação

```bash
git clone https://github.com/seu-usuario/bolao-bot
cd bolao-bot
go mod tidy
```

Copia e preenche o `.env`:

```env
DISCORD_TOKEN=
DISCORD_GUILD_ID=
DISCORD_ADMIN_ID=
DISCORD_RESULTS_CHANNEL_ID=
FOOTBALL_API_KEY=
FOOTBALL_COMPETITION=2000
DB_PATH=./bolao.db
```

## Estrutura

```
bolao-bot/
├── main.go
├── config/        # lê variáveis de ambiente
├── db/            # SQLite — queries e schema
├── models/        # structs e lógica de pontuação
├── football/      # cliente da API + traduções PT-BR
├── scheduler/     # sync diário + polling de jogos ao vivo
└── bot/           # comandos, handlers e embeds do Discord
```

## Licença

MIT
