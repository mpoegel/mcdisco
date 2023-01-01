package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/hpcloud/tail"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
}

func watchLogFile(filename string, discord *discordgo.Session, channelID string) {
	t, err := tail.TailFile(filename, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
	})
	if err != nil {
		logger.Fatal("tail failed", zap.Error(err))
	}
	logoutMsg := regexp.MustCompile(`.*:\s(\w+) left the game`)
	loginMsg := regexp.MustCompile(`.*:\s(\w+) joined the game`)
	logger.Info("starting tail", zap.String("filename", filename))
	for line := range t.Lines {
		if line.Err != nil {
			logger.Warn("failed to tail line", zap.String("filename", filename), zap.Error(err))
		}
		if matches := loginMsg.FindStringSubmatch(line.Text); len(matches) > 0 {
			msg := fmt.Sprintf("%s joined the game", matches[1])
			logger.Info(msg)
			if _, err = discord.ChannelMessageSend(channelID, msg); err != nil {
				logger.Fatal("failed to send discord message", zap.Error(err))
			}

		}
		if matches := logoutMsg.FindStringSubmatch(line.Text); len(matches) > 0 {
			msg := fmt.Sprintf("%s left the game", matches[1])
			logger.Info(msg)
			if _, err = discord.ChannelMessageSend(channelID, msg); err != nil {
				logger.Fatal("failed to send discord message", zap.Error(err))
			}
		}
	}
	logger.Info("tailing complete", zap.String("filename", filename))
}

func main() {
	defer logger.Sync()

	botToken := os.Getenv("APP_TOKEN")
	if len(botToken) == 0 {
		logger.Fatal("please set APP_TOKEN")
	}

	channelID := os.Getenv("CHANNEL_ID")
	if len(channelID) == 0 {
		logger.Fatal("please set CHANNEL_ID")
	}

	logFileName := flag.String("logfile", "latest.log", "minecraft log file")
	flag.Parse()

	discord, err := discordgo.New(fmt.Sprintf("Bot %s", botToken))
	if err != nil {
		logger.Fatal("could not create discord session", zap.Error(err))
	}
	defer discord.Close()

	watchLogFile(*logFileName, discord, channelID)
}
