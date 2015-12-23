// Copyright 2016 Author YuShuangqi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tokenauth

import (
	"time"
)

// Audience Info, token rely on audience.
// Contains secret string , tokenPeriod for generatating token string.
type Audience struct {
	Name        string
	ID          string // Unique key for audience
	Secret      string //audience secret string,can update.
	TokenPeriod uint64 //token period ,unit: seconds.
}

// Token Info
type Token struct {
	ClientID string // Audience.ID
	SingleID string // Single Token ID
	Value    string // Token string
	DeadLine int64  // Token Expiration date, time unix.
}

// Returns this token is expried.
// Note: never exprires if  token's deadLine =0
func (t *Token) Expired() bool {
	if t.DeadLine == 0 {
		return false
	}
	return time.Now().Unix() >= t.DeadLine
}

// Returns true if token clientID is empty and signleID is not empty.
func (t *Token) IsSingle() bool {
	return len(t.ClientID) == 0 && len(t.SingleID) > 0
}
