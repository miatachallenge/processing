package main

import (
	"context"
	"sort"

	r "gopkg.in/gorethink/gorethink.v4"
)

type Key struct {
	ID   string `json:"id" gorethink:"id"`
	Name string `json:"name" gorethink:"name"`
}

type tagsSlice []*Key

func (p tagsSlice) Len() int           { return len(p) }
func (p tagsSlice) Less(i, j int) bool { return p[i].ID < p[j].ID }
func (p tagsSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p tagsSlice) Sort()              { sort.Sort(p) }

func loadTags() {
	cursor, err := r.Table("keys").Run(db)
	if err != nil {
		panic(err)
	}
	keys := []*Key{}
	if err := cursor.All(&keys); err != nil {
		panic(err)
	}

	stateLock.Lock()
	state.Tags = keys
	stateLock.Unlock()

	go func() {
		cursor, err := r.Table("keys").Changes().Run(db, r.RunOpts{
			Context: context.Background(),
		})
		if err != nil {
			panic(err)
		}
		var change struct {
			OldVal *Key `gorethink:"old_val"`
			NewVal *Key `gorethink:"new_val"`
		}
		for cursor.Next(&change) {
			if change.OldVal != nil {
				// Remove old item
				stateLock.Lock()
				var foundIndex int = -1
				for i, tag := range state.Tags {
					if tag.ID == change.OldVal.ID {
						foundIndex = i
						break
					}
				}
				if foundIndex != -1 {
					state.Tags[foundIndex] = state.Tags[len(state.Tags)-1]
					state.Tags[len(state.Tags)-1] = nil
					state.Tags = state.Tags[:len(state.Tags)-1]

					state.Tags.Sort()
				}
				stateLock.Unlock()
			}

			if change.NewVal != nil {
				// New item
				stateLock.Lock()
				state.Tags = append(state.Tags, change.NewVal)
				state.Tags.Sort()
				stateLock.Unlock()
			}
		}
		if err := cursor.Err(); err != nil {
			panic(err)
		}
		if err := cursor.Close(); err != nil {
			panic(err)
		}
	}()
}
