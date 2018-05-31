package top

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/bitcoin/wallet"
	"github.com/memocash/memo/app/db"
	"github.com/memocash/memo/app/obj/rep"
)

type TopicFollower struct {
	Name            string
	SelfPkHash      []byte
	Reputation      *rep.Reputation
	MemoTopicFollow *db.MemoTopicFollow
}

func (t TopicFollower) GetAddressString() string {
	return t.GetAddress().GetEncoded()
}

func (t TopicFollower) GetAddress() wallet.Address {
	return wallet.GetAddressFromPkHash(t.MemoTopicFollow.PkHash)
}

func GetFollowersForTopic(topic string, selfPkHash []byte) ([]*TopicFollower, error) {
	dbFollowers, err := db.GetFollowersForTopic(topic)
	if err != nil {
		return nil, jerr.Get("error attaching followers for topic", err)
	}
	var topicFollowers []*TopicFollower
	for _, dbFollower := range dbFollowers {
		topicFollowers = append(topicFollowers, &TopicFollower{
			MemoTopicFollow: dbFollower,
			SelfPkHash:      selfPkHash,
		})
	}
	err = AttachNamesToTopicFollowers(topicFollowers)
	if err != nil {
		return nil, jerr.Get("error attaching names to topic followers", err)
	}
	if len(selfPkHash) != 0 {
		err = AttachReputationToTopicFollowers(topicFollowers)
		if err != nil {
			return nil, jerr.Get("error attaching reputation to topic followers", err)
		}
	}
	return topicFollowers, nil
}
