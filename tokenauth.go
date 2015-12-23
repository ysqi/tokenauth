// Copyright 2016 Author YuShuangqi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package Token generation and storage management .
// Simple Usage.
// 	import (
// 		"fmt"
// 		"github.com/ysqi/tokenauth"
// 	)
// 	func main() {
//
// 		if err := tokenauth.UseDeaultStore(); err != nil {
// 			panic(err)
// 		}
// 		defer tokenauth.Store.Close()
//
// 		// Ready.
// 		d := &tokenauth.DefaultProvider{}
// 		globalClient := tokenauth.NewAudienceNotStore("globalClient", d.GenerateSecretString)
//
// 		// New token
// 		token, err := tokenauth.NewSingleToken("singleID", globalClient, d.GenerateTokenString)
// 		if err != nil {
// 			fmt.Println("generate token fail,", err.Error())
// 			return
// 		}
// 		// Check token
// 		if checkToken, err := tokenauth.ValidateToken(token.Value); err != nil {
// 			fmt.Println("token check did not pass,", err.Error())
// 		} else {
// 			fmt.Println("token check pass,token Expiration date:", checkToken.DeadLine)
// 		}
//
// 	}
// Advanced Usage:
//
// 	secretFunc := func(clientID string) (secretString string) {
// 		return "myself secret for all client"
// 	}
// 	tokenFunc := func(audience *Audience) string {
// 		return "same token string"
// 	}
// 	globalClient := tokenauth.NewAudienceNotStore("globalClient", secretFunc)
// 	// New token
// 	t1, err := tokenauth.NewToken(globalClient, tokenFunc)
// 	t2, err := tokenauth.NewToken(globalClient, tokenFunc)
package tokenauth

import (
	"errors"
	"time"
)

// Token effective time,unti: seconds.
// Defult is 2 Hour.
var TokenPeriod uint64 = 7200 //2hour

//Global Token Store .
// Default use
var Store TokenStore

// Change token store and close old store.
// New token and New Audience whill be saved to new store,after use new store.
func ChangeTokenStore(newStore TokenStore) error {
	if newStore == nil {
		return errors.New("tokenauth: new store is nil.")
	}
	if Store != nil {
		if err := Store.Close(); err != nil {
			return err
		}
	}
	Store = newStore
	return nil
}

// Use default store.
// Default use bolt db file, "./data/tokendb.bolt" file open or create
func UseDeaultStore() error {

	s, err := NewStore("default", `{"path":"./data/tokendb.bolt"}`)
	if err != nil {
		return err
	}
	return ChangeTokenStore(s)
}

//Create Secret provider interface
type GenerateSecretString func(clientID string) (secretString string) //returns new secret string.

//Create token string provider interface
type GenerateTokenString func(audience *Audience) string //returns new token string

// New audience and this audience will be saved to store.
func NewAudience(name string, secretFunc GenerateSecretString) (*Audience, error) {

	audience := NewAudienceNotStore(name, secretFunc)

	//save to store
	if err := Store.SaveAudience(audience); err != nil {
		return nil, err
	} else {
		return audience, nil
	}
}

// Returns a new audience info,not save to store.
func NewAudienceNotStore(name string, secretFunc GenerateSecretString) *Audience {

	audience := &Audience{
		Name:        name,
		ID:          NewObjectId().Hex(),
		TokenPeriod: TokenPeriod,
	}
	audience.Secret = secretFunc(audience.ID)
	return audience
}

// New Token and this new token will be saved to store.
func NewToken(a *Audience, tokenFunc GenerateTokenString) (*Token, error) {
	token := &Token{
		ClientID: a.ID,
		Value:    tokenFunc(a),
	}
	if a.TokenPeriod == 0 {
		token.DeadLine = 0
	} else {
		token.DeadLine = time.Now().Unix() + int64(a.TokenPeriod)
	}

	if err := Store.SaveToken(token); err != nil {
		return nil, err
	} else {
		return token, nil
	}
}

// New Sign Token and this new token will be saved to store.
func NewSingleToken(singleID string, a *Audience, tokenFunc GenerateTokenString) (*Token, error) {
	token := &Token{
		SingleID: singleID,
		Value:    tokenFunc(a),
		DeadLine: time.Now().Unix() + int64(a.TokenPeriod),
	}
	if err := Store.SaveToken(token); err != nil {
		return nil, err
	} else {
		return token, nil
	}
}

// Returns Exist tokenstring or error.
// If token is exist but  expired, then delete token and return TokenExpired error.
func ValidateToken(tokenString string) (*Token, error) {

	if len(tokenString) == 0 {
		return nil, ERR_TokenEmpty
	}

	var token *Token
	var err error

	// Get token info
	if token, err = Store.GetToken(tokenString); err != nil {
		return nil, err
	}

	// Check token
	if token == nil || len(token.Value) == 0 {
		return nil, ERR_InvalidateToken
	}

	// Need delete token if token lose effectiveness
	if token.Expired() {
		if err = Store.DeleteToken(token.Value); err != nil {
			return nil, err
		}
		return token, ERR_TokenExpired
	}

	return token, nil
}

var (
	ERR_InvalidateToken = ValidationError{Code: "40001", Msg: "Invalid token"}
	ERR_TokenEmpty      = ValidationError{Code: "41001", Msg: "Token is empty"}
	ERR_TokenExpired    = ValidationError{Code: "42001", Msg: "Token is expired"}
)
