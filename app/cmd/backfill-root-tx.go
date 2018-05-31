package cmd

import (
	"github.com/spf13/cobra"
	"github.com/memocash/memo/app/db"
	"fmt"
	"log"
)

var backfillRootTx = &cobra.Command{
	Use:   "backfill-root-tx",
	RunE: func(c *cobra.Command, args []string) error {

		for i := 1; i < 40000; i++ {
			memoPost, err := db.GetMemoPostById(uint(i))
			if err != nil {
				if db.IsRecordNotFoundError(err) {
					fmt.Printf("all done\n")
					break
				}
				log.Fatal(err)
			}
			backfillRootTxHash(memoPost.TxHash)
		}
		return nil
	},
}

func backfillRootTxHash(txhash []byte) []byte {
	memoPost, err := db.GetMemoPost(txhash)
	if err != nil {
		if db.IsRecordNotFoundError(err) {
			fmt.Printf("all done\n")
			return nil
		}
		log.Fatal(err)
	}
	if len(memoPost.RootTxHash) > 0 {
		return memoPost.RootTxHash
	}
	if memoPost.ParentTxHash == nil {
		memoPost.RootTxHash = memoPost.TxHash
		memoPost.Save()
		fmt.Printf("saved top level %d\n", memoPost.Id)
		return memoPost.RootTxHash
	} else {
		memoPost.RootTxHash = backfillRootTxHash(memoPost.ParentTxHash)
		memoPost.Save()
		fmt.Printf("saved child %d \n", memoPost.Id)
		return memoPost.RootTxHash
	}
	return nil
}