package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	delete = regexp.MustCompile("\\`\\`\\`go|(\\`\\`\\`)")

	sintaxError = &discordgo.MessageEmbed{
		Title:       "SomethingWrong",
		Description: "check your code",
	}
)

func start() string {
	// opent the settings
	fs, err := os.Open("config.json")
	if err != nil {
		fmt.Println("something is wrong", err)
		return "error"
	}
	defer fs.Close()

	byteValue, _ := ioutil.ReadAll(fs)
	var conf map[string]string
	json.Unmarshal([]byte(byteValue), &conf)
	return "Bot " + conf["token"]

}

func main() {

	dg, err := discordgo.New(start())
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	fmt.Println("bot on ready")

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) > 9 {
		if m.Content[:9] == "$compile " {

			programName := "clientPrograms/" + m.ID + ".go"
			//make the programm
			fs, err := os.Create(programName)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "we cant create the archive")
				log.Println(err)
				return
			}

			//write the programm
			defer fs.Close()
			content := delete.ReplaceAllString(m.Content[9:], "")
			fmt.Println(content)
			_, err = fs.WriteString(content)
			if err != nil {

				s.ChannelMessageSendEmbed(m.ChannelID, sintaxError)
				log.Println(err)
				return
			}

			// compile the programm

			output, err := exec.Command("go", "run", programName).Output()
			outputEmbed := &discordgo.MessageEmbed{
				Title:       "Your programm is fine",
				Description: "```" + string(output) + "```",
			}
			s.ChannelMessageSendEmbed(m.ChannelID, outputEmbed)
			if err != nil {
				s.ChannelMessageSendEmbed(m.ChannelID, sintaxError)
				os.Remove(programName)
			}
			// remove the programm
			err = os.Remove(programName)
			if err != nil {
				fmt.Println("something is wrong", err)
			}
		}
	}
}
