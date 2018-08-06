package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line options
var (
	Token    string
	Name     string
	DeleteID string
	ListOnly bool
)

func init() {

	flag.StringVar(&Token, "t", "", "Owner Account Token")
	flag.StringVar(&Name, "n", "", "Name to give App/Bot")
	flag.StringVar(&DeleteID, "d", "", "Application ID to delete")
	flag.BoolVar(&ListOnly, "l", false, "List Applications Only")
	flag.Parse()

	if Token == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {

	var err error

	// Create a new Discord session using the provided login information.
	dg, err := discordgo.New(Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// If -l set, only display a list of existing applications
	// for the given account.
	if ListOnly {

		aps, err := dg.Applications()
		if err != nil {
			fmt.Println("error fetching applications,", err)
			return
		}

		for _, v := range aps {
			fmt.Println("-----------------------------------------------------")
			b, _ := json.MarshalIndent(v, "", " ")
			fmt.Println(string(b))
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

	if Name == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Create a new application.
	ap := &discordgo.Application{}
	ap.Name = Name
	ap, err = dg.ApplicationCreate(ap)
	if err != nil {
		fmt.Println("error creating new application,", err)
		return
	}

	fmt.Printf("Application created successfully:\n")
	b, _ := json.MarshalIndent(ap, "", " ")
	fmt.Println(string(b))

	// Create the bot account under the application we just created
	bot, err := dg.ApplicationBotCreate(ap.ID)
	if err != nil {
		fmt.Println("error creating bot account,", err)
		return
	}

	fmt.Printf("Bot account created successfully.\n")
	b, _ = json.MarshalIndent(bot, "", " ")
	fmt.Println(string(b))

	fmt.Println("Please save the above posted info in a secure place.")
	fmt.Println("You will need that information to login with your bot account.")
}
