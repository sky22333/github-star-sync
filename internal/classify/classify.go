package classify

import (
	"sort"
	"strings"

	"github-star-sync/internal/config"
	"github-star-sync/internal/model"
)

const otherKey = "其他"

// Group classifies repos with topic-frequency categories and language fallback.
func Group(repos []model.Repo, cfg config.ClassifyConfig) []model.Category {
	if len(repos) == 0 {
		return nil
	}

	topicCount := map[string]int{}
	for _, r := range repos {
		seen := map[string]struct{}{}
		for _, t := range r.Topics {
			t = normalizeKey(t)
			if t == "" {
				continue
			}
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			topicCount[t]++
		}
	}

	type ranked struct {
		name  string
		count int
	}
	var candidates []ranked
	for name, n := range topicCount {
		if n >= cfg.MinCount {
			candidates = append(candidates, ranked{name, n})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].count != candidates[j].count {
			return candidates[i].count > candidates[j].count
		}
		return candidates[i].name < candidates[j].name
	})
	if len(candidates) > cfg.MaxCategories {
		candidates = candidates[:cfg.MaxCategories]
	}

	valid := map[string]struct{}{}
	for _, c := range candidates {
		valid[c.name] = struct{}{}
	}

	buckets := map[string][]model.Repo{}
	display := map[string]string{}
	var order []string
	add := func(key, label string, r model.Repo) {
		if _, ok := buckets[key]; !ok {
			order = append(order, key)
			display[key] = label
		}
		buckets[key] = append(buckets[key], r)
	}

	for _, r := range repos {
		key, label := pickCategory(r, valid, topicCount, cfg.Fallback)
		add(key, label, r)
	}

	sort.SliceStable(order, func(i, j int) bool {
		a, b := order[i], order[j]
		if a == otherKey {
			return false
		}
		if b == otherKey {
			return true
		}
		if len(buckets[a]) != len(buckets[b]) {
			return len(buckets[a]) > len(buckets[b])
		}
		return display[a] < display[b]
	})

	out := make([]model.Category, 0, len(order))
	for _, key := range order {
		reposIn := buckets[key]
		sortRepos(reposIn, cfg.SortWithin)
		out = append(out, model.Category{
			Name:  display[key],
			Repos: reposIn,
		})
	}
	return out
}

func pickCategory(r model.Repo, valid map[string]struct{}, topicCount map[string]int, fallback string) (key, label string) {
	best := ""
	bestN := -1
	for _, t := range r.Topics {
		t = normalizeKey(t)
		if t == "" {
			continue
		}
		if _, ok := valid[t]; !ok {
			continue
		}
		if topicCount[t] > bestN || (topicCount[t] == bestN && t < best) {
			best = t
			bestN = topicCount[t]
		}
	}
	if best != "" {
		return best, titleTopic(best)
	}
	if fallback == "language" {
		lang := strings.TrimSpace(r.Language)
		if lang != "" {
			return normalizeKey(lang), lang
		}
	}
	return otherKey, otherKey
}

func normalizeKey(t string) string {
	return strings.ToLower(strings.TrimSpace(t))
}

func titleTopic(name string) string {
	if name == otherKey {
		return otherKey
	}
	parts := strings.Split(name, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

func sortRepos(repos []model.Repo, mode string) {
	sort.SliceStable(repos, func(i, j int) bool {
		a, b := repos[i], repos[j]
		switch mode {
		case "starred_at":
			if !a.StarredAt.Equal(b.StarredAt) {
				return a.StarredAt.After(b.StarredAt)
			}
		case "name":
			return strings.ToLower(a.FullName) < strings.ToLower(b.FullName)
		default:
			if a.Stars != b.Stars {
				return a.Stars > b.Stars
			}
		}
		return strings.ToLower(a.FullName) < strings.ToLower(b.FullName)
	})
}
