package memo

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/web"
	"github.com/memocash/memo/app/auth"
	"github.com/memocash/memo/app/bitcoin/memo"
	"github.com/memocash/memo/app/bitcoin/transaction"
	"github.com/memocash/memo/app/db"
	"github.com/memocash/memo/app/mutex"
	"github.com/memocash/memo/app/res"
	"net/http"
	"regexp"
	"os"
	"io"
	"strconv"
	"os/exec"
	"github.com/memocash/memo/app/config"
	"log"
	"image/jpeg"
	"github.com/nfnt/resize"
)

var setPicRoute = web.Route{
	Pattern:    res.UrlMemoSetProfilePic,
	NeedsLogin: true,
	Handler: func(r *web.Response) {
		user, err := auth.GetSessionUser(r.Session.CookieId)
		if err != nil {
			r.Error(jerr.Get("error getting session user", err), http.StatusInternalServerError)
			return
		}
		key, err := db.GetKeyForUser(user.Id)
		if err != nil {
			r.Error(jerr.Get("error getting key for user", err), http.StatusInternalServerError)
			return
		}
		hasSpendableTxOut, err := db.HasSpendable(key.PkHash)
		if err != nil {
			r.Error(jerr.Get("error getting spendable tx out", err), http.StatusInternalServerError)
			return
		}
		if ! hasSpendableTxOut {
			r.SetRedirect(res.UrlNeedFunds)
			return
		}
		r.Render()
	},
}

var setPicSubmitRoute = web.Route{
	Pattern:     res.UrlMemoSetProfilePicSubmit,
	NeedsLogin:  true,
	CsrfProtect: true,
	Handler: func(r *web.Response) {
		url := r.Request.GetFormValue("url")
		password := r.Request.GetFormValue("password")
		user, err := auth.GetSessionUser(r.Session.CookieId)
		if err != nil {
			r.Error(jerr.Get("error getting session user", err), http.StatusInternalServerError)
			return
		}
		key, err := db.GetKeyForUser(user.Id)
		if err != nil {
			r.Error(jerr.Get("error getting key for user", err), http.StatusInternalServerError)
			return
		}

		privateKey, err := key.GetPrivateKey(password)
		if err != nil {
			r.Error(jerr.Get("error getting private key", err), http.StatusUnauthorized)
			return
		}

		address := key.GetAddress()

		// fetch and save image
		urlMatch, err := regexp.Match(`(^http[s]?://[^\s]*[^.?!,)\s])`, []byte(url))
		if err != nil || urlMatch == false {
			r.Error(jerr.Get("must pass an image url", err), http.StatusUnprocessableEntity)
			return
		}

		response, err := http.Get(url)
		if err != nil {
			r.Error(jerr.Get("couldn't fetch remote image", err), http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		profilePicName := config.GetFilePaths().ProfilePicsPath + address.GetAddress().String()
		file, err := os.Create(profilePicName + ".jpg")
		if err != nil {
			r.Error(jerr.Get("couldn't create image file", err), http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(file, response.Body)
		if err != nil {
			r.Error(jerr.Get("couldn't save image file", err), http.StatusInternalServerError)
			return
		}
		file.Close()

		// Resize. vipsthumbnail (super fast) integration is off by default.
		if !config.GetFilePaths().UseVipsThumbnail {

			file, err := os.Open(profilePicName + ".jpg")
			if err != nil {
				log.Fatal(err)
			}

			// Decode jpeg into image.Image.
			img, err := jpeg.Decode(file)
			if err != nil {
				log.Fatal(err)
			}
			file.Close()

			// Resize to width 75 using Lanczos resampling and preserve aspect ratio.
			m := resize.Resize(75, 0, img, resize.Lanczos3)
			out, err := os.Create(profilePicName + "-75x75.jpg")
			if err != nil {
				log.Fatal(err)
			}
			defer out.Close()

			// Write new image to file.
			jpeg.Encode(out, m, nil)

		} else {
			err = resizeExternally(profilePicName + ".jpg", profilePicName + "-200x200.jpg", 200,200)
			if err != nil {
				r.Error(jerr.Get("couldn't resize image file", err), http.StatusInternalServerError)
				return
			}
			err = resizeExternally(profilePicName + ".jpg", profilePicName + "-75x75.jpg", 75,75)
			if err != nil {
				r.Error(jerr.Get("couldn't resize image file", err), http.StatusInternalServerError)
				return
			}
			err = resizeExternally(profilePicName + ".jpg", profilePicName + "-24x24.jpg", 24,24)
			if err != nil {
				r.Error(jerr.Get("couldn't resize image file", err), http.StatusInternalServerError)
				return
			}
		}
return
		var fee = int64(memo.MaxTxFee - memo.MaxPostSize + len([]byte(url)))
		var minInput = fee + transaction.DustMinimumOutput

		mutex.Lock(key.PkHash)
		txOut, err := db.GetSpendableTxOut(key.PkHash, minInput)
		if err != nil {
			mutex.Unlock(key.PkHash)
			r.Error(jerr.Get("error getting spendable tx out", err), http.StatusInternalServerError)
			return
		}

		tx, err := transaction.Create([]*db.TransactionOut{txOut}, privateKey, []transaction.SpendOutput{{
			Type:    transaction.SpendOutputTypeP2PK,
			Address: address,
			Amount:  txOut.Value - fee,
		}, {
			Type: transaction.SpendOutputTypeMemoSetPic,
			Data: []byte(url),
		}})
		if err != nil {
			mutex.Unlock(key.PkHash)
			r.Error(jerr.Get("error creating tx", err), http.StatusInternalServerError)
			return
		}

		fmt.Println(transaction.GetTxInfo(tx))
		transaction.QueueTx(tx)
		r.Write(tx.TxHash().String())
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
	path, err := exec.LookPath(config.GetFilePaths().VipsThumbnailPath)
	if err != nil {
		return err
	}
	cmd := exec.Command(path, args...)
	return cmd.Run()
}