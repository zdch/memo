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

var indexRoute = web.Route{
	Pattern: res.UrlTopics,
	Handler: func(r *web.Response) {
		if auth.IsLoggedIn(r.Session.CookieId) {
			followingRoute.Handler(r)
		} else {
			allRoute.Handler(r)
		}
	},
}

var allRoute = web.Route{
	Pattern: res.UrlTopicsAll,
	Handler: func(r *web.Response) {
		preHandler(r)
		offset := r.Request.GetUrlParameterInt("offset")
		searchString := html_parser.EscapeWithEmojis(r.Request.GetUrlParameter("s"))
		topics, err := db.GetUniqueTopics(uint(offset), searchString, []byte{})
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
			r.Helper["OffsetLink"] = fmt.Sprintf("%s?s=%s", strings.TrimLeft(res.UrlTopicsAll, "/"), searchString)
		} else {
			r.Helper["OffsetLink"] = fmt.Sprintf("%s?", res.UrlTopicsAll)
		}
		r.Helper["TopicPage"] = "all"
		r.RenderTemplate(res.UrlTopics)
	},
}

var followingRoute = web.Route{
	Pattern: res.UrlTopicsFollowing,
	Handler: func(r *web.Response) {
		preHandler(r)
		offset := r.Request.GetUrlParameterInt("offset")
		searchString := html_parser.EscapeWithEmojis(r.Request.GetUrlParameter("s"))
		user, err := auth.GetSessionUser(r.Session.CookieId)
		if err != nil {
			r.Error(jerr.Get("error getting session user", err), http.StatusInternalServerError)
			return
		}
		userPkHash, err := cache.GetUserPkHash(user.Id)
		if err != nil {
			r.Error(jerr.Get("error getting pk hash from cache", err), http.StatusInternalServerError)
			return
		}
		topics, err := db.GetUniqueTopics(uint(offset), searchString, userPkHash)
		if err != nil {
			r.Error(jerr.Get("error getting topics from db", err), http.StatusInternalServerError)
			return
		}
		if len(topics) == 0 {
			allRoute.Handler(r)
			return
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
			r.Helper["OffsetLink"] = fmt.Sprintf("%s?s=%s", strings.TrimLeft(res.UrlTopicsFollowing, "/"), searchString)
		} else {
			r.Helper["OffsetLink"] = fmt.Sprintf("%s?", res.UrlTopicsFollowing)
		}
		r.Helper["TopicPage"] = "following"
		r.RenderTemplate(res.UrlTopics)
	},
}

func setTopicFollowingCount(r *web.Response, userPkHash []byte) error {
	if len(userPkHash) == 0 {
		r.Helper["TopicFollowCount"] = 0
		return nil
	}
	count, err := db.GetMemoTopicFollowCountForUser(userPkHash)
	if err != nil {
		return jerr.Get("error getting topic follow count for user", err)
	}
	r.Helper["TopicFollowCount"] = count
	return nil
}
