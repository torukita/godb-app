package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/torukita/godb-app/db"
	"time"
)

var runmode runConfig

type runConfig struct {
	Interval  time.Duration
	MaxCount  int
	ReConnect bool
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run insert table",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ConnectDB(); err != nil {
			log.Error(err)
			return
		}
		defer db.Close()
		interval := runmode.Interval
		max := runmode.MaxCount

		count := 0
		for {
			count++
			start := time.Now()
			if err := db.AddMemo(fmt.Sprintf("data-%d", count), ""); err != nil {
				log.Errorf("[%s] Failed data (%d). it started at [%s] %s", time.Now().Format(time.StampMilli), count, start.Format(time.StampMilli), err)
				break
			}
			if max == 1 {
				log.Info("Added one data")
			}
			log.Debugf("Added data (%d)", count)
			time.Sleep(interval)
			if count == max {
				break
			}
		}

		if runmode.ReConnect && count != max { // failure happened during running
			reconnect := 0
			for {
				reconnect++
				if err := db.Connect(); err != nil {
					log.Error(fmt.Sprintf("[%s] %s", time.Now().Format(time.StampMilli), err))
					//					time.Sleep(1 * time.Second)
					continue
				}
				log.Warnf("[%s] db connection recovered again by reconnect=%d times", time.Now().Format(time.StampMilli), reconnect)
				if err := db.AddMemo(fmt.Sprintf("reconnect data-%d", count), ""); err != nil {
					log.Errorf("[%s] Failed data (%d) again. %s", time.Now().Format(time.StampMilli), count, err)
				} else {
					log.Infof("[%s] Added data (%d) after reconnection", time.Now().Format(time.StampMilli), count)
					break
				}
			}
		}
	},
}

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump table",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ConnectDB(); err != nil {
			log.Error(err)
			return
		}
		if err := db.DumpMemo(); err != nil {
			log.Error(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(dumpCmd)
	runCmd.Flags().DurationVarP(&runmode.Interval, "interval", "i", 0*time.Second, "inteval sec (5s)")
	runCmd.Flags().IntVarP(&runmode.MaxCount, "num", "n", 1, "max count")
	runCmd.Flags().BoolVar(&runmode.ReConnect, "reconnect", false, "enable reconnect db")
}
