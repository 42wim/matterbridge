package main

import (
	"flag"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line options
var (
	Email    string
	Password string
	Token    string
	AppName  string
	DeleteID string
	ListOnly bool
)

func init() {

	flag.StringVar(&Email, "e", "", "Account Email")
	flag.StringVar(&Password, "p", "", "Account Password")
	flag.StringVar(&Token, "t", "", "Account Token")
	flag.StringVar(&DeleteID, "d", "", "Application ID to delete")
	flag.BoolVar(&ListOnly, "l", false, "List Applications Only")
	flag.StringVar(&AppName, "a", "", "App/Bot Name")
	flag.Parse()
}

func main() {

	var err error
	// Create a new Discord session using the provided login information.
	dg, err := discordgo.New(Email, Password, Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// If -l set, only display a list of existing applications
	// for the given account.
	if ListOnly {
		aps, err2 := dg.Applications()
		if err2 != nil {
			fmt.Println("error fetching applications,", err)
			return
		}

		for k, v := range aps {
			fmt.Printf("%d : --------------------------------------\n", k)
			fmt.Printf("ID: %s\n", v.ID)
			fmt.Printf("Name: %s\n", v.Name)
			fmt.Printf("Secret: %s\n", v.Secret)
			fmt.Printf("Description: %s\n", v.Description)
		}
		return
	}

	// if -d set, delete the given Application
	if DeleteID != "" {
		err = dg.ApplicationDelete(DeleteID)
		if err != nil {
			fmt.Println("error deleting application,", err)
		}
		return
	}

	// Create a new application.
	ap := &discordgo.Application{}
	ap.Name = AppName
	ap, err = dg.ApplicationCreate(ap)
	if err != nil {
		fmt.Println("error creating new applicaiton,", err)
		return
	}

	fmt.Printf("Application created successfully:\n")
	fmt.Printf("ID: %s\n", ap.ID)
	fmt.Printf("Name: %s\n", ap.Name)
	fmt.Printf("Secret: %s\n\n", ap.Secret)

	// Create the bot account under the application we just created
	bot, err := dg.ApplicationBotCreate(ap.ID)
	if err != nil {
		fmt.Println("error creating bot account,", err)
		return
	}

	fmt.Printf("Bot account created successfully.\n")
	fmt.Printf("ID: %s\n", bot.ID)
	fmt.Printf("Username: %s\n", bot.Username)
	fmt.Printf("Token: %s\n\n", bot.Token)
	fmt.Println("Please save the above posted info in a secure place.")
	fmt.Println("You will need that information to login with your bot account.")

	return
}
