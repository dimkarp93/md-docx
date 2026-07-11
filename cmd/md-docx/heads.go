package main

import (
	"regexp"
	"strconv"
	"strings"
)

type headFilter struct {
	level int
	name  string
}

var headsEntryRe = regexp.MustCompile(`^h([1-6]):(.+)$`)

func parseHeadFilters(spec string) ([]headFilter, error) {
	var filters []headFilter
	for _, entry := range strings.Split(spec, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if m := headsEntryRe.FindStringSubmatch(strings.ToLower(entry)); m != nil {
			level, _ := strconv.Atoi(m[1])
			name := strings.TrimSpace(entry[strings.Index(entry, ":")+1:])
			filters = append(filters, headFilter{level: level, name: name})
			continue
		}
		filters = append(filters, headFilter{level: 0, name: entry})
	}
	return filters, nil
}

func matchesHeadFilter(b block, filters []headFilter) bool {
	if b.kind != blockHeading {
		return false
	}
	plain := plainText(b.text)
	for _, f := range filters {
		if (f.level == 0 || f.level == b.level) && strings.EqualFold(plain, f.name) {
			return true
		}
	}
	return false
}

func filterByHeads(blocks []block, filters []headFilter, hideMatched bool) []block {
	var out []block
	includeLevel := -1
	for _, b := range blocks {
		matched := false
		if b.kind == blockHeading {
			if includeLevel != -1 && b.level <= includeLevel {
				includeLevel = -1
			}
			matched = matchesHeadFilter(b, filters)
			if includeLevel == -1 && matched {
				includeLevel = b.level
			}
		}
		if includeLevel == -1 {
			continue
		}
		if matched && hideMatched {
			continue
		}
		out = append(out, b)
	}
	return out
}
