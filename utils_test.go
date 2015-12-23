// Copyright 2016 Author YuShuangqi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tokenauth_test

import (
	"github.com/ysqi/tokenauth"
	. "gopkg.in/check.v1"
)

func (s *S) TestUtls_RandomString(c *C) {

	strs := make([]string, 20)
	for i := 0; i < 10; i++ {
		strs[i] = tokenauth.GenerateRandomString(10, false)
		strs[20-i-1] = tokenauth.GenerateRandomString(10, true)
	}
	for i := 0; i < 20; i++ {
		str := strs[i]
		for n, nstr := range strs {
			if n != i {
				c.Assert(str, Not(Equals), nstr, Commentf("Generated RandomString is not unique"))
			}
		}
	}
}

func (s *S) TestUtls_RandomStringLength(c *C) {

	for i := 0; i < 20; i++ {
		str := tokenauth.GenerateRandomString(i, false)
		c.Check(len(str), Equals, i)
	}

}
