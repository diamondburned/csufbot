package badger

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v3"
	"github.com/diamondburned/csufbot/internal/csufbot"
	"github.com/diamondburned/csufbot/internal/lms"
	"github.com/pkg/errors"
)

type CourseStore struct {
	db *badger.DB
}

func (store *CourseStore) Course(id lms.CourseID) (*csufbot.Course, error) {
	var course *csufbot.Course
	return course, unmarshalString(store.db, "course", string(id), &course)
}

func (store *CourseStore) Courses(out map[lms.CourseID]csufbot.Course) error {
	return store.db.View(func(txn *badger.Txn) error {
		for id := range out {
			var c csufbot.Course
			k := joinKeys("course", []byte(string(id)))

			if err := unmarshalFromTxn(txn, k, &c); err != nil {
				return errors.Wrapf(err, "failed to get course ID %s", id)
			}

			out[id] = c
		}

		return nil
	})
}

func (store *CourseStore) UpsertCourses(courses ...csufbot.Course) error {
	type courseJSON struct {
		id   []byte
		json []byte
	}

	var jsons = make([]courseJSON, len(courses))

	for i, course := range courses {
		b, err := json.Marshal(course)
		if err != nil {
			return err
		}

		jsons[i] = courseJSON{
			id:   []byte(course.ID),
			json: b,
		}
	}

	return store.db.Update(func(txn *badger.Txn) error {
		for _, course := range jsons {
			if err := txn.Set(joinKeys("course", course.id), course.json); err != nil {
				return err
			}
		}

		return nil
	})
}
