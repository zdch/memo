package topics

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/web"
	"github.com/memocash/memo/app/auth"
	"github.com/memocash/memo/app/db"
	"github.com/memocash/memo/app/res"
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
