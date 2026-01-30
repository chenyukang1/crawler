/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	crawler "github.com/chenyukang1/crawler/internal"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/recipes"
)

var (
	showList bool
	recipe   string
)

// recipeCmd  represents the example command
var recipeCmd = &cobra.Command{
	Use:   "recipe",
	Short: "crawler pre-configed recipes",
	Run: func(cmd *cobra.Command, args []string) {
		if showList {
			for _, r := range recipes.List() {
				fmt.Println(r)
			}
			return
		}
		if recipe != "" {
			r := recipes.Get(recipe)
			if r == nil {
				fmt.Printf("Recipe %s not found, try again.\n", recipe)
				os.Exit(1)
			}
			r.Run(crawler.Get(), &spider.GlobalRegistry)
		}
	},
}

func InitRecipe() {
	rootCmd.AddCommand(recipeCmd)

	recipeCmd.Flags().BoolVarP(&showList, "list", "l", false, "show recipe list")
	recipeCmd.Flags().StringVarP(&recipe, "run", "r", "", "run specified recipe")
}
