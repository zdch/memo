package memo

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/web"
	"github.com/memocash/memo/app/auth"
	"github.com/memocash/memo/app/bitcoin/transaction"
	"github.com/memocash/memo/app/bitcoin/transaction/build"
	"github.com/memocash/memo/app/db"
	"github.com/memocash/memo/app/mutex"
	"github.com/memocash/memo/app/res"
	"github.com/memocash/memo/app/util"
	"net/http"
	"regexp"
	"strings"
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

// Transform https://imgur.com/xSSV7Sg into https://i.imgur.com/xSSV7Sg.jpg and return the string.
func processImgurUrl(url string) (string, error) {
	// Nothing to do.
	if util.ValidateImgurDirectLink(url) {
		return url, nil
	}

	// Transform to direct link and validate.
	var re = regexp.MustCompile(`(https://([a-z]+\.)?imgur\.com/)([^\s]*)`)
	url = re.ReplaceAllString(url, `https://i.imgur.com/$3.jpg`)

	if util.ValidateImgurDirectLink(url) {
		return url, nil
	} else {
		return "", jerr.New("invalid imgur link")
	}
}

var setPicSubmitRoute = web.Route{
	Pattern:     res.UrlMemoSetProfilePicSubmit,
	NeedsLogin:  true,
	CsrfProtect: true,
	Handler: func(r *web.Response) {
		url := r.Request.GetFormValue("url")
		url, err := processImgurUrl(url)
		if err != nil {
			r.Error(jerr.Get("invalid profile pic url", err), http.StatusInternalServerError)
			return
		}

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
		contentType := response.Header.Get("content-type")
		if contentType == "image/png" && strings.HasSuffix(url, "jpg") {
			url = strings.TrimSuffix(url, "jpg") + "png"
		}
		response.Body.Close()

		//pic.FetchProfilePic(url, address.GetAddress().String())

		pkHash := privateKey.GetPublicKey().GetAddress().GetScriptAddress()
		mutex.Lock(pkHash)

		tx, err := build.ProfilePic(url, privateKey)
		if err != nil {
			mutex.Unlock(pkHash)
			r.Error(jerr.Get("error building like tx", err), http.StatusInternalServerError)
			return
		}

		fmt.Println(transaction.GetTxInfo(tx))
		transaction.QueueTx(tx)
		r.Write(tx.TxHash().String())
	},
}
