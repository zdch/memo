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

		pic_file_name := "profile-pic"
		file, err := os.Create(pic_file_name + ".jpg")
		if err != nil {
			return jerr.Get("couldn't create image file", err)
		}

		_, err = io.Copy(file, response.Body)
		if err != nil {
			return jerr.Get("couldn't save image file", err)
		}
		file.Close()

		err = resizeExternally(pic_file_name + ".jpg", pic_file_name + "-200x200.jpg", 200,200)
		if err != nil {
			return jerr.Get("couldn't resize image file", err)
		}
		err = resizeExternally(pic_file_name + ".jpg", pic_file_name + "-75x75.jpg", 75,75)
		if err != nil {
			return jerr.Get("couldn't resize image file", err)
		}
		err = resizeExternally(pic_file_name + ".jpg", pic_file_name + "-32x32.jpg", 32,32)
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