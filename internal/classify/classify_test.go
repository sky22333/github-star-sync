package classify

import (
	"strings"
	"testing"

	"github-star-sync/internal/config"
	"github-star-sync/internal/model"
)

func TestGroupTopicFreq(t *testing.T) {
	repos := []model.Repo{
		{FullName: "a/p1", Topics: []string{"proxy", "vpn"}, Stars: 10, Language: "Go"},
		{FullName: "a/p2", Topics: []string{"proxy"}, Stars: 20, Language: "Rust"},
		{FullName: "a/p3", Topics: []string{"proxy", "rust"}, Stars: 5, Language: "Rust"},
		{FullName: "a/g1", Topics: nil, Stars: 3, Language: "Go"},
		{FullName: "a/g2", Topics: []string{"go"}, Stars: 8, Language: "Go"},
		{FullName: "a/x1", Topics: []string{"rare"}, Stars: 1, Language: ""},
	}
	cats := Group(repos, config.ClassifyConfig{
		MaxCategories: 12,
		MinCount:      2,
		Fallback:      "language",
		SortWithin:    "stars",
	})
	if len(cats) < 2 {
		t.Fatalf("expected multiple categories, got %#v", cats)
	}
	foundProxy := false
	goCount := 0
	for _, c := range cats {
		if c.Name == "Proxy" {
			foundProxy = true
			if len(c.Repos) < 2 {
				t.Fatalf("proxy should have multiple repos: %#v", c.Repos)
			}
			if c.Repos[0].Stars < c.Repos[1].Stars {
				t.Fatalf("expected stars desc order")
			}
		}
		if c.Name == "Rare" {
			t.Fatalf("rare topic should not become category with min_count=2")
		}
		if strings.EqualFold(c.Name, "go") {
			goCount++
		}
	}
	if !foundProxy {
		t.Fatalf("expected Proxy category, got %#v", cats)
	}
	if goCount != 1 {
		t.Fatalf("expected single Go category after key normalize, got %d in %#v", goCount, cats)
	}
}
