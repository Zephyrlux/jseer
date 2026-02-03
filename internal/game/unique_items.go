package game

import "sync"

type uniqueItemsFile struct {
	UniqueItems []uniqueItemEntry `json:"uniqueItems"`
}

type uniqueItemEntry struct {
	RangeStart int   `json:"rangeStart"`
	RangeEnd   int   `json:"rangeEnd"`
	Items      []int `json:"items"`
}

var uniqueOnce sync.Once

func loadUniqueItemsConfig() {
	defaultRanges := []struct {
		Min int
		Max int
	}{
		{Min: 100001, Max: 199999},
		{Min: 1000, Max: 1300},
	}
	uniqueRanges = defaultRanges
	uniqueIDs = uniqueIDs[:0]

	var cfg uniqueItemsFile
	if !readConfigJSON("unique-items.json", &cfg) {
		return
	}

	uniqueRanges = []struct {
		Min int
		Max int
	}{}
	uniqueIDs = uniqueIDs[:0]
	for _, entry := range cfg.UniqueItems {
		if entry.RangeStart > 0 && entry.RangeEnd >= entry.RangeStart {
			uniqueRanges = append(uniqueRanges, struct {
				Min int
				Max int
			}{Min: entry.RangeStart, Max: entry.RangeEnd})
		}
		for _, id := range entry.Items {
			if id > 0 {
				uniqueIDs = append(uniqueIDs, id)
			}
		}
	}
	if len(uniqueRanges) == 0 {
		uniqueRanges = defaultRanges
	}
}
