package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/torukita/godb-app/db"
)

var Version = "0.0.6"
var cfgFile string
var envName string
var logLevel string

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}

var rootCmd = &cobra.Command{
	Use:     "",
	Short:   "",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		return

		if err := ConnectDB(); err != nil {
			log.Error(err)
			return
		}
		db.AddMemo("root command", "")
		db.DumpMemo()
		fmt.Printf("Count=%d\n", db.CountMemo())
	},
}

func ConnectDB() error {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)

	if err := db.Load(cfgFile); err != nil {
		return err
	}
	if err := db.SetEnv(envName); err != nil {
		return err
	}
	if err := db.Connect(); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./dbconf.yml", "config file")
	rootCmd.PersistentFlags().StringVarP(&envName, "env", "e", "development", "environment in config file")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "panic|fatal|error|warn|info|debug|trace")
}
