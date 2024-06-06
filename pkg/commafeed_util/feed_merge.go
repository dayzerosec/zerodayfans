package commafeed_util

import "github.com/dayzerosec/zerodayfans/pkg/commafeed"

// MergeEntries will take two []Entry arrays that are already sorted by `.Date` and merge them into a single []Entry
// array that is also sorted by `.Date`
func MergeEntries(a []commafeed.Entry, b []commafeed.Entry) []commafeed.Entry {
	var merged []commafeed.Entry

	i, j, k := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i].Date > b[j].Date {
			merged = append(merged, a[i])
			i++
		} else {
			merged = append(merged, b[j])
			j++
		}
		k++
	}

	// Since we don't know when list still has elements just try to merge remainder of both
	for i < len(a) {
		merged = append(merged, a[i])
		i++
		k++
	}

	for j < len(b) {
		merged = append(merged, b[j])
		j++
		k++
	}

	return merged
}
