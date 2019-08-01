package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/torukita/godb-app/db"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Truncate table in sampledb",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ConnectDB(); err != nil {
			log.Error(err)
			return
		}
		db.DeleteMemos()
		if db.CountMemo() == 0 {
			log.Info("Truncate is done.\n")
		} else {
			log.Error("Truncate is not done.\n")
		}
	},
}

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "prepare table and ...",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("prepare command ...")
		if err := ConnectDB(); err != nil {
			log.Error(err)
			return
		}
		if err := db.CreateTable(); err != nil {
			log.Error(err)
			return
		}
		log.Info("prepare is done")
	},
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "drop table and ...",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("cleanup command ...")
		if err := ConnectDB(); err != nil {
			log.Error(err)
			return
		}
		if err := db.DropTable(); err != nil {
			log.Error(err)
			return
		}
		log.Info("cleanup is done")
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(prepareCmd)
	rootCmd.AddCommand(cleanupCmd)
}
