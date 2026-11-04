package cmd

import (
	"os"
	"time"

	"github.com/atlas-depin/cygnus/cmd/types"
	"github.com/atlas-depin/cygnus/core"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	startMaxAttempts = 3
	startRetryDelay  = 5 * time.Second
)

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the storage provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			logLevel, err := cmd.Flags().GetString(types.FlagLogLevel)
			if err != nil {
				return err
			}

			cw := zerolog.ConsoleWriter{Out: os.Stderr}
			log.Logger = zerolog.New(cw).With().Timestamp().Caller().Logger()

			switch logLevel {
			case "debug":
				log.Logger = log.Logger.Level(zerolog.DebugLevel)
			case "info":
				log.Logger = log.Logger.Level(zerolog.InfoLevel)
			case "warn":
				log.Logger = log.Logger.Level(zerolog.WarnLevel)
			case "error":
				log.Logger = log.Logger.Level(zerolog.ErrorLevel)
			}

			var startErr error
			for attempt := 1; attempt <= startMaxAttempts; attempt++ {
				// Build a fresh App for every attempt: Start() releases the
				// storage engine and gRPC connection when it returns, so a
				// failed instance can never be started again.
				app, err := core.NewApp(home)
				if err != nil {
					startErr = err
				} else {
					startErr = app.Start()
				}

				if startErr == nil {
					return nil
				}

				log.Err(startErr).Msg("Failed to start Cygnus.")
				if attempt == startMaxAttempts {
					break
				}

				log.Info().Msgf(
					"Attempting restart in %d seconds (attempt %d of %d)...",
					int(startRetryDelay.Seconds()), attempt+1, startMaxAttempts,
				)
				time.Sleep(startRetryDelay)
			}

			return startErr
		},
	}

	return cmd
}
