package topics

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/web"
	"github.com/memocash/memo/app/auth"
	"github.com/memocash/memo/app/db"
	"github.com/memocash/memo/app/obj/topic_followers"
	"github.com/memocash/memo/app/res"
	"net/http"
	"net/url"
	"strings"
)

var followersRoute = web.Route{
	Pattern: res.UrlTopicsFollowers + "/" + urlTopicName.UrlPart(),
	Handler: func(r *web.Response) {
		preHandler(r)
		topicRaw := r.Request.GetUrlNamedQueryVariable(urlTopicName.Id)
		unescaped, err := url.QueryUnescape(topicRaw)
		offset := r.Request.GetUrlParameterInt("offset")
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
		topicFollowers, err := topic_followers.GetFollowersForTopic(unescaped, userPkHash)
		if err != nil {
			r.Error(jerr.Get("error getting followers for topic", err), http.StatusInternalServerError)
			return
		}
		if len(topicFollowers) == 0 {
			r.Error(jerr.Get("no topic followers", err), http.StatusUnprocessableEntity)
			return
		}
		r.Helper["Title"] = fmt.Sprintf("Memo Topic Followers - %s", topicFollowers[0].MemoTopicFollow.Topic)
		r.Helper["Topic"] = topicFollowers[0].MemoTopicFollow.Topic
		r.Helper["TopicFollowers"] = topicFollowers
		res.SetPageAndOffset(r, offset)
		r.Helper["OffsetLink"] = fmt.Sprintf("%s/%s?", strings.TrimLeft(res.UrlTopicsFollowers, "/"), topicFollowers[0].MemoTopicFollow.Topic)
		r.RenderTemplate(res.UrlTopicsFollowers)
	},
}
