package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/mplewis/figyr"

	_ "github.com/joho/godotenv/autoload"
)

const desc = "Copy emoji to your Discord server."
const discordEmojiURL = "https://cdn.discordapp.com/emojis/%s.png"

type Config struct {
	DiscordBotToken string `figyr:"required,description=Your bot's Discord API token"`
	BindChannelName string `figyr:"default=mojikopi,description=The name of the channel in which this bot will listen"`
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var c Config
	figyr.New(desc).MustParse(&c)

	dg, err := discordgo.New(fmt.Sprintf("Bot %s", c.DiscordBotToken))
	check(err)
	dg.AddHandler(buildListener(c))
	check(dg.Open())
	fmt.Println("Bot is online.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	dg.Close()
}

func buildListener(cfg Config) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	listen := func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		if len(m.Content) == 0 {
			return
		}
		c, err := s.Channel(m.ChannelID)
		if err != nil {
			fmt.Println(err)
			return
		}
		if c.Name != cfg.BindChannelName {
			return
		}

		emojis := m.GetCustomEmojis()
		if len(emojis) == 0 {
			s.ChannelMessageSend(m.ChannelID, "Sorry, I didn't find any custom emojis in your message.")
			return
		}
		for _, emoji := range emojis {
			fmt.Println(emoji)
			e, err := copyEmojiByID(s, m.GuildID, emoji.ID, emoji.Name)
			if err != nil {
				fmt.Println(err)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sorry, I couldn't copy `%s`:\n```\n%s\n```", emoji.Name, err))
				continue
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Copied `%s` to this server: <:%s:%s>", emoji.Name, e.Name, e.ID))
		}
	}
	return listen
}

func copyEmojiByID(s *discordgo.Session, guildID string, id string, name string) (*discordgo.Emoji, error) {
	url := fmt.Sprintf(discordEmojiURL, id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to HTTP GET emoji data: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status on GET emoji data was not 200: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read emoji data: %w", err)
	}

	encoded := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(data))
	params := discordgo.EmojiParams{Name: name, Image: encoded}
	return s.GuildEmojiCreate(guildID, &params)
}
