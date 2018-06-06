package memo

import (
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/web"
	"github.com/memocash/memo/app/auth"
	"github.com/memocash/memo/app/bitcoin/memo"
	"github.com/memocash/memo/app/bitcoin/transaction"
	"github.com/memocash/memo/app/db"
	"github.com/memocash/memo/app/mutex"
	"github.com/memocash/memo/app/profile"
	"github.com/memocash/memo/app/res"
	"net/http"
)

var likeRoute = web.Route{
	Pattern:    res.UrlMemoLike + "/" + urlTxHash.UrlPart(),
	NeedsLogin: true,
	Handler: func(r *web.Response) {
		txHashString := r.Request.GetUrlNamedQueryVariable(urlTxHash.Id)
		txHash, err := chainhash.NewHashFromStr(txHashString)
		if err != nil {
			r.Error(jerr.Get("error getting transaction hash", err), http.StatusInternalServerError)
			return
		}
		user, err := auth.GetSessionUser(r.Session.CookieId)
		if err != nil {
			r.Error(jerr.Get("error getting session user", err), http.StatusInternalServerError)
			return
		}
		key, err := db.GetKeyForUser(user.Id)
		if err != nil {
			r.Error(jerr.Get("error getting key for user", err), http.StatusInternalServerError)
			return
		}
		hasSpendableTxOut, err := db.HasSpendable(key.PkHash)
		if err != nil {
			r.Error(jerr.Get("error getting spendable tx out", err), http.StatusInternalServerError)
			return
		}
		if ! hasSpendableTxOut {
			r.SetRedirect(res.UrlNeedFunds)
			return
		}
		post, err := profile.GetPostByTxHashWithReplies(txHash.CloneBytes(), key.PkHash, 0)
		if err != nil {
			r.Error(jerr.Get("error getting post", err), http.StatusInternalServerError)
			return
		}
		err = profile.AttachParentToPosts([]*profile.Post{post})
		if err != nil {
			r.Error(jerr.Get("error attaching parent to post", err), http.StatusInternalServerError)
			return
		}
		err = profile.AttachPollsToPosts([]*profile.Post{post})
		if err != nil {
			r.Error(jerr.Get("error attaching polls to posts", err), http.StatusInternalServerError)
			return
		}
		err = profile.AttachLikesToPosts([]*profile.Post{post})
		if err != nil {
			r.Error(jerr.Get("error attaching likes to posts", err), http.StatusInternalServerError)
			return
		}
		r.Helper["Post"] = post
		r.RenderTemplate(res.UrlMemoLike)
	},
}

var likeSubmitRoute = web.Route{
	Pattern:     res.UrlMemoLikeSubmit,
	NeedsLogin:  true,
	CsrfProtect: true,
	Handler: func(r *web.Response) {
		txHashString := r.Request.GetFormValue("txHash")
		txHash, err := chainhash.NewHashFromStr(txHashString)
		if err != nil {
			r.Error(jerr.Get("error getting transaction hash", err), http.StatusInternalServerError)
			return
		}
		memoPost, err := db.GetMemoPost(txHash.CloneBytes())
		if err != nil {
			r.Error(jerr.Get("error getting memo_post", err), http.StatusInternalServerError)
			return
		}

		password := r.Request.GetFormValue("password")
		user, err := auth.GetSessionUser(r.Session.CookieId)
		if err != nil {
			r.Error(jerr.Get("error getting session user", err), http.StatusInternalServerError)
			return
		}
		key, err := db.GetKeyForUser(user.Id)
		if err != nil {
			r.Error(jerr.Get("error getting key for user", err), http.StatusInternalServerError)
			return
		}

		privateKey, err := key.GetPrivateKey(password)
		if err != nil {
			r.Error(jerr.Get("error getting private key", err), http.StatusUnauthorized)
			return
		}

		userAddress := key.GetAddress()
		postAddress := memoPost.GetAddress()

		var likeTxBytes = txHash.CloneBytes()
		var tip = int64(r.Request.GetFormValueInt("tip"))

		var minInput = int64(memo.InputFee + 2*memo.OutputFee + len(likeTxBytes)) + transaction.DustMinimumOutput
		if tip > 0 {
			minInput += memo.InputFee + tip
		}
		fmt.Printf("minInput: %d\n", minInput)

		mutex.Lock(key.PkHash)
		txOuts, err := db.GetSpendableTxOuts(key.PkHash, minInput)
		if err != nil {
			mutex.Unlock(key.PkHash)
			r.Error(jerr.Get("error getting spendable tx out", err), http.StatusPaymentRequired)
			return
		}

		transactions := []transaction.SpendOutput{{
			Type: transaction.SpendOutputTypeMemoLike,
			Data: likeTxBytes,
		}}
		if tip != 0 {
			if tip < transaction.DustMinimumOutput {
				mutex.Unlock(key.PkHash)
				r.Error(jerr.Get("error tip not above dust limit", err), http.StatusUnprocessableEntity)
				return
			}
			if tip > 1e8 {
				mutex.Unlock(key.PkHash)
				r.Error(jerr.Get("error trying to tip too much", err), http.StatusUnprocessableEntity)
				return
			}
			transactions = append(transactions, transaction.SpendOutput{
				Type:    transaction.SpendOutputTypeP2PK,
				Address: postAddress,
				Amount:  tip,
			})
		}

		var totalInputs int64
		for _, txOut := range txOuts {
			totalInputs += txOut.Value
		}

		var fee = int64(len(txOuts)*memo.InputFee + (len(transactions)+1)*memo.OutputFee)
		var change = totalInputs - fee - tip
		fmt.Printf("not enough funds (c: %d, i: %d, f: %d, t: %d)", change, totalInputs, fee, tip)
		if change < transaction.DustMinimumOutput {
			mutex.Unlock(key.PkHash)
			r.Error(jerr.Newf("not enough funds (c: %d, i: %d, f: %d, t: %d)", change, totalInputs, fee, tip), http.StatusUnprocessableEntity)
			return
		}
		transactions = append(transactions, transaction.SpendOutput{
			Type:    transaction.SpendOutputTypeP2PK,
			Address: userAddress,
			Amount:  change,
		})
		var tx *wire.MsgTx
		tx, err = transaction.Create(txOuts, privateKey, transactions)
		if err != nil {
			mutex.Unlock(key.PkHash)
			r.Error(jerr.Get("error creating tx", err), http.StatusInternalServerError)
			return
		}

		fmt.Println(transaction.GetTxInfo(tx))
		fmt.Printf("TxSize: %d\n", tx.SerializeSize())
		transaction.QueueTx(tx)
		r.Write(tx.TxHash().String())
	},
}
