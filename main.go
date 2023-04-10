package main

import (
	"fmt"

	"github.com/mplewis/figyr"

	_ "github.com/joho/godotenv/autoload"
)

const desc = "Copy emoji to your Discord server."
const perms = 309237647360 // send messages, create public threads, send messages in threads

type Config struct {
	DiscordBotToken string `figyr:"required,description=Your bot's Discord API token"`
}

func main() {
	var c Config
	figyr.New(desc).MustParse(&c)
	fmt.Println(c)
}
