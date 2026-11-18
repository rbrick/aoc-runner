package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aoc-runner",
	Short: "A utility for making Advent Of Code easier",
	Long: `aoc-runner (Temp name) is a small LeetCode-style platform specifcally for Advent Of Code. 
	 It provides a backend API for fetching puzzles from the AoC website and a service for actually running and testing
	 the code against test-input, and a Web UI with a built-in code-editor and puzzle browser.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aoc-runner.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		appDirectory := filepath.Join(home, ".aoc-runner")

		if _, err := os.Stat(appDirectory); os.IsNotExist(err) {
			cobra.CheckErr(os.Mkdir(appDirectory, os.ModePerm))
		}

		// Search config in home directory with name ".aoc-runner" (without extension).
		viper.AddConfigPath(appDirectory)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")

		// set defaults here
		viper.SetDefault("server.database", filepath.Join(home, "aoc.db"))
		viper.SetDefault("server.host", "localhost:8080")

	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
