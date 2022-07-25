package web

import (
	"github.com/fvk113/go-tkt-convenios/auth"
	"github.com/fvk113/go-tkt-convenios/sql"
	"net/http"
)

func HandleAuthenticated(path string, tokenManager *auth.TokenManager, f func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(path, InterceptFatal(InterceptCORS(auth.InterceptAuth(tokenManager, f))))
}

func HandleTransactional(path string, databaseConfig sql.DatabaseConfig, f func(txContext *sql.TxCtx, w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(path, InterceptFatal(InterceptCORS(sql.InterceptTransactional(&databaseConfig, auth.Auditable(f)))))
}

func HandleAuthenticatedTransactional(path string, tokenManager *auth.TokenManager, databaseConfig sql.DatabaseConfig, f func(txContext *sql.TxCtx, w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(path, InterceptFatal(InterceptCORS(auth.InterceptAuth(tokenManager, sql.InterceptTransactional(&databaseConfig, auth.Auditable(f))))))
}

func HandleReadOnlyTransactional(path string, databaseConfig sql.DatabaseConfig, f func(txContext *sql.TxCtx, w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(path, InterceptFatal(InterceptCORS(sql.InterceptReadOnlyTransactional(&databaseConfig, f))))
}

func HandleNonTransactional(path string, f func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(path, InterceptFatal(InterceptCORS(f)))
}
