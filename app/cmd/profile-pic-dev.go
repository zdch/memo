package cmd

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/spf13/cobra"
	"net/http"
	"regexp"
	"os"
	"io"
	"strconv"
	"os/exec"
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

		file, err := os.Create("profile-pic.jpg")
		if err != nil {
			return jerr.Get("couldn't create image file", err)
		}

//		newImage := image.Resize(160, 0, original_image, resize.Lanczos3)

		_, err = io.Copy(file, response.Body)
		if err != nil {
			return jerr.Get("couldn't save image file", err)
		}
		file.Close()

		err = resizeExternally("profile-pic.jpg", "profile-pic-resized.jpg", 200,200)
		if err != nil {
			return jerr.Get("couldn't resize image file", err)
		}

		fmt.Println("success!")
		return nil
	},
}

func resizeExternally(from string, to string, width uint, height uint) error {
	var args = []string{
		"--size", strconv.FormatUint(uint64(width), 10) + "x" +
			strconv.FormatUint(uint64(height), 10),
		"--output", to,
		"--crop",
		from,
	}
	path, err := exec.LookPath("vipsthumbnail.exe")
	if err != nil {
		return err
	}
	cmd := exec.Command(path, args...)
	return cmd.Run()
}