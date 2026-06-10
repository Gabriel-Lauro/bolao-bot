package bot

import (
	"bolao-bot/config"
	"bolao-bot/db"
	"bolao-bot/scheduler"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session   *discordgo.Session
	db        *db.DB
	scheduler *scheduler.Scheduler
}

func New(database *db.DB, sched *scheduler.Scheduler) (*Bot, error) {
	s, err := discordgo.New("Bot " + config.C.DiscordToken)
	if err != nil {
		return nil, err
	}

	b := &Bot{
		session:   s,
		db:        database,
		scheduler: sched,
	}

	s.AddHandler(b.onInteraction)
	s.Identify.Intents = discordgo.IntentsGuilds

	return b, nil
}

func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return err
	}
	log.Println("Bot conectado ao Discord")
	b.registerCommands()
	return nil
}

func (b *Bot) Stop() {
	b.session.Close()
}

func (b *Bot) Session() *discordgo.Session {
	return b.session
}

func (b *Bot) onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		b.handleCommand(s, i)
	} else if i.Type == discordgo.InteractionModalSubmit {
		b.handleModal(s, i)
	} else if i.Type == discordgo.InteractionMessageComponent {
		b.handleComponent(s, i)
	}
}

func (b *Bot) isAdmin(userID string) bool {
	return userID == config.C.DiscordAdminID
}

func userID(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		return i.Member.User.ID
	}
	return i.User.ID
}

func username(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		if i.Member.Nick != "" {
			return i.Member.Nick
		}
		return i.Member.User.Username
	}
	return i.User.Username
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string, ephemeral bool) {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   flags,
		},
	})
}

func respondEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, ephemeral bool) {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  flags,
		},
	})
}
