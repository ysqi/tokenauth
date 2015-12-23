// Copyright 2016 Author YuShuangqi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tokenauth_test

import (
	"fmt"
	"github.com/ysqi/tokenauth"
	. "gopkg.in/check.v1"
	"sync"
	"time"
)

func (s *S) TestStore_Bolt_Init(c *C) {

	st := tokenauth.NewBoltDBFileStore()
	defer st.Close()

	err := st.Open("")
	c.Assert(err, NotNil)

	err = st.Open("path")
	c.Assert(err, NotNil)

	err = st.Open("{}")
	c.Assert(err, NotNil)

	err = st.Open(`{"goodpath":""}`)
	c.Assert(err, NotNil)

	err = st.Open(`{"path":""}`)
	c.Assert(err, NotNil)

	file := tempfile()
	err = st.Open(fmt.Sprintf(`{"path":"%s"}`, file))
	c.Assert(err, IsNil)
}

func (s *S) TestStore_Bolt_DBPath(c *C) {

	st := tokenauth.NewBoltDBFileStore()
	defer st.Close()

	c.Assert(st.DBPath(), Equals, "")

	file := tempfile()
	st.Open(fmt.Sprintf(`{"path":"%s"}`, file))
	c.Assert(st.DBPath(), Equals, file)
}

func (s *S) TestStore_Bolt_Audience_Save(c *C) {

	st := openBoltStore()
	defer st.Close()

	err := st.SaveAudience(nil)
	c.Assert(err, NotNil)

	err = st.SaveAudience(&tokenauth.Audience{})
	c.Assert(err, NotNil)

	item := newAudience()
	err = st.SaveAudience(item)
	c.Assert(err, IsNil)

	newItem, err := st.GetAudience(item.ID)
	c.Assert(err, IsNil)
	c.Assert(newItem, DeepEquals, item)
}

func (s *S) TestStore_Bolt_Audience_Save_Repeat(c *C) {

	st := openBoltStore()
	defer st.Close()

	item := newAudience()
	item.Name = "oldAudience"
	item.Secret = "oldSecret"
	item.TokenPeriod = 1
	st.SaveAudience(item)

	item.Name = "newAudience"
	item.Secret = "newSecret"
	item.TokenPeriod = 2
	st.SaveAudience(item)

	newItem, err := st.GetAudience(item.ID)
	c.Assert(err, IsNil)
	c.Assert(newItem, DeepEquals, item)
}

func (s *S) TestStore_Bolt_Audience_Save_RepeatWithToken(c *C) {

	st := openBoltStore()
	defer st.Close()

	item := newAudience()
	st.SaveAudience(item)
	tokens := make([]*tokenauth.Token, 10)
	for i := 0; i < 10; i++ {
		tokens[i], _ = tokenauth.NewToken(item, keyPorvider.GenerateTokenString)
		st.SaveToken(tokens[i])
	}

	//save again
	st.SaveAudience(item)

	for i := 0; i < 10; i++ {
		newToken, err := st.GetToken(tokens[i].Value)
		c.Assert(err, IsNil)
		c.Assert(newToken, IsNil)
	}
}

func (s *S) TestStore_Bolt_Audience_Delete(c *C) {

	st := openBoltStore()
	defer st.Close()

	err := st.DeleteAudience("")
	c.Assert(err, NotNil)

	err = st.DeleteAudience(" ")
	c.Assert(err, IsNil)

	err = st.DeleteAudience("empty")
	c.Assert(err, IsNil)

	item := newAudience()
	st.SaveAudience(item)
	err = st.DeleteAudience(item.ID)
	c.Assert(err, IsNil)

	newItem, err := st.GetAudience(item.ID)
	c.Assert(err, IsNil)
	c.Assert(newItem, IsNil)
}

func (s *S) TestStore_Bolt_Audience_DeleteWithToken(c *C) {

	st := openBoltStore()
	defer st.Close()

	item := newAudience()
	st.SaveAudience(item)

	tokens := make([]*tokenauth.Token, 10)
	for i := 0; i < 10; i++ {
		tokens[i], _ = tokenauth.NewToken(item, keyPorvider.GenerateTokenString)
		st.SaveToken(tokens[i])
	}

	err := st.DeleteAudience(item.ID)
	c.Assert(err, IsNil)

	newItem, err := st.GetAudience(item.ID)
	c.Assert(err, IsNil)
	c.Assert(newItem, IsNil)

	for i := 0; i < 10; i++ {
		newToken, err := st.GetToken(tokens[i].Value)
		c.Assert(err, IsNil)
		c.Assert(newToken, IsNil)
	}
}

func (s *S) TestStore_Bolt_Audience_Get(c *C) {

	st := openBoltStore()
	defer st.Close()

	item, err := st.GetAudience("")
	c.Assert(err, NotNil)

	item, err = st.GetAudience("  ")
	c.Assert(err, IsNil)
	c.Assert(item, IsNil)

	item, err = st.GetAudience("empty")
	c.Assert(err, IsNil)
	c.Assert(item, IsNil)

	items := make([]*tokenauth.Audience, 10)
	for i := 0; i < 10; i++ {
		items[i] = newAudience()
		st.SaveAudience(items[i])
	}
	for i := 0; i < 10; i++ {
		newItem, err := st.GetAudience(items[i].ID)
		c.Assert(err, IsNil)
		c.Assert(newItem, DeepEquals, items[i])
	}

}

func (s *S) TestStore_Bolt_Token_Save_Empty(c *C) {
	st := openBoltStore()
	defer st.Close()

	err := st.SaveToken(nil)
	c.Assert(err, NotNil)

	err = st.SaveToken(&tokenauth.Token{})
	c.Assert(err, NotNil)

	err = st.SaveToken(&tokenauth.Token{ClientID: "id"})
	c.Assert(err, NotNil)

	err = st.SaveToken(&tokenauth.Token{Value: "value"})
	c.Assert(err, NotNil)

	err = st.SaveToken(&tokenauth.Token{ClientID: "id", Value: "value"})
	c.Assert(err, NotNil)

	err = st.SaveToken(&tokenauth.Token{ClientID: "id", SingleID: "SingleID", Value: "value"})
	c.Assert(err, NotNil)

	err = st.SaveToken(&tokenauth.Token{SingleID: "SingleID", Value: "value"})
	c.Assert(err, IsNil)

}

func (s *S) TestStore_Bolt_Token_Save(c *C) {
	st := openBoltStore()
	defer st.Close()

	for ii := 0; ii < 10; ii++ {

		item := newAudience()
		st.SaveAudience(item)

		tokens := make([]*tokenauth.Token, 10)
		for i := 0; i < 10; i++ {
			tokens[i], _ = tokenauth.NewToken(item, keyPorvider.GenerateTokenString)
			st.SaveToken(tokens[i])
		}
		for i := 0; i < 10; i++ {
			newToken, err := st.GetToken(tokens[i].Value)
			c.Assert(err, IsNil)
			c.Assert(newToken, DeepEquals, tokens[i])
		}
	}

}

func (s *S) TestStore_Bolt_SinglnToken_Save(c *C) {
	st := openBoltStore()
	defer st.Close()

	item := newAudience()
	var err error
	for ii := 0; ii < 10; ii++ {

		tokens := make([]*tokenauth.Token, 10)
		for i := 0; i < 10; i++ {
			tokens[i], err = tokenauth.NewSingleToken(fmt.Sprintf("singleID%d", ii), item, keyPorvider.GenerateTokenString)
			c.Assert(err, IsNil)
			c.Assert(tokens[i], NotNil)
			st.SaveToken(tokens[i])
		}
		for i := 0; i < 10; i++ {
			newToken, err := st.GetToken(tokens[i].Value)
			c.Assert(err, IsNil)
			if i != 9 {
				c.Assert(newToken, IsNil)
			} else {
				c.Assert(newToken, DeepEquals, tokens[i])
			}
		}
	}

}

func (s *S) TestStore_Bolt_Token_Save_Effective(c *C) {
	st := openBoltStore()
	defer st.Close()

	item := newAudience()
	st.SaveAudience(item)

	token, _ := tokenauth.NewToken(item, keyPorvider.GenerateTokenString)

	token.DeadLine = 0
	err := st.SaveToken(token)
	c.Assert(err, IsNil)

	token.DeadLine = time.Now().Unix() - 1
	err = st.SaveToken(token)
	c.Assert(err, NotNil)

	token.DeadLine = time.Now().Unix() + 1
	err = st.SaveToken(token)
	c.Assert(err, IsNil)

}

func (s *S) TestStore_Bolt_Token_Get_Empty(c *C) {
	st := openBoltStore()
	defer st.Close()

	token, err := st.GetToken("")
	c.Assert(err, NotNil)
	c.Assert(token, IsNil)

	token, err = st.GetToken(" ")
	c.Assert(err, IsNil)
	c.Assert(token, IsNil)

	token, err = st.GetToken("empty")
	c.Assert(err, IsNil)
	c.Assert(token, IsNil)

}

func (s *S) TestStore_Bolt_Token_Effective(c *C) {
	st := openBoltStore()
	defer st.Close()

	item := newAudience()
	st.SaveAudience(item)

	token, _ := tokenauth.NewToken(item, keyPorvider.GenerateTokenString)
	token.DeadLine = time.Now().Unix() + 2
	st.SaveToken(token)

	newToken, err := st.GetToken(token.Value)
	c.Assert(err, IsNil)
	c.Assert(newToken, DeepEquals, token)

	time.Sleep(time.Second * 3)
	c.Log("current time:", time.Now().Unix(), ",tokenInfo:", token)

	newToken, err = st.GetToken(token.Value)
	c.Assert(err, IsNil)
	c.Assert(newToken, NotNil)
	c.Assert(newToken.Expired(), Equals, true)
}

func (s *S) TestStore_Bolt_Token_Delete(c *C) {

	st := openBoltStore()
	defer st.Close()

	err := st.DeleteToken("")
	c.Assert(err, NotNil)

	err = st.DeleteToken(" ")
	c.Assert(err, NotNil)

	err = st.DeleteToken("empty")
	c.Assert(err, NotNil)

	item := newAudience()
	st.SaveAudience(item)
	token, _ := tokenauth.NewToken(item, keyPorvider.GenerateTokenString)
	st.SaveToken(token)

	err = st.DeleteToken(token.Value)
	c.Assert(err, IsNil)

	newToken, err := st.GetToken(token.Value)
	c.Assert(err, IsNil)
	c.Assert(newToken, IsNil)
}

func (s *S) T2estStore_Bolt_Token_Get_Complicating(c *C) {
	st := openBoltStore()
	defer st.Close()

	count := 10
	oneJobItemCount := 100000
	wg := sync.WaitGroup{}
	wg.Add(count * oneJobItemCount)
	for job := 0; job < count; job++ {

		item := newAudience()
		st.SaveAudience(item)

		tokensChannel := make(chan *tokenauth.Token, oneJobItemCount)

		go func() {
			for i := 0; i < oneJobItemCount; i++ {
				token, _ := tokenauth.NewToken(item, keyPorvider.GenerateTokenString)
				err := st.SaveToken(token)
				c.Assert(err, IsNil)
				tokensChannel <- token
			}
		}()

		go func() {
			sum := 0
			select {
			case token := <-tokensChannel:
				wg.Done()
				newToken, err := st.GetToken(token.Value)
				c.Assert(err, IsNil)
				c.Assert(newToken, DeepEquals, token)

				err = st.DeleteToken(token.Value)
				c.Assert(err, IsNil)

				sum += 1

				if sum == oneJobItemCount {
					break
				}
			}
		}()
	}

	wg.Wait()

}

var keyPorvider = tokenauth.DefaultProvider{}

func newAudience() *tokenauth.Audience {
	item, _ := tokenauth.NewAudience("test", keyPorvider.GenerateSecretString)
	return item
}
