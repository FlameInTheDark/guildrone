package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/FlameInTheDark/guildrone"
)

// Variables used for command line parameters
var (
	Token string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	// Create a new session
	s, err := guildrone.New(Token)
	if err != nil {
		panic(err)
	}

	// Create a new event handler
	eh := func(s *guildrone.Session, e *guildrone.ChatMessageCreated) {
		if e.Message.Content == "ping" {
			s.ChannelMessageCreate(e.Message.ChannelID, "pong")
		}
	}

	// Register the event handler
	s.AddHandler(eh)

	err = s.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Guilded session.
	s.Close()
}
