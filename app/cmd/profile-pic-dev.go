package cmd

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/spf13/cobra"
	"net/http"
	"regexp"
	"os"
	"io"
)

var profilePic = &cobra.Command{
	Use: "profile-pic",
	RunE: func(c *cobra.Command, args []string) error {

		url := "https://i.imgur.com/OcXUNjc.jpg"
		if len(args) == 1 {
			var url_match, err = regexp.Match(`(^http[s]?://[^\s]*[^.?!,)\s])`, []byte(args[0]))
			if err != nil || url_match == false {
				return jerr.Get("must pass an image url", err)
			}
			url = args[0]
		}

		response, err := http.Get(url)
		if err != nil {
			return jerr.Get("couldn't fetch image", err)
		}
		defer response.Body.Close()

		file, err := os.Create("asdf.jpg")
		if err != nil {
			return jerr.Get("couldn't create image file", err)
		}

		_, err = io.Copy(file, response.Body)
		if err != nil {
			return jerr.Get("couldn't save image file", err)
		}
		file.Close()
		fmt.Println("success!")
		return nil
	},
}