package top

import (
	"bytes"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/db"
)

func AttachNamesToTopicFollowers(topicFollowers []*TopicFollower) error {
	var namePkHashes [][]byte
	for _, topicFollower := range topicFollowers {
		for _, namePkHash := range namePkHashes {
			if bytes.Equal(namePkHash, topicFollower.MemoTopicFollow.PkHash) {
				continue
			}
		}
		namePkHashes = append(namePkHashes, topicFollower.MemoTopicFollow.PkHash)
	}
	setNames, err := db.GetNamesForPkHashes(namePkHashes)
	if err != nil {
		return jerr.Get("error getting set names for pk hashes", err)
	}
	for _, setName := range setNames {
		for _, topicFollower := range topicFollowers {
			if bytes.Equal(topicFollower.MemoTopicFollow.PkHash, setName.PkHash) {
				topicFollower.Name = setName.Name
			}
		}
	}
	return nil
}
