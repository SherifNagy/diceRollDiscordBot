package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	token                                   string
	dt                                      int
	da                                      int
	memberName                              string
	dtv                                     [9]int = [9]int{2, 3, 4, 6, 8, 10, 12, 20, 100}
	notationRegex                                  = regexp.MustCompile("([\\d]+)d([\\d]+)((?:[+\\-\\/*][\\d]+)+)?")
	ErrInvalidNotation                             = errors.New("invalid notation")
	ErrInvalidOperationString                      = errors.New("operations has to start with an operator (+-/*)")
	ErrOnlyDigitsAreSupportedAfterOperation        = errors.New("only digits are supported after an operation")
)

type OperationType int

const (
	Unknown OperationType = iota
	Add
	Subtract
	Divide
	Multiply
)
const stringOperations = "+-/*"

type Operation struct {
	Type   OperationType
	Number int
}

type DiceNotation struct {
	diceAmount int
	diceType   int
	Operations []*Operation
}

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

func executeNotation(notation *DiceNotation) (int, []int) {
	numbers := []int{}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < notation.diceAmount; i++ {
		n := rand.Intn(notation.diceType-1+1) + 1
		numbers = append(numbers, n)
	}

	finalSum := 0
	for _, num := range numbers {
		finalSum += num
	}

	for _, op := range notation.Operations {
		switch op.Type {
		case Add:
			finalSum += op.Number
			break
		case Subtract:
			finalSum -= op.Number
			break
		case Multiply:
			finalSum *= op.Number
			break
		case Divide:
			finalSum /= op.Number
			break
		}
	}

	return finalSum, numbers
}

func parseDiceNotation(notation string) (*DiceNotation, error) {
	if !notationRegex.MatchString(notation) {
		return nil, ErrInvalidNotation
	}

	diceNotation := DiceNotation{
		Operations: []*Operation{},
	}

	parsedNotation := notationRegex.FindStringSubmatch(notation)
	if len(parsedNotation) < 3 {
		return nil, ErrInvalidNotation
	}

	dices, err := strconv.Atoi(parsedNotation[1])
	if err != nil {
		return nil, ErrInvalidNotation
	}
	diceNotation.diceAmount = dices

	sides, err := strconv.Atoi(parsedNotation[2])
	if err != nil {
		return nil, ErrInvalidNotation
	}
	diceNotation.diceType = sides

	if len(parsedNotation) > 3 {
		operationsStr := parsedNotation[3]
		if strings.TrimSpace(operationsStr) == "" {
			return &diceNotation, nil
		}

		firstChar := string(operationsStr[0])
		if !strings.Contains(stringOperations, firstChar) {
			return nil, ErrInvalidOperationString
		}

		operation := &Operation{
			Type: Unknown,
		}
		var tempIntStr string

		for _, char := range operationsStr {
			switch string(char) {
			case "*",
				"/",
				"-",
				"+":
				if operation.Type != Unknown {
					runeInt, err := strconv.Atoi(tempIntStr)
					if err != nil {
						return nil, ErrOnlyDigitsAreSupportedAfterOperation
					}

					operation.Number = runeInt
					diceNotation.Operations = append(diceNotation.Operations, operation)

					operation = &Operation{
						Type: Unknown,
					}
					tempIntStr = ""
				}
			}
			switch string(char) {
			case "+":
				operation.Type = Add
				break
			case "-":
				operation.Type = Subtract
				break
			case "/":
				operation.Type = Divide
				break
			case "*":
				operation.Type = Multiply
				break
			default:
				tempIntStr += string(char)
				break
			}
		}

		if operation.Type != Unknown {
			runeInt, err := strconv.Atoi(tempIntStr)
			if err != nil {
				return nil, ErrOnlyDigitsAreSupportedAfterOperation
			}

			operation.Number = runeInt
			diceNotation.Operations = append(diceNotation.Operations, operation)

			operation = &Operation{
				Type: Unknown,
			}
			tempIntStr = ""
		}
	}

	return &diceNotation, nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	memberName := ""
	fullMessage := strings.ToLower(string(m.Content))
	diceArray := []string{}

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
	// main logic to initiate the bot
	if strings.HasPrefix(fullMessage, "/roll ") {
		fullText := strings.Split(fullMessage, " ")
		diceCall := strings.Split(fullText[1], "d")
		da, _ = strconv.Atoi(diceCall[0])
		if da > 20 {
			return
		}
		if strings.ContainsAny(diceCall[1], "/-*+") {
			for _, stringOperation := range stringOperations {
				if strings.ContainsAny(diceCall[1], "dD") {
					break
				}
				dt, _ := strconv.Atoi(strings.Split(diceCall[1], string(stringOperation))[0])
				// need to get rid of that
				fmt.Println(dt)
			}

		} else {
			dt, _ = strconv.Atoi(diceCall[1])

		}
		dtvm := make(map[int]bool)
		for i := 0; i < len(dtv); i++ {
			dtvm[dtv[i]] = true
		}
		if _, ok := dtvm[dt]; ok {
			parsedNotation, err := parseDiceNotation(fullText[1])
			if err != nil {
				log.Fatal(err)
			}

			finalSum, diceResults := executeNotation(parsedNotation)
			for u := range diceResults {
				number := diceResults[u]
				text := strconv.Itoa(number)
				diceArray = append(diceArray, text)
			}
			s.ChannelMessageSend(m.ChannelID, memberName+" Roll: "+"["+strings.Join(diceArray, ", ")+"]"+" Results: ["+strconv.Itoa(finalSum)+"]")
		}
	}
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
