package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
	"strings"
	"time"
)

// HelpReporter
func (b *Bot) HelpReporter(m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "sent command 'help'")
	help := "```go\n`Standard Commands List`\n```\n" +
		"**`" + o.DiscordPrefix + "help`** ->  show help commands.\n" +
		"**`" + o.DiscordPrefix + "join`** ->  join a voice channel.\n" +
		"**`" + o.DiscordPrefix + "leave`** ->  leave a voice channel.\n" +
		"**`" + o.DiscordPrefix + "play [station_short_name]`** ->  play the specified station.\n" +
		"**`" + o.DiscordPrefix + "stop`**  ->  stop the player.\n" +
		"**`" + o.DiscordPrefix + "np`**  ->  show what's now playing.\n" +
		"**`" + o.DiscordPrefix + "vol [1-100]`**  -> set the player volume.\n" +
		"```go\n`Owner Commands List`\n```\n" +
		"**`" + o.DiscordPrefix + "ignore`**  ->  ignore commands of a channel.\n" +
		"**`" + o.DiscordPrefix + "unignore`**  ->  unignore commands of a channel.\n"

	b.ChMessageSend(m.ChannelID, help)
}

// JoinReporter
func JoinReporter(v *VoiceInstance, m *discordgo.MessageCreate, s *discordgo.Session) {
	log.Println("INFO:", m.Author.Username, "send 'join'")
	voiceChannelID := SearchVoiceChannel(m.Author.ID)
	if voiceChannelID == "" {
		log.Println("ERROR: Voice channel id not found.")
		ChMessageSend(m.ChannelID, "[**Music**] <@"+m.Author.ID+"> You need to join a voice channel!")
		return
	}
	if v != nil {
		log.Println("INFO: Voice Instance already created.")
	} else {
		guildID := SearchGuild(m.ChannelID)
		// create new voice instance
		mutex.Lock()
		v = new(VoiceInstance)
		voiceInstances[guildID] = v
		v.guildID = guildID
		v.session = s
		mutex.Unlock()
		//v.InitVoice()
	}
	var err error
	v.voice, err = dg.ChannelVoiceJoin(v.guildID, voiceChannelID, false, false)
	if err != nil {
		v.Stop()
		log.Println("ERROR: Error to join in a voice channel: ", err)
		return
	}
	v.voice.Speaking(false)
	log.Println("INFO: New Voice Instance created")
	ChMessageSend(m.ChannelID, "[**Music**] I've joined a voice channel!")
}

// LeaveReporter
func LeaveReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'leave'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		return
	}
	v.Stop()
	time.Sleep(200 * time.Millisecond)
	v.voice.Disconnect()
	log.Println("INFO: Voice channel destroyed")
	mutex.Lock()
	delete(voiceInstances, v.guildID)
	mutex.Unlock()
	dg.UpdateStatus(0, o.DiscordStatus)
	ChMessageSend(m.ChannelID, "[**Music**] I left the voice channel!")
}

// PlayReporter
func PlayReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "sent command 'play'")

	if v == nil {
		log.Println("INFO: The bot is not in a voice channel.")
		ChMessageSend(m.ChannelID, "Join a voice channel before playing music.")
		return
	}

	if len(strings.Fields(m.Content)) > 1 {
		err := azuracast.GetNowPlaying(v, strings.Fields(m.Content)[1])
		if err != nil {
			ChMessageSend(m.ChannelID, "Error: Could not retrieve station information.")
			log.Println("ERROR: AzuraCast API call returned", err.Error())
			return
		}
	} else if v.station == nil {
		ChMessageSend(m.ChannelID, "You must specify a station ID number or shortcode after the command the first time you play.")
		return
	}

	radio := PkgRadio{
		data: v.station.ListenURL,
		v:    v,
	}

	go func() {
		radioSignal <- radio
	}()

	ChMessageSend(m.ChannelID, "Starting playback of "+v.station.Name+"!")
}

// ReadioReporter
func RadioReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'radio'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
		return
	}
	if len(strings.Fields(m.Content)) < 2 {
		ChMessageSend(m.ChannelID, "[**Music**] You need to specify a url!")
		return
	}
	radio := PkgRadio{"", v}
	radio.data = strings.Fields(m.Content)[1]

	go func() {
		radioSignal <- radio
	}()
	ChMessageSend(m.ChannelID, "[**Music**] **`"+m.Author.Username+"`** I'm playing a radio now!")
}

// StopReporter
func StopReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'stop'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
		return
	}
	voiceChannelID := SearchVoiceChannel(m.Author.ID)
	if v.voice.ChannelID != voiceChannelID {
		ChMessageSend(m.ChannelID, "[**Music**] <@"+m.Author.ID+"> You need to join in my voice channel for send stop!")
		return
	}
	v.Stop()
	dg.UpdateStatus(0, o.DiscordStatus)
	log.Println("INFO: The bot stop play audio")
	ChMessageSend(m.ChannelID, "[**Music**] I'm stoped now!")
}

// PauseReporter
func PauseReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'pause'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		return
	}

	v.Stop()
	dg.UpdateStatus(0, config.DiscordStatus)

	log.Println("INFO: The bot stopped playing audio")
	ChMessageSend(m.ChannelID, "Stopping radio playback...")
}

// Return Now Playing information
func NowPlayingReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "sent command 'np'")

	if v == nil {
		log.Println("INFO: The bot is not in a voice channel.")
		ChMessageSend(m.ChannelID, "The bot is not currently active.")
		return
	}

	err := azuracast.UpdateNowPlaying(v)
	if err != nil {
		ChMessageSend(m.ChannelID, "Error: Could not retrieve now playing information.")
		log.Println("ERROR: AzuraCast API call returned", err.Error())
		return
	}

	ChMessageSend(m.ChannelID, "**Now Playing on "+v.station.Name+":** "+v.np.Title+" by "+v.np.Artist)
}

func VolumeReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "sent command 'vol'")

	if v == nil {
		log.Println("INFO: The bot is not in a voice channel.")
		ChMessageSend(m.ChannelID, "Join a voice channel before setting the volume.")
		return
	}

	if len(strings.Fields(m.Content)) > 1 {
		volPercent, err := strconv.Atoi(strings.Fields(m.Content)[1])

		if volPercent > 100 {
			volPercent = 100
		} else if volPercent < 1 {
			volPercent = 1
		}

		if err != nil {
			ChMessageSend(m.ChannelID, "Error: Could not set volume.")
			log.Println("ERROR: Volume parse error", err.Error())
		}

		v.volume = int(math.Ceil(float64(volPercent) * 255.0/100.0))

		log.Println("INFO: Volume set to ", v.volume, " from user input ", volPercent)

		ChMessageSend(m.ChannelID, "Volume updated to "+strconv.Itoa(volPercent)+"%.")
	} else {
		v.volume = 0
		ChMessageSend(m.ChannelID, "Volume reset to default.")
	}

	if v.is_playing && v.station != nil {
		radio := PkgRadio{
			data: v.station.ListenURL,
			v:    v,
		}

		go func() {
			radioSignal <- radio
		}()
	}


}
