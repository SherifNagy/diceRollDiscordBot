package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	dt      int
	da      int
	token   string
	diceSum int
	dtv     [9]int = [9]int{2, 3, 4, 6, 8, 10, 12, 20, 100}
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

func diceRoll(diceAmount int, dice int) []int {
	v := make([]int, diceAmount)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < diceAmount; i++ {
		v[i] = rand.Intn(dt-1+1) + 1
	}
	return v
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()

}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	diceSum := 0
	memberName := ""
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// check if the members has Nickname or not
	if m.Member.Nick == "" {
		memberName = m.Author.Username
	} else {

		memberName = m.Member.Nick
	}
	fullMessage := string(m.Content)
	diceResults := []string{}
	if strings.HasPrefix(fullMessage, "/roll ") {
		fullText := strings.Split(fullMessage, " ")
		diceCall := strings.Split(fullText[1], "d")
		da, _ = strconv.Atoi(diceCall[0])
		dt, _ = strconv.Atoi(diceCall[1])
		dtvm := make(map[int]bool)
		for i := 0; i < len(dtv); i++ {
			dtvm[dtv[i]] = true
		}
		if da > 20 {
			return
		} else if _, ok := dtvm[dt]; ok {
			diceResult := diceRoll(da, dt)
			for u := range diceResult {
				number := diceResult[u]
				diceSum += number
				text := strconv.Itoa(number)
				diceResults = append(diceResults, text)
			}
			s.ChannelMessageSend(m.ChannelID, memberName+" Roll: "+"["+strings.Join(diceResults, ", ")+"]"+" Results: ["+strconv.Itoa(diceSum)+"]")
		}

	}
}
