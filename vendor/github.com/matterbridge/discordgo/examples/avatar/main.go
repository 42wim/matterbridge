package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token      string
	AvatarFile string
	AvatarURL  string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&AvatarFile, "f", "", "Avatar File Name")
	flag.StringVar(&AvatarURL, "u", "", "URL to the avatar image")
	flag.Parse()

	if Token == "" || (AvatarFile == "" && AvatarURL == "") {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {

	// Create a new Discord session using the provided login information.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Declare these here so they can be used in the below two if blocks and
	// still carry over to the end of this function.
	var base64img string
	var contentType string

	// If we're using a URL link for the Avatar
	if AvatarURL != "" {

		resp, err := http.Get(AvatarURL)
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

		contentType = http.DetectContentType(img)
		base64img = base64.StdEncoding.EncodeToString(img)
	}

	// If we're using a local file for the Avatar
	if AvatarFile != "" {
		img, err := ioutil.ReadFile(AvatarFile)
		if err != nil {
			fmt.Println(err)
		}

		contentType = http.DetectContentType(img)
		base64img = base64.StdEncoding.EncodeToString(img)
	}

	// Now lets format our base64 image into the proper format Discord wants
	// and then call UserUpdate to set it as our user's Avatar.
	avatar := fmt.Sprintf("data:%s;base64,%s", contentType, base64img)
	_, err = dg.UserUpdate("", "", "", avatar, "")
	if err != nil {
		fmt.Println(err)
	}
}
