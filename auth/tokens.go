package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/fvk113/go-tkt-convenios/sql"
	"github.com/fvk113/go-tkt-convenios/util"
	"strings"
	"sync"
	"time"
)

type TokensConfig struct {
	Secret       *string `json:"secret"`
	TokenTimeout *int    `json:"tokenTimeout"`
}

type WebTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type WebTokenPayload struct {
	UserId         string    `json:"userId"`
	MinutesTimeout int       `json:"minutesTimeout"`
	CreationTime   time.Time `json:"creationTime"`
}

type TokenEntry struct {
	UserId         string
	CreationTime   time.Time
	ExpirationTime time.Time
	LastTime       time.Time
	TokenString    string
	tokenId        *int64
}

type TokenManager struct {
	databaseConfig sql.DatabaseConfig
	jwtConfig      TokensConfig
	tokenMap       map[string]*TokenEntry
	mux            sync.Mutex
}

func (o *TokensConfig) Validate() {
	if o.Secret == nil {
		panic("Invalid secret")
	}
	if o.TokenTimeout == nil {
		panic("Invalid tokenTimeout")
	}
}

func (o *TokenManager) EvictToken(tokenString string) {
	entry := o.doEvictToken(tokenString)
	sql.ExecuteTransactional(o.databaseConfig, func(txCtx *sql.TxCtx, args ...interface{}) interface{} {
		NewApi(txCtx).RemoveToken(*entry.tokenId)
		return nil
	})
}

func (o *TokenManager) doEvictToken(value string) *TokenEntry {
	o.mux.Lock()
	defer o.mux.Unlock()
	entry := o.tokenMap[value]
	delete(o.tokenMap, value)
	return entry
}

func (o *TokenManager) CreateToken(userId *string) string {

	header := WebTokenHeader{Alg: "HS256", Typ: "JWT"}

	payload := WebTokenPayload{UserId: *userId, MinutesTimeout: *o.jwtConfig.TokenTimeout, CreationTime: time.Now()}

	b, err := json.Marshal(header)
	util.CheckErr(err)
	content1 := base64.RawURLEncoding.EncodeToString(b)

	b, err = json.Marshal(payload)
	util.CheckErr(err)
	content2 := base64.RawURLEncoding.EncodeToString(b)

	content := content1 + "." + content2

	keyString := *o.jwtConfig.Secret
	key := []byte(keyString)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(content))
	value := string(content)

	tokenEntry := o.registerToken(&payload, value)
	sql.ExecuteTransactional(o.databaseConfig, func(txCtx *sql.TxCtx, args ...interface{}) interface{} {
		token := &Token{Value: &value, UsuarioId: userId, CreationTime: &tokenEntry.CreationTime, ExpirationTime: &tokenEntry.ExpirationTime,
			LastTime: &tokenEntry.LastTime}
		token = NewApi(txCtx).CreateToken(*token)
		tokenEntry.tokenId = token.Id
		return nil
	})

	return value
}

func (o *TokenManager) ValidateToken(token string) *TokenEntry {
	o.mux.Lock()
	defer o.mux.Unlock()
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		panic("Invalid token")
	}
	payload := WebTokenPayload{}
	decodeTokenPart(&payload, parts[1])
	tokenEntry, ok := o.tokenMap[token]
	if !ok {
		return nil
	}
	if tokenEntry.ExpirationTime.Before(time.Now()) {
		delete(o.tokenMap, token)
		return nil
	}
	if tokenEntry.UserId != payload.UserId {
		return nil
	}
	tokenEntry.LastTime = time.Now()
	tokenEntry.ExpirationTime = tokenEntry.LastTime.Add(time.Minute * time.Duration(*o.jwtConfig.TokenTimeout))
	tokenCopy := *tokenEntry

	sql.ExecuteTransactional(o.databaseConfig, func(txCtx *sql.TxCtx, args ...interface{}) interface{} {
		NewApi(txCtx).UpdateTokenTime(*tokenEntry.tokenId, tokenEntry.ExpirationTime, tokenEntry.LastTime)
		return nil
	})

	return &tokenCopy
}

func (o *TokenManager) registerToken(payload *WebTokenPayload, token string) *TokenEntry {
	o.mux.Lock()
	defer o.mux.Unlock()
	expiration := time.Now().Add(time.Minute * time.Duration(*o.jwtConfig.TokenTimeout))
	te := TokenEntry{UserId: payload.UserId, CreationTime: time.Now(), ExpirationTime: expiration, LastTime: time.Now(), TokenString: token}
	o.tokenMap[token] = &te
	return &te
}

func (o *TokenManager) Shrink() {
	sql.ExecuteTransactional(o.databaseConfig, func(txCtx *sql.TxCtx, args ...interface{}) interface{} {
		NewApi(txCtx).RemoveExpiredToken()
		return nil
	})
}

func (o *TokenManager) Load() {
	snapshot := sql.ExecuteTransactional(o.databaseConfig, func(txCtx *sql.TxCtx, args ...interface{}) interface{} {
		return NewApi(txCtx).ListUnexpiredToken()
	}).([]Token)
	for _, t := range snapshot {
		entry := TokenEntry{tokenId: t.Id, LastTime: *t.LastTime, ExpirationTime: *t.ExpirationTime, CreationTime: *t.CreationTime,
			UserId: *t.UsuarioId, TokenString: *t.Value}
		o.tokenMap[entry.TokenString] = &entry
	}
}

func NewTokenManager(databaseConfig sql.DatabaseConfig, jwtConfig TokensConfig) *TokenManager {
	tm := TokenManager{databaseConfig: databaseConfig, jwtConfig: jwtConfig, tokenMap: make(map[string]*TokenEntry), mux: sync.Mutex{}}
	return &tm
}

func decodeTokenPart(i interface{}, part string) {
	jsonBytes, err := base64.RawURLEncoding.DecodeString(part)
	util.CheckErr(err)
	util.JsonDecode(i, bytes.NewReader(jsonBytes))
}
