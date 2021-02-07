package badger

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/csufbot"
	"github.com/pkg/errors"
)

type UserStore struct {
	*needDatabase
}

func (store UserStore) User(id discord.UserID) (*csufbot.User, error) {
	var user *csufbot.User
	return user, store.unmarshal("user", u64b(uint64(id)), &user)
}

func (store UserStore) Sync(id discord.UserID, new csufbot.UserInService) error {
	keyBuf := joinKeys("user", u64b(uint64(id)))
	var user *csufbot.User

	return store.db.Update(func(txn *badger.Txn) error {
		if err := unmarshalFromTxn(txn, keyBuf, &user); err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Wrap(err, "failed to get previous user state")
			}

			// Initialize a new user.
			user = &csufbot.User{ID: id}
		}

		var found bool
		for i, svc := range user.Services {
			if svc.ServiceHost == new.ServiceHost {
				found = true
				user.Services[i] = new
				break
			}
		}

		if !found {
			user.Services = append(user.Services, new)
		}

		return marshalToTxn(txn, keyBuf, user)
	})
}
