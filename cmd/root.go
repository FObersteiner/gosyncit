/*
Copyright © 2023 Florian Obersteiner <f.obersteiner@posteo.de>

License: see LICENSE in the root directory of the repo.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version    = "0.0.17" // see CHANGELOG.md
	verbose    bool       // global option
	cfgFile    string     // global option
	dryRun     bool       // global option
	noCleanDst bool       // option for copy and mirror
	skipHidden bool       // option for mirror and sync
	// SFTP-specific
	port             int
	reverseDirection bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "gosyncit",
	Version: version,
	Short:   "copy, mirror and sync directories",
	Long: `Copy, mirror and sync directories.

Made with cobra CLI library for Go.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was an error while executing the CLI : '%s'\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gosyncit.toml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

		// Search config in home directory with name ".gosyncit" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("toml")
		viper.SetConfigName(".gosyncit")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func verboseprint(a ...any) {
	if verbose {
		fmt.Println(a...)
	}
}

func verboseprintf(format string, a ...any) {
	if verbose {
		fmt.Printf(format, a...)
	}
}
