package pic

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/config"
	"github.com/memocash/memo/app/res"
	"github.com/memocash/memo/app/util"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	ResizeLg  = 640
	ResizeMed = 128
	ResizeSm  = 24
)

// Call when a profile pic doesn't exist on the file system.
func FetchProfilePic(url string, address string) error {
	if ! util.ValidateImgurDirectLink(url) {
		return jerr.New("invalid imgur link")
	}
	response, err := http.Get(url)
	if err != nil {
		return jerr.Get("couldn't fetch remote image", err)
	}
	defer response.Body.Close()

	if _, err := os.Stat(res.PicPath); os.IsNotExist(err) {
		err = os.Mkdir(res.PicPath, 0755)
		if err != nil {
			return jerr.Get("unable to create pic path", err)
		}
	}
	var fileEnding = "jpg"
	if strings.HasSuffix(url, "png") {
		fileEnding = "png"
	}
	profilePicName := res.PicPath + address
	file, err := os.Create(profilePicName + "." + fileEnding)
	if err != nil {
		return jerr.Get("couldn't create image file", err)
	}

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return jerr.Get("couldn't save image file", err)
	}
	err = file.Close()
	if err != nil {
		return jerr.Get("error closing file", err)
	}

	// Resize. vipsthumbnail (super fast) integration is off by default.
	if !config.GetFilePaths().UseVipsThumbnail {
		file, err := os.Open(profilePicName + "." + fileEnding)
		if err != nil {
			return jerr.Get("couldn't open fetched profile pic", err)
		}

		// Decode jpeg into image.Image.
		var img image.Image
		if fileEnding == "jpg" {
			img, err = jpeg.Decode(file)
			if err != nil {
				return jerr.Get("couldn't decode jpg profile pic", err)
			}
		} else {
			img, err = png.Decode(file)
			if err != nil {
				return jerr.Get("couldn't decode png profile pic", err)
			}
		}

		widths := []int{ResizeSm, ResizeMed, ResizeLg}
		for _, width := range widths {

			// Some square crop handling.
			ratio := float32(img.Bounds().Max.X) / float32(img.Bounds().Max.Y)
			ratioY := float32(img.Bounds().Max.Y) / float32(img.Bounds().Max.X)
			if ratioY > ratio {
				ratio = ratioY
			}
			resizeWidth := uint(float32(width) * ratio)

			// Resize to resizeWidth using Lanczos resampling and preserve aspect ratio.
			resizedImg := resize.Resize(resizeWidth, 0, img, resize.Lanczos3)

			croppedImg, err := cutter.Crop(resizedImg, cutter.Config{
				Width:  width,
				Height: width,
				Mode:   cutter.Centered,
			})
			if err != nil {
				return jerr.Get("error cropping image", err)
			}

			out, err := os.Create(profilePicName + "-" + strconv.Itoa(width) + "x" + strconv.Itoa(width) + "." + fileEnding)
			if err != nil {
				return jerr.Get("couldn't create profile pic file", err)
			}

			// Write new image to file.
			if fileEnding == "jpg" {
				err = jpeg.Encode(out, croppedImg, nil)
				if err != nil {
					return jerr.Get("error encoding cropped image", err)
				}
			} else {
				err = png.Encode(out, croppedImg)
				if err != nil {
					return jerr.Get("error encoding cropped image", err)
				}
			}
			err = out.Close()
			if err != nil {
				return jerr.Get("error saving cropped image", err)
			}
		}
	} else {
		err = ResizeExternally(profilePicName+"."+fileEnding, profilePicName+"-"+strconv.Itoa(ResizeSm)+"x"+strconv.Itoa(ResizeSm)+"."+fileEnding, ResizeSm, ResizeSm)
		if err != nil {
			return jerr.Get("couldn't resize image file", err)
		}
		err = ResizeExternally(profilePicName+"."+fileEnding, profilePicName+"-"+strconv.Itoa(ResizeMed)+"x"+strconv.Itoa(ResizeMed)+"."+fileEnding, ResizeMed, ResizeMed)
		if err != nil {
			return jerr.Get("couldn't resize image file", err)
		}
		err = ResizeExternally(profilePicName+"."+fileEnding, profilePicName+"-"+strconv.Itoa(ResizeLg)+"x"+strconv.Itoa(ResizeLg)+"."+fileEnding, ResizeLg, ResizeLg)
		if err != nil {
			return jerr.Get("couldn't resize image file", err)
		}
	}
	err = os.Remove(profilePicName + "." + fileEnding)
	if err != nil {
		return jerr.Get("error removing profile pic", err)
	}

	return nil
}
