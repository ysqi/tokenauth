// Copyright 2016 Author YuShuangqi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tokenauth_test

import (
	"github.com/ysqi/tokenauth"
	. "gopkg.in/check.v1"
)

func (s *S) TestObjectId_New(c *C) {
	// Generate 10 ids
	ids := make([]tokenauth.ObjectId, 10)
	for i := 0; i < 10; i++ {
		ids[i] = tokenauth.NewObjectId()
	}
	for i := 1; i < 10; i++ {
		prevId := ids[i-1]
		id := ids[i]
		// Test for uniqueness among all other 9 generated ids
		for j, tid := range ids {
			if j != i {
				c.Assert(id, Not(Equals), tid, Commentf("Generated ObjectId is not unique"))
			}
		}
		// Check that timestamp was incremented and is within 30 seconds of the previous one
		secs := id.Time().Sub(prevId.Time()).Seconds()
		c.Assert((secs >= 0 && secs <= 30), Equals, true, Commentf("Wrong timestamp in generated ObjectId"))
		// Check that machine ids are the same
		c.Assert(id.Machine(), DeepEquals, prevId.Machine())
		// Check that pids are the same
		c.Assert(id.Pid(), Equals, prevId.Pid())
		// Test for proper increment
		delta := int(id.Counter() - prevId.Counter())
		c.Assert(delta, Equals, 1, Commentf("Wrong increment in generated ObjectId"))
	}
}
