package cmd

import (
	"os"
	"os/signal"
	"syscall"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/moolen/tdep/pkg/orchestrator"
	"github.com/moolen/tdep/pkg/webhook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// orchestratorCmd represents the json command
var orchestratorCmd = &cobra.Command{
	Use:   "orchestrator",
	Short: "runs the orchestrator",
	Long:  "runs the orchestrator",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		tfeClient, err := tfe.NewClient(&tfe.Config{
			Token: os.Getenv("TFE_TOKEN"),
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Info("starting orchestrator")
		orc, err := orchestrator.New(tfeClient, 4)
		if err != nil {
			log.Fatal(err)
		}
		orc.Start()

		log.Info("starting webhook server")
		srv := webhook.NewReceiverServer(":8080")
		stop := make(chan struct{})

		go func() {
			// wait for signal
			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
			<-signals
			log.Info("stopping webhook server")
			stop <- struct{}{}
		}()

		// blocks until signal is sent
		srv.ListenAndServe(stop)

		log.Info("stopping orchestrator")
		if err := orc.Stop(); err != nil {
			log.Error("orchestrator shutdown error: %v", err)
		}
		log.Info("done with cleanup. exiting")
	},
}

func init() {
	rootCmd.AddCommand(orchestratorCmd)
}
