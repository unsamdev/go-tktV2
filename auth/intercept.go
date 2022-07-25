package auth

import (
	"fmt"
	"github.com/fvk113/go-tkt/sql"
	"net/http"
)

func Auditable(delegate func(txCtx *sql.TxCtx, w http.ResponseWriter, r *http.Request)) func(txCtx *sql.TxCtx, w http.ResponseWriter, r *http.Request) {
	return func(txCtx *sql.TxCtx, w http.ResponseWriter, r *http.Request) {
		tokenEntry := r.Context().Value("tokenEntry")
		if tokenEntry != nil {
			txCtx.ExecSql(fmt.Sprintf("set local tkt.user_name to '%d';", tokenEntry.(*TokenEntry).UserId))
		}
		txCtx.ExecSql(fmt.Sprintf("set local tkt.context to '%s';", r.URL.Path))
		delegate(txCtx, w, r)
	}
}
