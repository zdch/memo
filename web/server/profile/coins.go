package profile

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/web"
	"github.com/memocash/memo/app/auth"
	"github.com/memocash/memo/app/db"
	"github.com/memocash/memo/app/res"
	"net/http"
)

var coinsRoute = web.Route{
	Pattern:    res.UrlProfileCoins,
	NeedsLogin: true,
	Handler: func(r *web.Response) {
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
		txOuts, err := db.GetSpendableTransactionOutputsForPkHash(key.PkHash)
		if err != nil {
			r.Error(jerr.Get("error getting spendable tx outputs for user", err), http.StatusInternalServerError)
			return
		}
		var totalValue int64
		for _, txOut := range txOuts {
			totalValue += txOut.Value
		}
		r.Helper["TxOuts"] = txOuts
		r.Helper["TotalValue"] = totalValue
		r.RenderTemplate(res.TmplProfileCoins)
	},
}
