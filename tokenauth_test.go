// Copyright 2016 Author YuShuangqi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tokenauth_test

import (
	"fmt"
	"github.com/ysqi/tokenauth"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) SetUpSuite(c *C) {
	tokenauth.ChangeTokenStore(openBoltStore())
}

func NewSecret(clientID string) (Secret string) {
	return "TestForSecretString"
}

func GenerateTokenString(audience *tokenauth.Audience) string {
	return "TestForNewTokenString"
}

func (s *S) TestAudience_New(c *C) {

	audience, err := tokenauth.NewAudience("forTest", NewSecret)
	c.Assert(err, IsNil)
	c.Assert(audience, NotNil)
	c.Assert(audience.Name, Equals, "forTest")
	c.Assert(audience.Secret, Equals, "TestForSecretString")
	c.Assert(audience.TokenPeriod, Equals, tokenauth.TokenPeriod)

	newAudience, err := tokenauth.Store.GetAudience(audience.ID)
	c.Assert(err, IsNil)
	c.Assert(newAudience, NotNil)
	c.Assert(newAudience, DeepEquals, audience)
}

func (s *S) TestToken_New(c *C) {

	tokenauth.TokenPeriod = 10 //10s

	audience, _ := tokenauth.NewAudience("forTest", NewSecret)
	token, err := tokenauth.NewToken(audience, GenerateTokenString)
	c.Assert(err, IsNil)
	c.Assert(token, NotNil)
	c.Assert(token.ClientID, Equals, audience.ID)
	c.Assert(token.Value, Equals, "TestForNewTokenString")
	// c.Assert((token.DeadLine - time.Now().Unix()), Equals, int64(tokenauth.TokenPeriod))

	newToken, err := tokenauth.Store.GetToken(token.Value)
	c.Assert(err, IsNil)
	c.Assert(newToken, NotNil)
	c.Assert(newToken, DeepEquals, token)
}

func (s *S) TestToken_Valid(c *C) {

	token, err := tokenauth.ValidateToken("")
	c.Assert(err, NotNil)
	c.Assert(token, IsNil)

	token, err = tokenauth.ValidateToken(" ")
	c.Assert(err, NotNil)
	c.Assert(token, IsNil)

	token, err = tokenauth.ValidateToken("empty")
	c.Assert(err, NotNil)
	c.Assert(token, IsNil)

	tokenauth.TokenPeriod = 2 //2s
	audience, _ := tokenauth.NewAudience("forTest", NewSecret)
	token, _ = tokenauth.NewToken(audience, GenerateTokenString)

	newToken, err := tokenauth.ValidateToken(token.Value)
	c.Assert(err, IsNil)
	c.Assert(newToken, DeepEquals, token)

	time.Sleep(3 * time.Second)
	newToken, err = tokenauth.ValidateToken(token.Value)
	c.Assert(err, NotNil)
	c.Assert(newToken, NotNil)
	c.Assert(newToken.Expired(), Equals, true)
}

// tempfile returns a temporary file path.
func tempfile() string {
	f, _ := ioutil.TempFile("", "bolt-store-")
	f.Close()
	os.Remove(f.Name())
	// safe string
	return strings.Replace(f.Name(), `\`, `/`, -1)
}

func openBoltStore() *tokenauth.BoltDBFileStore {
	st := tokenauth.NewBoltDBFileStore()
	file := tempfile()
	if err := st.Open(fmt.Sprintf(`{"path":"%s"}`, file)); err != nil {
		panic(fmt.Sprintf("cannot init boltdb file %q :%s", file, err.Error()))
	}
	return st
}
