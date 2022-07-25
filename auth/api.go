package auth

import (
	"github.com/fvk113/go-tkt-convenios/sql"
	"github.com/fvk113/go-tkt-convenios/util"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type Api struct {
	txCtx     *sql.TxCtx
	sequences *sql.Sequences
}

func (o *Api) ValidarCredenciales(nombre string, clearTextPassword string) *Account {
	usuario := o.FindUsuarioByNombre(nombre)
	if usuario == nil {
		return nil
	}
	err := bcrypt.CompareHashAndPassword([]byte(*usuario.Password), []byte(clearTextPassword))
	if err == nil {
		return usuario
	} else {
		return nil
	}
}

func (o *Api) FindUsuario(id int64) *Account {
	return o.txCtx.FindStruct(Account{}, `select * from account where id = $1`, id).(*Account)
}

func (o *Api) FindUsuarioByNombre(nombre string) *Account {
	return o.txCtx.FindStruct(Account{}, `select * from account where nombre = $1`, nombre).(*Account)
}

func (o *Api) ListUsuarioByPattern(pattern string) []Account {
	sentence := `select * 
			from usuario u`
	params := make([]interface{}, 0)
	if len(pattern) > 0 {
		sentence += ` where lower(u.nombre) like $1 or lower(u.descripcion) like $1`
		params = append(params, "%"+strings.ToLower(pattern)+"%")
	}
	sentence += ` order by u.nombre`
	return o.txCtx.QueryStruct(Account{}, sentence, params...).([]Account)
}

func (o *Api) UpdatePassword(idUsuario int64, clearTextPassword string) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(clearTextPassword), bcrypt.DefaultCost)
	util.CheckErr(err)
	hashedPassword := string(bytes)
	o.txCtx.ExecSql("update account set password = $1 where id = $2", hashedPassword, idUsuario)
}

func (o *Api) CreateToken(data Token) *Token {
	id := o.txCtx.Seq().Next("token")
	data.Id = &id
	o.txCtx.InsertEntity("auth", data)
	return &data
}

func (o *Api) UpdateTokenTime(id int64, expirationTime time.Time, lastTime time.Time) {
	o.txCtx.ExecSql("update token set expirationtime = $2, lasttime = $3 where id = $1", id, expirationTime, lastTime)
}

func (o *Api) RemoveToken(id int64) {
	o.txCtx.ExecSql("delete from token where id = $1", id)
}

func (o *Api) FindTokenByValue(value string) *Token {
	return o.txCtx.FindStruct(Token{}, `select * from token where value = $1`, value).(*Token)
}

func (o *Api) ListUnexpiredToken() []Token {
	return o.txCtx.QueryStruct(Token{}, `select * from token where expirationtime > now()`).([]Token)
}

func (o *Api) RemoveExpiredToken() {
	o.txCtx.ExecSql("delete from token where expirationtime <= now()")
}
func (o *Api) DeleteUsuario(id string) {
	o.txCtx.ExecSql("delete from account where usuario_id = $1", id)
	o.txCtx.ExecSql("delete from usuario where id = $1", id)
}

func (o *Api) CountAdmin() int64 {
	var c *int64
	o.txCtx.QuerySingleton("select count(*) from account where role_name = ADMIN", []interface{}{&c})
	return *c
}

func NewApi(txContext *sql.TxCtx) *Api {
	return &Api{txCtx: txContext, sequences: txContext.Seq()}
}
