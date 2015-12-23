// Copyright 2016 Author YuShuangqi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tokenauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"os"
	"path/filepath"
)

// Store implement by boltdb,see:https://github.com/boltdb/bolt
type BoltDBFileStore struct {
	Alias  string
	db     *bolt.DB
	dbPath string
}

var (
	// all tokens save in this buckert
	buckert_alltokens = []byte("bk_all_tokeninfo")
	// a one audience tokens save relation in audience's buckert child bukert.
	buckert_oneAudienceTokens = []byte("bk_one_audience_tokens")
	// one audience info key
	audienceInfoKey                 = []byte("one_audience")
	buckert_singletokens_singledids = []byte("bk_token_singleIDs")
)

func (store *BoltDBFileStore) DBPath() string {
	return store.dbPath
}

//delete audience and all tokens of this audience
func (store *BoltDBFileStore) deleteAudience(id string, tx *bolt.Tx) error {
	bk := tx.Bucket([]byte(id))
	if bk == nil {
		return nil
	}

	tokensBk := tx.Bucket(buckert_alltokens)
	if tokensBk != nil {
		audienceTokensBk := bk.Bucket(buckert_oneAudienceTokens)
		err := audienceTokensBk.ForEach(func(k, v []byte) error {
			return tokensBk.Delete(k)
		})
		if err != nil {
			return err
		}
	}
	// Last need delete client bucket
	return tx.DeleteBucket([]byte(id))
}

// Save audience into store.
// Returns error if error occured during execution.
func (store *BoltDBFileStore) SaveAudience(audience *Audience) error {

	if audience == nil || len(audience.ID) == 0 {
		return errors.New("audience id is empty.")
	}

	bytes, err := json.Marshal(audience)
	if err != nil {
		return err
	}

	return store.db.Update(func(tx *bolt.Tx) error {
		// need delete old audience info before save
		if err := store.deleteAudience(audience.ID, tx); err != nil {
			return err
		}

		bk, err := tx.CreateBucket([]byte(audience.ID))
		if err != nil {
			return err
		}
		if _, err = bk.CreateBucket(buckert_oneAudienceTokens); err != nil {
			return err
		}
		if err = bk.Put(audienceInfoKey, bytes); err != nil {
			return err
		}
		return nil
	})

}

// Delete audience and  all tokens of audience.
func (store *BoltDBFileStore) DeleteAudience(audienceID string) error {
	if len(audienceID) == 0 {
		return errors.New("audienceID is emtpty.")
	}

	return store.db.Update(func(tx *bolt.Tx) error {
		return store.deleteAudience(audienceID, tx)
	})
}

// Get audience info or returns error.
func (store *BoltDBFileStore) GetAudience(audienceID string) (audience *Audience, err error) {

	if len(audienceID) == 0 {
		return nil, errors.New("audienceID is emtpty.")
	}

	err = store.db.View(func(tx *bolt.Tx) error {
		bk := tx.Bucket([]byte(audienceID))
		// not found
		if bk == nil {
			return nil
		}
		bytes := bk.Get(audienceInfoKey)
		if bytes == nil {
			return nil
		}
		audience = &Audience{}
		if err := json.Unmarshal(bytes, audience); err != nil {
			audience = nil
			return err
		} else {
			return nil
		}
	})

	return

}

// Save token to store. return error when save fail.
// Save token json to store and save the relation of token with client if not single model.
// The first , token must not empty and effectiveness.
// Does not consider concurrency.
func (store *BoltDBFileStore) SaveToken(token *Token) error {
	if token == nil || len(token.Value) == 0 {
		return errors.New("token tokenString is empty.")
	}
	if len(token.ClientID) == 0 && len(token.SingleID) == 0 {
		return errors.New("token clientid and singleid,It can't be empty")
	}
	if token.Expired() {
		return errors.New("token is expired,not need save.")
	}

	//first to get token byte data
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return store.db.Update(func(tx *bolt.Tx) error {

		bk, err := tx.CreateBucketIfNotExists(buckert_alltokens)
		if err != nil {
			return err
		}

		// Singlge token has no client. But need delete old token
		if token.IsSingle() {
			if idsBK, err := tx.CreateBucketIfNotExists(buckert_singletokens_singledids); err != nil {
				return err
			} else {
				key := []byte(token.SingleID)

				// Find and delete old token.
				oldTokenValueData := idsBK.Get(key)
				if oldTokenValueData != nil {
					if err = store.deleteToken(string(oldTokenValueData), tx); err != nil {
						return err
					}
				}

				// Save new token relation
				if err = idsBK.Put(key, []byte(token.Value)); err != nil {
					return err
				}
			}

		} else {
			// Need add token key to client blucket.
			// Only save the relation of token with client.
			if au := tx.Bucket([]byte(token.ClientID)); au == nil {
				return errors.New("can not found audience, not save audience before save token ?")
			} else if err = au.Bucket(buckert_oneAudienceTokens).Put([]byte(token.Value), []byte("")); err != nil {
				return err
			}
		}
		// Safe check.
		if err != nil {
			return err
		}
		err = bk.Put([]byte(token.Value), tokenBytes)
		return err

	})
}

//Get token info if find in store,or return error
func (store *BoltDBFileStore) GetToken(tokenString string) (token *Token, err error) {
	if len(tokenString) == 0 {
		return nil, errors.New("tokenString is empty.")
	}

	err = store.db.View(func(tx *bolt.Tx) error {

		bk := tx.Bucket(buckert_alltokens)
		if bk == nil {
			return nil
		}
		tokenBytes := bk.Get([]byte(tokenString))
		if tokenBytes == nil {
			return nil
		}

		token = &Token{}
		if err := json.Unmarshal(tokenBytes, token); err != nil {
			token = nil
			return err
		} else {
			return nil
		}
	})

	return
}

// Delete token
// Returns error if delete token fail.
func (store *BoltDBFileStore) DeleteToken(tokenString string) error {

	if len(tokenString) == 0 {
		return errors.New("incompatible tokenString")
	}

	return store.db.Update(func(tx *bolt.Tx) error {
		return store.deleteToken(tokenString, tx)
	})
}

// Delete token
// Returns error if delete token fail.
func (store *BoltDBFileStore) deleteToken(tokenString string, tx *bolt.Tx) error {
	bk := tx.Bucket(buckert_alltokens)
	if bk == nil {
		return errors.New("incompatible tokenString")
	}

	key := []byte(tokenString)
	tokenBytes := bk.Get(key)
	// Not found
	if tokenBytes == nil {
		return errors.New("incompatible tokenString")
	}

	err := bk.Delete(key)
	if err != nil {
		return err
	}

	//clear the relation token with client
	token := &Token{}
	err = json.Unmarshal(tokenBytes, token)
	if err == nil && token.IsSingle() == false {
		err = tx.Bucket([]byte(token.ClientID)).Bucket(buckert_oneAudienceTokens).Delete(key)
	}
	return err
}

// Open db if db is not opened.
// Returns error if open new db fail or close old db fail if exist
func (store *BoltDBFileStore) open(dbPath string) error {

	// Do not open same db again
	if store.db != nil && dbPath == store.db.Path() {
		return nil
	}

	//check file dir path or create dir.
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(dir, 0666)
		}
		if err != nil {
			return err
		}
	}

	if db, err := bolt.Open(dbPath, 0666, nil); err != nil {
		return err
	} else {
		//close old db before use new db
		if store.db != nil {
			err = store.db.Close()
			if err != nil {
				db.Close() //need close new db
				return errors.New("store: close old db fail," + err.Error())
			}
		}
		store.db = db
		store.dbPath = db.Path()
	}

	return nil
}

// Close bolt db
func (store *BoltDBFileStore) Close() error {
	if store.db != nil {
		return store.db.Close()
	}
	return nil
}

// Delete token if token expired
func (store *BoltDBFileStore) DeleteExpired() {

	if store.db == nil {
		return
	}

	store.db.View(func(tx *bolt.Tx) error {
		// Get all tokens bucket.
		bk := tx.Bucket(buckert_alltokens)
		if bk == nil {
			return nil
		}
		// Foreach all tokens.
		bk.ForEach(func(k, v []byte) error {
			token := &Token{}
			if err := json.Unmarshal(v, token); err == nil {
				// Will delete token when expired
				if token.Expired() {
					store.DeleteToken(token.Value)
				}
			}
			return nil
		})
		return nil
	})
	return

}

// Init and Open BoltDBF.
// config is json string.
// e.g:
//  {"path":"./data/tokenbolt.db"}
func (store *BoltDBFileStore) Open(config string) error {

	if len(config) == 0 {
		return errors.New("boltdbStore: bolt db store config is empty")
	}

	var cf map[string]string

	if err := json.Unmarshal([]byte(config), &cf); err != nil {
		return fmt.Errorf("boltdbStore: unmarshal %p fail:%s", config, err.Error())
	}

	if path, ok := cf["path"]; !ok {
		return errors.New("boltdbStore: bolt db store config has no path key.")
	} else {
		return store.open(path)
	}

}

// new Bolt DB file store instance.
func NewBoltDBFileStore() *BoltDBFileStore {

	return &BoltDBFileStore{Alias: "BoltDBFileStore"}
}

func init() {
	RegStore("default", NewBoltDBFileStore())
}
