/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chenyukang1/crawler/recipes"
)

var showList bool

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
	},
}

func InitRecipe() {
	rootCmd.AddCommand(recipeCmd)
	recipeCmd.Flags().BoolVarP(&showList, "list", "l", false, "show recipe list")
}
