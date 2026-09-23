package filter

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/dimkarp93/md-libs/markdown"
)

type HeadFilter struct {
	Level int
	Name  string
}

var headsEntryRe = regexp.MustCompile(`^h([1-6]):(.+)$`)

func ParseHeads(spec string) ([]HeadFilter, error) {
	var filters []HeadFilter
	for _, entry := range strings.Split(spec, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if m := headsEntryRe.FindStringSubmatch(strings.ToLower(entry)); m != nil {
			level, _ := strconv.Atoi(m[1])
			name := strings.TrimSpace(entry[strings.Index(entry, ":")+1:])
			filters = append(filters, HeadFilter{Level: level, Name: name})
			continue
		}
		filters = append(filters, HeadFilter{Level: 0, Name: entry})
	}
	return filters, nil
}

func Matches(b markdown.Block, filters []HeadFilter) bool {
	if b.Kind != markdown.Heading {
		return false
	}
	plain := markdown.PlainText(b.Text)
	for _, f := range filters {
		if (f.Level == 0 || f.Level == b.Level) && strings.EqualFold(plain, f.Name) {
			return true
		}
	}
	return false
}

func ByHeads(blocks []markdown.Block, filters []HeadFilter, hideMatched bool) []markdown.Block {
	var out []markdown.Block
	includeLevel := -1
	for _, b := range blocks {
		matched := false
		if b.Kind == markdown.Heading {
			if includeLevel != -1 && b.Level <= includeLevel {
				includeLevel = -1
			}
			matched = Matches(b, filters)
			if includeLevel == -1 && matched {
				includeLevel = b.Level
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
