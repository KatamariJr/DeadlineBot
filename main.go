package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

const Day = time.Hour * 24

// viper values
const tokenStr = "discord-token"
const channelIDStr = "channel-id"

// intervals before the deadline at which to echo a message
var intervals = []time.Duration{
	Day * 25,
	Day * 21,
	Day * 14,
	Day * 7,
	Day * 5,
	Day * 3,
	Day,
	time.Hour * 6,
	time.Hour * 1,
}

// text values that will be echoed when the interval passes, mirrors the values in intervals
var intervalsText = []string{
	"25 days",
	"Three weeks",
	"Two weeks",
	"One week",
	"Five days",
	"Three days",
	"One day",
	"Six hours",
	"One hour",
}

func main() {

	//parse the timestamp string
	timestampStr := flag.String("timestamp", "", "")
	flag.Parse()
	timestamp, err := time.Parse(time.RFC822, *timestampStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	setupViper()

	//set up all the intervals in a slice
	now := time.Now()
	intervalTimes := []time.Time{}
	for _, v := range intervals {
		alertTime := timestamp.Add(-v)

		//stop adding times if the interval time has already passed
		if alertTime.Before(now) {
			continue
		}

		intervalTimes = append(intervalTimes, alertTime)
		fmt.Println(alertTime)
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + viper.GetString(tokenStr))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// create a channel to listen on until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	maxIntervals := len(intervalTimes)
	currentInterval := 0

	//create the timer channel which will fire at the next interval
	timerChan := time.After(time.Until(intervalTimes[currentInterval]))

	intervalsText = intervalsText[len(intervalsText)-maxIntervals:]

	channelID := viper.GetString(channelIDStr)

	for true {
		select {
		case _, ok := <-sc:
			if ok {
				// Cleanly close down the Discord session.
				dg.Close()
				os.Exit(0)
			} else {
				fmt.Println("Channel closed!")
			}
		case _, ok := <-timerChan:
			if ok {
				fmt.Println("Time elapsed!")
				_, err = dg.ChannelMessageSend(channelID, fmt.Sprintf("%s remaining until the submission deadline!", intervalsText[currentInterval]))
				if err != nil {
					fmt.Printf("unable to send message to channel %v: %v", channelID, err)
				}
				currentInterval++

				if currentInterval == maxIntervals {
					fmt.Println("Intervals exhausted")
					dg.Close()
					os.Exit(0)
				}

				//set timerChan to the next interval
				timerChan = time.After(time.Until(intervalTimes[currentInterval]))

			}
		default:
			//fmt.Println("No value ready, moving on.")
		}
	}
}

// setupViper will prepare viper for reading values
func setupViper() {
	const configName = "bot"
	var configLocations = []string{
		"/etc/dwellingofduels",
		"$HOME/.dwelling",
		".",
		"./bot",
	}

	//set config name and locations
	viper.SetConfigName(configName)
	viper.SetConfigType("json")
	for _, l := range configLocations {
		viper.AddConfigPath(l)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			//log.Printf("no config file '%s' found, searched the following directories %v", configName, configLocations)
		} else {
			fmt.Printf("Fatal error config file: %s \n", err)
		}
	}
}
