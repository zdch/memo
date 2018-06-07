package cache

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/db"
)

func GetHasPic(pkHash []byte) (int8, error) {
	var hasPic int8
	err := GetItem(getHasPicName(pkHash), &hasPic)
	if err != nil {
		setPic, err := db.GetPicForPkHash(pkHash)
		if err != nil {
			return 0, jerr.Get("error determining has pic", err)
		}
		if setPic == nil {
			SetHasPic(pkHash, 0)
			hasPic = 0
		} else {
			SetHasPic(pkHash, 1)
			hasPic = 1
		}
	}

	return hasPic, nil
}

func SetHasPic(pkHash []byte, hasPic int8) error {
	err := SetItem(getHasPicName(pkHash), hasPic)
	if err != nil {
		return jerr.Get("error setting has pic", err)
	}
	return nil
}

func ClearHasPic(pkHash []byte) error {
	err := DeleteItem(getHasPicName(pkHash))
	if err != nil {
		return jerr.Get("error clearing has pic", err)
	}
	return nil
}

func getHasPicName(pkHash []byte) string {
	return fmt.Sprintf("has-pic-%x", pkHash)
}
