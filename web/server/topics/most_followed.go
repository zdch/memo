package topics

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/web"
	"github.com/memocash/memo/app/auth"
	"github.com/memocash/memo/app/cache"
	"github.com/memocash/memo/app/db"
	"github.com/memocash/memo/app/html-parser"
	"github.com/memocash/memo/app/res"
	"net/http"
	"strings"
)

var mostFollowedRoute = web.Route{
	Pattern: res.UrlTopicsMostFollowed,
	Handler: func(r *web.Response) {
		preHandler(r)
		offset := r.Request.GetUrlParameterInt("offset")
		searchString := html_parser.EscapeWithEmojis(r.Request.GetUrlParameter("s"))
		topics, err := db.GetUniqueTopics(uint(offset), searchString, []byte{}, db.TopicOrderTypeFollowers)
		if err != nil {
			r.Error(jerr.Get("error getting topics from db", err), http.StatusInternalServerError)
			return
		}
		var userPkHash []byte
		if auth.IsLoggedIn(r.Session.CookieId) {
			user, err := auth.GetSessionUser(r.Session.CookieId)
			if err != nil {
				r.Error(jerr.Get("error getting session user", err), http.StatusInternalServerError)
				return
			}
			userPkHash, err = cache.GetUserPkHash(user.Id)
			if err != nil {
				r.Error(jerr.Get("error getting pk hash from cache", err), http.StatusInternalServerError)
				return
			}
		}
		err = setTopicFollowingCount(r, userPkHash)
		if err != nil {
			r.Error(jerr.Get("error setting topic follow count for user", err), http.StatusInternalServerError)
			return
		}
		r.Helper["Title"] = "Memo Topics"
		r.Helper["Topics"] = topics
		r.Helper["SearchString"] = searchString
		res.SetPageAndOffset(r, offset)
		if searchString != "" {
			r.Helper["OffsetLink"] = fmt.Sprintf("%s?s=%s", strings.TrimLeft(res.UrlTopicsMostFollowed, "/"), searchString)
		} else {
			r.Helper["OffsetLink"] = fmt.Sprintf("%s?", res.UrlTopicsMostFollowed)
		}
		r.Helper["TopicPage"] = "most-followed"
		r.RenderTemplate(res.UrlTopics)
	},
}
