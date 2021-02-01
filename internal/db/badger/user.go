package badger

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/csufbot"
)

type UserStore struct {
	db *badger.DB
}

func (store *UserStore) User(id discord.UserID) (*csufbot.User, error) {
	var user *csufbot.User
	return user, unmarshal(store.db, "user", u64b(uint64(id)), &user)
}

func (store *UserStore) Sync(id discord.UserID, new csufbot.UserInService) error {
	keyBuf := joinKeys("user", u64b(uint64(id)))
	var user *csufbot.User

	return store.db.Update(func(txn *badger.Txn) error {
		if err := unmarshalFromTxn(txn, keyBuf, &user); err != nil {
			return err
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
