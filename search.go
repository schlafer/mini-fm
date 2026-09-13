package main

import (
	"os"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func search(query string) []os.DirEntry {
	if query == "" {
		return nil
	}

	var matches []os.DirEntry
	queryLower := strings.ToLower(query)

	for _, f := range Files {
		name := f.Name()
		nameLower := strings.ToLower(name)

		if strings.Contains(nameLower, queryLower) || fuzzy.MatchFold(query, name) {
			matches = append(matches, f)
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		iName := strings.ToLower(matches[i].Name())
		jName := strings.ToLower(matches[j].Name())

		if iName == queryLower {
			return true
		}
		if jName == queryLower {
			return false
		}

		iHasPrefix := strings.HasPrefix(iName, queryLower)
		jHasPrefix := strings.HasPrefix(jName, queryLower)
		if iHasPrefix && !jHasPrefix {
			return true
		}
		if !iHasPrefix && jHasPrefix {
			return false
		}

		if len(iName) != len(jName) {
			return len(iName) < len(jName)
		}

		return iName < jName
	})

	return matches
}

// CurrentList returns visible entry names, so active search results, or
// the full directory listing when there's no search. cached until
// underlying Files/Results change (see invalidateList)
func CurrentList() []string {
	if listCache != nil {
		return listCache
	}
	src := Files
	if len(Results) > 0 {
		src = Results
	}
	names := make([]string, len(src))
	for i, f := range src {
		names[i] = f.Name()
	}
	listCache = names
	return names
}

// drops memoized CurrentList() result
// call after reassigning Files or Results
func invalidateList() {
	listCache = nil
}

// ret name of selected entry in active list, and
// whether one exists. shared guard against empty or out-of-range selection
func currentName() (string, bool) {
	list := CurrentList()
	if len(list) == 0 || Selected >= len(list) {
		return "", false
	}
	return list[Selected], true
}

func MoveCursor(n int) {
	list := CurrentList()
	if len(list) == 0 {
		return
	}

	Selected += n
	if Selected < 0 {
		Selected = len(list) - 1
	} else if Selected >= len(list) {
		Selected = 0
	}

	// keep the selection inside the visible window
	visibleHeight := height - reservedRows
	TopIndex = min(TopIndex, Selected)
	TopIndex = max(TopIndex, Selected-visibleHeight+1)
}

// put cursor on named entry and scroll it into view
// unknown names are ignored
func jumpTo(name string) {
	list := CurrentList()
	for i, n := range list {
		if n == name {
			Selected = i
			visibleHeight := height - reservedRows
			TopIndex = min(TopIndex, Selected)
			TopIndex = max(TopIndex, Selected-visibleHeight+1)
			return
		}
	}
}
