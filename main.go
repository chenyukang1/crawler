/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/chenyukang1/crawler/cmd"
)

func main() {
	cmd.InitRoot()
	cmd.InitRecipe()
	cmd.Execute()
}
