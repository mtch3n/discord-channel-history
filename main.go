package main

import (
	"encoding/csv"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	token    = "TOKEN"
	serverId = "SERVER ID"
	discord  *discordgo.Session

	exportPath = "dc-history.csv"
)

func init() {
	dc, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating discord session: %s", err)
	}

	discord = dc
}

func main() {
	log.Info("Starting bot")

	guild := findGuild(serverId)
	channel := findChannel("🙋self-introductions", guild)
	if channel == nil {
		log.Fatalf("Channel not found")
	}

	m := fetchAllMessages(channel.ID)
	csvExport(m)
}

func mapToDcMessage(msg *discordgo.Message) []string {
	return []string{
		msg.Content,
		msg.Timestamp.String(),
		msg.Author.ID,
		msg.Author.Username,
	}
}

func getHeaders() []string {
	return []string{
		"content",
		"timestamp",
		"author_id",
		"author_username",
	}
}

func csvExport(messages []*discordgo.Message) {
	// just do overwrite no matter what
	f, err := os.Create(exportPath)
	if err != nil {
		log.Fatalf("Error creating file: %s", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(getHeaders()); err != nil {
		log.Fatalf("Error writing headers: %s", err)
	}

	for _, msg := range messages {
		if err := w.Write(mapToDcMessage(msg)); err != nil {
			log.Fatalf("Error writing record: %s", err)
		}
	}
}

func fetchAllMessages(chanId string) []*discordgo.Message {
	// need optimize for rate limit

	var nextMsgId string
	var data []*discordgo.Message

	messages, err := discord.ChannelMessages(chanId, 1, "", "", "")
	if err != nil {
		log.Fatalf("Error getting messages: %s", err)
	}
	data = append(data, messages...)
	nextMsgId = messages[0].ID

	for {
		if messages, err = discord.ChannelMessages(chanId, 100, nextMsgId, "", ""); err == nil {
			if len(messages) == 0 {
				log.Infof("No more messages to fetch")
				break
			}
			data = append(data, messages...)
			nextMsgId = messages[len(messages)-1].ID
		} else {
			log.Fatalf("Error getting messages: %s", err)
		}
	}

	return data
}

func findGuild(id string) *discordgo.UserGuild {
	allGuilds, _ := discord.UserGuilds(0, "", "", false)
	for _, guild := range allGuilds {
		if guild.ID == id {
			log.Infof("Found guild: %s\n", id)
			return guild
		}
	}
	log.Errorf("Guild not found: %s\n", id)
	return nil
}

func findChannel(name string, guilds *discordgo.UserGuild) *discordgo.Channel {

	channels, _ := discord.GuildChannels(guilds.ID)
	for _, channel := range channels {
		if channel.Name == name {
			log.Infof("Found channel: %s", name)
			return channel
		}
	}
	log.Errorf("Channel not found: %s", name)
	return nil
}
