package cmd

import (
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/db"
	"github.com/spf13/cobra"
	"log"
)

var backfillRootTx = &cobra.Command{
	Use: "backfill-root-tx",
	RunE: func(c *cobra.Command, args []string) error {
		var printStatus = func(offset uint, postsUpdated int) {
			fmt.Printf("offset: %d, posts-updated: %d\n", offset, postsUpdated)
		}
		var offset uint
		var postsUpdated int
		for ; offset < 100000; offset += 25 {
			memoPosts, err := db.GetPosts(offset)
			if err != nil {
				if db.IsRecordNotFoundError(err) {
					break
				}
				log.Fatal(jerr.Get("error getting memo posts", err))
			}
			if len(memoPosts) == 0 {
				break
			}
			for _, memoPost := range memoPosts {
				if len(memoPost.ParentTxHash) == 0 {
					continue
				}
				var rootTxHash []byte
				prevMemoPost, err := db.GetMemoPost(memoPost.ParentTxHash)
				if err != nil {
					jerr.Get("error getting reply post from db", err).Print()
					continue
				} else {
					if len(prevMemoPost.ParentTxHash) > 0 {
						rootTxHash = prevMemoPost.RootTxHash
					} else {
						rootTxHash = prevMemoPost.TxHash
					}
				}
				memoPost.RootTxHash = rootTxHash
				err = memoPost.Save()
				if err != nil {
					log.Fatal(jerr.Get("error saving memo post", err))
				}
				postsUpdated++
				if postsUpdated%1000 == 0 {
					printStatus(offset, postsUpdated)
				}
			}
		}
		printStatus(offset, postsUpdated)
		fmt.Printf("all done\n")
		return nil
	},
}
