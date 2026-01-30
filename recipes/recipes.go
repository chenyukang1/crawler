package recipes

import (
	crawler "github.com/chenyukang1/crawler/internal"
	"github.com/chenyukang1/crawler/internal/spider"
)

type Recipe interface {
	Run(app *crawler.App, registry *spider.Registry)
}

var registry map[string]Recipe = make(map[string]Recipe)

func List() []string {
	var recipes []string
	for k := range registry {
		recipes = append(recipes, k)
	}
	return recipes
}

func Get(name string) Recipe {
	var (
		recipe Recipe
		ok     bool
	)
	if recipe, ok = registry[name]; !ok {
		return nil
	}
	return recipe
}
