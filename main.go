package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"cloud.google.com/go/translate"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
)

var (
	ctx context.Context
	client *translate.Client
	err error
)

func init() {
	ctx = context.Background()
	client, err = translate.NewClient(ctx)
	defer client.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN is not set.")
	}
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatal(err)
	}
	defer dg.Close()

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if strings.HasPrefix(m.Content, "!tr") {
		slice := strings.Split(m.Content, " ")
		mes := strings.Join(slice[1:], " ")
		lang, err := DetectLanguage(mes)
		if err != nil {
			return
		}
		trans := Translate(mes, lang)
		s.ChannelMessageSend(m.ChannelID, trans)
	}
}

func DetectLanguage(mes string) (language.Tag, error) {
	lang, err := client.DetectLanguage(ctx, []string{mes})
	if err != nil {
		log.Fatal(err)
	}
	
	if len(lang) == 0 || len(lang[0]) == 0 {
		return language.English, fmt.Errorf("DetectLanguage return value empty")
	}
	return lang[0][0].Language, nil
}

func Translate(mes string, lang language.Tag) string {
	targetLang := language.English
	option := &translate.Options{
		Source: language.Japanese,
		Format: translate.Text,
	}
	if lang != language.Japanese {
		targetLang = language.Japanese
		option.Source = lang
	}
	res, err := client.Translate(ctx, []string{mes}, targetLang, option)
	if err != nil {
		log.Fatal(err)
	}
	trans := []string{}
	for _, r := range res {
		trans = append(trans, r.Text)
	}
	return strings.Join(trans, "\n")
}
