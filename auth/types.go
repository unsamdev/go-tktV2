package auth

import "time"

type Account struct {
	Id         *string `json:"id"`
	Email      *string `json:"email"`
	First_name *string `json:"first_name"`
	Is_enabled *bool   `json:"is_enabled"`
	Last_name  *string `json:"last_name"`
	Login      *string `json:"login"`
	Password   *string `json:"password"`
	Role_name  *string `json:"role_name"`
}

type Token struct {
	Id             *int64     `json:"id"`
	Value          *string    `json:"value"`
	UsuarioId      *string    `json:"usuarioId" sql:"usuario_id"`
	CreationTime   *time.Time `json:"creationTime"`
	ExpirationTime *time.Time `json:"expirationTime"`
	LastTime       *time.Time `json:"lastTime"`
}
