/*
Copyright © 2025 Bastian Dietrich <bastian@gudd-it.de>
*/
package cmd

import (
	"fmt"
	"go-smtp-relay/internal/proxy"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "smtp-relay",
	Short: "Lightweight local SMTP relay that forwards messages to an upstream SMTP server",
	Long: `smtp-relay starts a local SMTP server and forwards incoming emails to a configured upstream SMTP server.
It supports SMTP authentication and optionally overwriting the sender address.`,
	Run: proxy.Run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/etc/smtp-relay.yaml", "config file")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// Preferred kebab-case flags
	rootCmd.Flags().String("listen-address", "127.0.0.1", "Address/interface to bind the local SMTP listener to (e.g. 0.0.0.0, 127.0.0.1)")
	rootCmd.Flags().Int("listen-port", 25, "Port to listen on for the local SMTP server")
	rootCmd.Flags().String("upstream-host", "", "Hostname or IP of the upstream SMTP server to relay to")
	rootCmd.Flags().Int("upstream-port", 0, "Port of the upstream SMTP server (e.g. 587)")
	rootCmd.Flags().String("upstream-user", "", "Username for authenticating to the upstream SMTP server")
	rootCmd.Flags().String("upstream-password", "", "Password for authenticating to the upstream SMTP server")
	rootCmd.Flags().String("overwrite-sender", "", "Optional: overwrite the envelope sender address for relayed messages")

	// Bind preferred flags to Viper config keys
	_ = viper.BindPFlag("listen.address", rootCmd.Flags().Lookup("listen-address"))
	_ = viper.BindPFlag("listen.port", rootCmd.Flags().Lookup("listen-port"))
	_ = viper.BindPFlag("upstream.host", rootCmd.Flags().Lookup("upstream-host"))
	_ = viper.BindPFlag("upstream.port", rootCmd.Flags().Lookup("upstream-port"))
	_ = viper.BindPFlag("upstream.user", rootCmd.Flags().Lookup("upstream-user"))
	_ = viper.BindPFlag("upstream.password", rootCmd.Flags().Lookup("upstream-password"))
	_ = viper.BindPFlag("overwrite.sender", rootCmd.Flags().Lookup("overwrite-sender"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.SetEnvPrefix("SMTP_RELAY")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
