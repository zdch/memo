package db

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/bitcoin/script"
	"html"
	"time"
)

type MemoTopicFollow struct {
	Id         uint   `gorm:"primary_key"`
	TxHash     []byte `gorm:"unique;size:50"`
	ParentHash []byte
	PkHash     []byte `gorm:"index:pk_hash"`
	PkScript   []byte
	Topic      string `gorm:"index:topic"`
	BlockId    uint
	Block      *Block
	Unfollow   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (m MemoTopicFollow) Save() error {
	result := save(&m)
	if result.Error != nil {
		return jerr.Get("error saving memo follow topic", result.Error)
	}
	return nil
}

func (m MemoTopicFollow) GetTransactionHashString() string {
	hash, err := chainhash.NewHash(m.TxHash)
	if err != nil {
		jerr.Get("error getting chainhash from memo follow topic", err).Print()
		return ""
	}
	return hash.String()
}

func (m MemoTopicFollow) GetScriptString() string {
	return html.EscapeString(script.GetScriptString(m.PkScript))
}

func (m MemoTopicFollow) GetTimeString() string {
	if m.BlockId != 0 {
		return m.Block.Timestamp.Format("2006-01-02 15:04:05")
	}
	return "Unconfirmed"
}

func GetMemoTopicFollow(txHash []byte) (*MemoTopicFollow, error) {
	var memoFollowTopic MemoTopicFollow
	err := find(&memoFollowTopic, MemoTopicFollow{
		TxHash: txHash,
	})
	if err != nil {
		return nil, jerr.Get("error getting memo follow topic", err)
	}
	return &memoFollowTopic, nil
}

func GetMemoTopicFollowCountForUser(pkHash []byte) (uint, error) {
	if len(pkHash) == 0 {
		return 0, nil
	}
	db, err := getDb()
	if err != nil {
		return 0, jerr.Get("error getting db", err)
	}
	sql := "" +
		"SELECT COALESCE(SUM(IF(unfollow, 0, 1)), 0) AS following_count " +
		"FROM memo_topic_follows " +
		"JOIN (" +
		"	SELECT MAX(id) AS id" +
		"	FROM memo_topic_follows" +
		"	WHERE pk_hash = ?" +
		"	GROUP BY memo_topic_follows.topic" +
		") sq ON (sq.id = memo_topic_follows.id)"
	query := db.Raw(sql, pkHash)
	var cnt uint
	row := query.Row()
	err = row.Scan(&cnt)
	if err != nil {
		if IsNoRowsInResultSetError(err) {
			return 0, nil
		}
		return 0, jerr.Get("error in topic is following query", err)
	}
	return cnt, nil
}

func IsFollowingTopic(pkHash []byte, topic string) (bool, error) {
	if len(pkHash) == 0 || topic == "" {
		return false, nil
	}
	db, err := getDb()
	if err != nil {
		return false, jerr.Get("error getting db", err)
	}
	sql := "" +
		"SELECT " +
		"	COALESCE(unfollow, 1) AS is_following " +
		"FROM memo_topic_follows " +
		"JOIN (" +
		"	SELECT MAX(id) AS id" +
		"	FROM memo_topic_follows" +
		"	WHERE pk_hash = ? AND topic = ?" +
		") sq ON (sq.id = memo_topic_follows.id)"
	query := db.Raw(sql, pkHash, topic)
	var cnt uint
	row := query.Row()
	err = row.Scan(&cnt)
	if err != nil {
		if IsNoRowsInResultSetError(err) {
			return false, nil
		}
		return false, jerr.Get("error in topic is following query", err)
	}
	return cnt == 0, nil
}

func GetFollowersForTopic(topic string) ([]*MemoTopicFollow, error) {
	db, err := getDb()
	if err != nil {
		return nil, jerr.Get("error getting db", err)
	}
	joinSql := "" +
		"JOIN (" +
		"SELECT MAX(id) AS id " +
		"FROM memo_topic_follows " +
		"GROUP BY pk_hash, topic" +
		") sq ON (sq.id = memo_topic_follows.id)"
	query := db.
		Joins(joinSql).
		Where("topic = ?", topic).
		Where("unfollow = 0")
	var memoTopicFollows []*MemoTopicFollow
	result := query.Find(&memoTopicFollows)
	if result.Error != nil {
		if IsNoRowsInResultSetError(result.Error) {
			return nil, nil
		}
		return nil, jerr.Get("error in topic followers query", result.Error)
	}
	return memoTopicFollows, nil
}

func GetFollowerCountForTopic(topic string) (uint, error) {
	db, err := getDb()
	if err != nil {
		return 0, jerr.Get("error getting db", err)
	}
	joinSql := "" +
		"JOIN (" +
		"SELECT MAX(id) AS id " +
		"FROM memo_topic_follows " +
		"GROUP BY pk_hash, topic" +
		") sq ON (sq.id = memo_topic_follows.id)"
	query := db.
		Table("memo_topic_follows").
		Joins(joinSql).
		Where("topic = ?", topic).
		Where("unfollow = 0")
	var cnt uint
	result := query.Count(&cnt)
	if result.Error != nil {
		if IsNoRowsInResultSetError(result.Error) {
			return 0, nil
		}
		return 0, jerr.Get("error in topic followers query", result.Error)
	}
	return cnt, nil
}
