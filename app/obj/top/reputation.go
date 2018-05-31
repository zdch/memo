package top

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/obj/rep"
)

func AttachReputationToTopicFollowers(topicFollowers []*TopicFollower) error {
	for _, topicFollower := range topicFollowers {
		reputation, err := rep.GetReputation(topicFollower.SelfPkHash, topicFollower.MemoTopicFollow.PkHash)
		if err != nil {
			return jerr.Get("error getting reputation", err)
		}
		topicFollower.Reputation = reputation
	}
	return nil
}
