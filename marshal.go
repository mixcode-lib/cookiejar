// Copyright 2023 mixcode@github
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cookiejar

import (
	"encoding/json"
	"sort"
)

type jsonEntry struct {
	D, K string
	E    entry
}

// Marshal a cookejar into a JSON.
// If persistentOnly is set then only the persistent cookies are marshalled.
func (j *Jar) MarshalJson(persistentOnly bool) (data []byte, err error) {
	e := j.extract(persistentOnly)
	return json.Marshal(e)
}

// Marshal a cookiejar into an indented JSON.
func (j *Jar) MarshalJsonIndent(persistentOnly bool, prefix, indent string) (data []byte, err error) {
	e := j.extract(persistentOnly)
	return json.MarshalIndent(e, prefix, indent)
}

// convert the cookie map to name-sorted array
func (j *Jar) extract(persistentOnly bool) []jsonEntry {
	entries := make([]jsonEntry, 0)
	for ek, em := range j.entries {
		for vk, e := range em {
			if !persistentOnly || e.Persistent {
				entries = append(entries, jsonEntry{D: ek, K: vk, E: e})
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].E.seqNum < entries[j].E.seqNum
	})
	return entries
}

// Merge json marshalled by MarshalJson into the current cookie jar.
func (j *Jar) MergeJson(data []byte) (err error) {

	var entries []jsonEntry
	err = json.Unmarshal(data, &entries)
	if err != nil {
		return
	}

	for _, em := range entries {
		submap, ok := j.entries[em.D]
		if !ok {
			submap = make(map[string]entry)
		}
		id, e := em.K, em.E
		if old, ok := submap[id]; ok {
			//if old.Creation.Before(e.Creation) {
			//	e.Creation = old.Creation
			//}
			//if old.LastAccess.After(e.LastAccess) {
			//	e.LastAccess = old.LastAccess
			//}
			e.seqNum = old.seqNum
		} else {
			e.seqNum = j.nextSeqNum
			j.nextSeqNum++
		}
		submap[em.K] = e
		j.entries[em.D] = submap
	}

	return nil
}

// Clear all entries
func (j *Jar) Clear() {
	clear(j.entries)
	j.nextSeqNum = 0
}
