package cache

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/db"
)

type ProfilePic struct {
	Has       bool
	Id        uint
	Extension string
}

func GetProfilePic(pkHash []byte) (ProfilePic, error) {
	var profilePic ProfilePic
	err := GetItem(getHasPicName(pkHash), &profilePic)
	if err == nil {
		return profilePic, nil
	}
	setPic, err := db.GetPicForPkHash(pkHash)
	if err != nil {
		return ProfilePic{}, jerr.Get("error determining has pic", err)
	}
	if setPic != nil {
		profilePic = ProfilePic{
			Has:       true,
			Id:        setPic.Id,
			Extension: setPic.GetExtension(),
		}
	}
	SetProfilePic(pkHash, profilePic)
	return profilePic, nil
}

func SetProfilePic(pkHash []byte, profilePic ProfilePic) error {
	err := SetItem(getHasPicName(pkHash), profilePic)
	if err != nil {
		return jerr.Get("error setting has pic", err)
	}
	return nil
}

func ClearHasPic(pkHash []byte) error {
	err := DeleteItem(getHasPicName(pkHash))
	if err != nil && ! IsMissError(err) {
		return jerr.Get("error clearing has pic", err)
	}
	return nil
}

func getHasPicName(pkHash []byte) string {
	return fmt.Sprintf("profile-pic-%x", pkHash)
}
