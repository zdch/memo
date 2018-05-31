package cache

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
)

func GetLastTopicList(cookieId string) (string, error) {
	var lastTopicList string
	err := GetItem(getBLastTopicListName(cookieId), &lastTopicList)
	if err != nil {
		if IsMissError(err) {
			return "", nil
		}
		return "", jerr.Get("error getting lsat topic list ", err)
	}
	return lastTopicList, nil
}

func SetLastTopicList(cookieId string, topicList string) error {
	err := SetItem(getBLastTopicListName(cookieId), topicList)
	if err != nil {
		return jerr.Get("error setting last topic list", err)
	}
	return nil
}

func getBLastTopicListName(cookieId string) string {
	return fmt.Sprintf("last-topic-list-%s", cookieId)
}
