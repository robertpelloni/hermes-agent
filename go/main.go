package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "code-cli",
		Short: "Hermes Agent CLI (Go)",
		Long:  `A foundational CLI parser in Go for the Ultimate Agentic Coding Harness.`,
	}

	var chatMessage string
	var chatCmd = &cobra.Command{
		Use:   "chat",
		Short: "Start a chat session",
		Run: func(cmd *cobra.Command, args []string) {
			if chatMessage != "" {
				fmt.Printf("Starting chat with message: %s\n", chatMessage)
			} else {
				fmt.Println("Starting interactive chat session...")
			}
		},
	}
	chatCmd.Flags().StringVarP(&chatMessage, "message", "m", "", "The message to send")

	var configCmd = &cobra.Command{
		Use:   "config [key]",
		Short: "Manage configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				fmt.Printf("Reading config key: %s\n", args[0])
			} else {
				fmt.Println("Opening interactive config editor...")
			}
		},
	}

	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(configCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
