// Copyright 2016 Author YuShuangqi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tokenauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

const (
	// default secret length.
	SecretLength = 32
)

type DefaultProvider struct {
	Name string
}

func (d *DefaultProvider) GenerateSecretString(clientID string) (secretString string) {

	return GenerateRandomString(SecretLength, false)
}

func (d *DefaultProvider) GenerateTokenString(audience *Audience) string {

	if audience == nil {
		panic("audience is nil")
	}

	hash := hmac.New(sha256.New, []byte(audience.Secret))

	info := fmt.Sprintf("%s:%s:%d", audience.ID, GenerateRandomString(6, false), time.Now().Unix())
	hash.Write([]byte(info))

	return base64.StdEncoding.EncodeToString(hash.Sum(nil))

}
