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
		topics, err := db.GetUniqueTopics(uint(offset), searchString, userPkHash, db.TopicOrderTypeRecent)
		if err != nil {
			r.Error(jerr.Get("error getting topics from db", err), http.StatusInternalServerError)
			return
		}
		if len(topics) == 0 {
			allRoute.Handler(r)
			return
		}
		err = db.AttachUnreadToTopics(topics, userPkHash)
		if err != nil {
			r.Error(jerr.Get("error attaching unread to topics", err), http.StatusInternalServerError)
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
		err = cache.SetLastTopicList(r.Session.CookieId, "following")
		if err != nil {
			jerr.Get("error setting last topic list", err).Print()
		}
		r.RenderTemplate(res.UrlTopics)
	},
}
