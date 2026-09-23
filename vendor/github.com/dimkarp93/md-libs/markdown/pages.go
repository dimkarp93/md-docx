package markdown

import (
	"bufio"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func SplitPages(src string) []string {
	scanner := bufio.NewScanner(strings.NewReader(src))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	isDelim := func(line string) bool {
		return strings.TrimSpace(line) == "---"
	}

	start := 0
	pages := []string{""}
	if len(lines) > 0 && isDelim(lines[0]) {
		closeIdx := -1
		for i := 1; i < len(lines); i++ {
			if isDelim(lines[i]) {
				closeIdx = i
				break
			}
		}
		if closeIdx != -1 {
			pages[0] = strings.Join(lines[1:closeIdx], "\n")
			start = closeIdx + 1
		}
	}

	inCode := false
	var current []string
	for i := start; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCode = !inCode
			current = append(current, line)
			continue
		}
		if !inCode && isDelim(line) {
			pages = append(pages, strings.Join(current, "\n"))
			current = nil
			continue
		}
		current = append(current, line)
	}
	pages = append(pages, strings.Join(current, "\n"))

	return pages
}

func ParsePageRanges(spec string) ([]int, error) {
	seen := map[int]struct{}{}
	for _, tok := range strings.Split(spec, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if strings.Contains(tok, "-") {
			parts := strings.SplitN(tok, "-", 2)
			lo, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid range %q", tok)
			}
			hi, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid range %q", tok)
			}
			if lo > hi {
				return nil, fmt.Errorf("invalid range %q: start > end", tok)
			}
			for i := lo; i <= hi; i++ {
				seen[i] = struct{}{}
			}
			continue
		}
		n, err := strconv.Atoi(tok)
		if err != nil {
			return nil, fmt.Errorf("invalid page %q", tok)
		}
		seen[n] = struct{}{}
	}

	result := make([]int, 0, len(seen))
	for n := range seen {
		result = append(result, n)
	}
	sort.Ints(result)
	return result, nil
}
