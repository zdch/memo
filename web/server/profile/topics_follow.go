package profile

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/web"
	"github.com/memocash/memo/app/auth"
	"github.com/memocash/memo/app/bitcoin/wallet"
	"github.com/memocash/memo/app/db"
	"github.com/memocash/memo/app/profile"
	"github.com/memocash/memo/app/res"
	"net/http"
	"strings"
)

var topicsFollowingRoute = web.Route{
	Pattern:    res.UrlProfileTopicsFollowing + "/" + urlAddress.UrlPart(),
	Handler: func(r *web.Response) {
		offset := r.Request.GetUrlParameterInt("offset")
		addressString := r.Request.GetUrlNamedQueryVariable(urlAddress.Id)
		address := wallet.GetAddressFromString(addressString)
		pkHash := address.GetScriptAddress()
		var userPkHash []byte
		if auth.IsLoggedIn(r.Session.CookieId) {
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
			userPkHash = key.PkHash
		}

		pf, err := profile.GetProfile(pkHash, userPkHash)
		if err != nil {
			r.Error(jerr.Get("error getting profile for hash", err), http.StatusInternalServerError)
			return
		}
		r.Helper["Profile"] = pf
		topics, err := db.GetUniqueTopics(uint(offset), "", pf.PkHash, db.TopicOrderTypeRecent)
		if err != nil {
			r.Error(jerr.Get("error setting following for hash", err), http.StatusInternalServerError)
			return
		}
		r.Helper["Topics"] = topics
		r.Helper["OffsetLink"] = fmt.Sprintf("%s/%s", strings.TrimLeft(res.UrlProfileTopicsFollowing, "/"), address.GetEncoded())
		res.SetPageAndOffset(r, offset)
		r.RenderTemplate(res.UrlProfileTopicsFollowing)
	},
}
