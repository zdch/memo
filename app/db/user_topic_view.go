package db

import (
	"github.com/jchavannes/jgo/jerr"
	"time"
)

type UserTopicView struct {
	Id         uint `gorm:"primary_key"`
	UserPkHash []byte
	Topic      string
	LastPostId uint
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func GetLastTopicPostIds(userPkHash []byte, topics []string) ([]*UserTopicView, error) {
	db, err := getDb()
	if err != nil {
		return nil, jerr.Get("error getting db", err)
	}
	var userTopicViews []*UserTopicView
	result := db.
		Where("user_pk_hash = ?", userPkHash).
		Where("topic IN (?)", topics).
		Find(&userTopicViews)
	if result.Error != nil {
		return nil, jerr.Get("error running last topic post ids query", result.Error)
	}
	return userTopicViews, nil
}

func GetLastTopicPostId(userPkHash []byte, topic string) (uint, error) {
	var userTopicView UserTopicView
	err := find(&userTopicView, UserTopicView{
		UserPkHash: userPkHash,
		Topic:      topic,
	})
	if err == nil {
		return userTopicView.LastPostId, nil
	}
	if ! IsRecordNotFoundError(err) {
		return 0, jerr.Get("error finding last post id", err)
	}
	return 0, nil
}

func SetLastTopicPostId(userPkHash []byte, topic string, lastPostId uint) error {
	var userTopicView UserTopicView
	err := find(&userTopicView, UserTopicView{
		UserPkHash: userPkHash,
		Topic:      topic,
	})
	if err != nil {
		if ! IsRecordNotFoundError(err) {
			return jerr.Get("error getting last user topic viewed from db", err)
		}
		userTopicView = UserTopicView{
			UserPkHash: userPkHash,
			Topic:      topic,
			LastPostId: lastPostId,
		}
		err := create(&userTopicView)
		if err != nil {
			return jerr.Get("error creating user topic viewed", err)
		}
		return nil
	} else {
		userTopicView.LastPostId = lastPostId
		result := save(userTopicView)
		if result.Error != nil {
			return jerr.Get("error saving user topic viewed", err)
		}
		return nil
	}
}
