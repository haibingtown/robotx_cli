package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	baseURL string
	apiKey  string
)

var rootCmd = &cobra.Command{
	Use:   "robotx",
	Short: "RobotX CLI - Deploy AI applications to RobotX platform",
	Long: `RobotX CLI is a command-line tool for deploying AI applications to the RobotX platform.
It provides a simple interface for AI agents to create, build, and deploy projects.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.robotx.yaml)")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "", "RobotX server base URL")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "RobotX API key")

	viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".robotx")
	}

	viper.SetEnvPrefix("ROBOTX")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
