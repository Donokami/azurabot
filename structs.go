package main

import (
	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"os/exec"
	"sync"
)

type Bot struct {
	config         Options
	dg             *discordgo.Session
	voiceInstances map[string]*VoiceInstance
	purgeTime      int64
	purgeQueue     []PurgeMessage
	mutex          sync.Mutex
	radioSignal    chan PkgRadio
	azuracast      AzuraCast
}

type Options struct {
	DiscordToken      string `env:"DISCORD_TOKEN,required=true"`
	DiscordStatus     string `env:"DISCORD_STATUS,default=Ready to play!"`
	DiscordPrefix     string `env:"DISCORD_PREFIX,default=!"`
	DiscordPurgeTime  int64  `env:"DISCORD_PURGETIME,default=60"`
	DiscordPlayStatus bool   `env:"DISCORD_PLAYSTATUS,default=true"`
	DiscordVolume     int	 `env:"DISCORD_VOLUME,default=100"`
	AzuracastUrl      string `env:"AZURACAST_URL,default=http://nginx"`
}

type PurgeMessage struct {
	ID, ChannelID string
	TimeSent      int64
}

type Channel struct {
	db *bolt.DB
}

type PkgRadio struct {
	data string
	v    *VoiceInstance
}

type VoiceInstance struct {
	voice      *discordgo.VoiceConnection
	session    *discordgo.Session
	encoder    *dca.EncodeSession
	stream     *dca.StreamingSession
	run        *exec.Cmd
	queueMutex sync.Mutex
	audioMutex sync.Mutex
	recv       []int16
	guildID    string
	channelID  string
	volume	   int
	is_playing bool
	station    *AzuraCastStation
	np         *AzuraCastNowPlaying
}
