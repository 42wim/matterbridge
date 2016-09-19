package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Email       string
	Password    string
	Token       string
	URL         string
	BotID       string
	BotUsername string
)

func init() {

	flag.StringVar(&Email, "e", "", "Account Email")
	flag.StringVar(&Password, "p", "", "Account Password")
	flag.StringVar(&Token, "t", "", "Account Token")
	flag.StringVar(&URL, "l", "http://bwmarrin.github.io/discordgo/img/discordgo.png", "Link to the avatar image")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided login information.
	// Use discordgo.New(Token) to just use a token for login.
	dg, err := discordgo.New(Email, Password, Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	bot, err := dg.User("@me")
	if err != nil {
		fmt.Println("error fetching the bot details,", err)
		return
	}

	BotID = bot.ID
	BotUsername = bot.Username
	changeAvatar(dg)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
}

// Helper function to change the avatar
func changeAvatar(s *discordgo.Session) {

	resp, err := http.Get(URL)
	if err != nil {
		fmt.Println("Error retrieving the file, ", err)
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	img, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading the response, ", err)
		return
	}

	base64 := base64.StdEncoding.EncodeToString(img)

	avatar := fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(img), base64)

	_, err = s.UserUpdate("", "", BotUsername, avatar, "")
	if err != nil {
		fmt.Println("Error setting the avatar, ", err)
	}

}
