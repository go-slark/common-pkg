package main

import (
	"fmt"
	"github.com/smallfish-root/common-pkg/xcmd/hack/proto"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hack",
	Short: "hack plugin",
	Long:  "hack plugin",
}

func init() {
	rootCmd.AddCommand(proto.CreateCmd)
	rootCmd.AddCommand(proto.InstallCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Print(err)
		return
	}
}
