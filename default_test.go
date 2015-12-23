// Copyright 2016 Author YuShuangqi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tokenauth_test

import (
	"github.com/ysqi/tokenauth"
	. "gopkg.in/check.v1"
)

func (s *S) TestProvider_NewSecret(c *C) {

	provider := &tokenauth.DefaultProvider{}
	audience, _ := tokenauth.NewAudience("test", provider.GenerateSecretString)

	var secret string
	for i := 0; i < 10; i++ {
		newSecret := provider.GenerateSecretString(audience.ID)

		c.Check(len(newSecret), Equals, 32)
		c.Assert(newSecret, Not(Equals), secret, Commentf("New Secret is not unique"))
		secret = newSecret
	}

}

func (s *S) TestProvider_Tokne_New(c *C) {

	provider := &tokenauth.DefaultProvider{}
	audience, _ := tokenauth.NewAudience("test", provider.GenerateSecretString)

	tokens := make([]string, 20)
	for i := 0; i < 10; i++ {
		token, _ := tokenauth.NewToken(audience, provider.GenerateTokenString)
		tokens[i] = token.Value
	}

	//update secret
	audience.Secret = provider.GenerateSecretString(audience.ID)
	for i := 10; i < 20; i++ {
		token, _ := tokenauth.NewToken(audience, provider.GenerateTokenString)
		tokens[i] = token.Value
	}

	for i := 0; i < 20; i++ {
		token := tokens[i]
		for j, token2 := range tokens {
			if j != i {
				c.Assert(token, Not(Equals), token2, Commentf("Generated Token String is not unique"))
			}
		}
	}

}
