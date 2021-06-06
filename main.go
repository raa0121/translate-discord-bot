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
	"github.com/joho/godotenv"
	"golang.org/x/text/language"
)

var (
	ctx context.Context
	client *translate.Client
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
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
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		return
	}
	var detectedRole *discordgo.Role
	for _, role := range roles {
		if role.Name == "translation" {
			detectedRole = role
			break
		}
	}

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	}
	isSendable := false
	for _, permission := range channel.PermissionOverwrites {
		if permission.ID == detectedRole.ID && permission.Type == discordgo.PermissionOverwriteTypeRole && permission.Allow == 2048 {
			isSendable = true
			break
		}
	}
	if isSendable == false {
		return
	}
	mes := m.Content
	lang, err := DetectLanguage(mes)
	if err != nil {
		return
	}
	trans := Translate(mes, lang)
	s.ChannelMessageSend(m.ChannelID, trans)
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
