package container

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/pkg/signal"
	"github.com/sirupsen/logrus"
)

// ForwardAllSignals forwards signals to the container
func ForwardAllSignals(ctx context.Context, cli command.Cli, cid string) chan os.Signal {
	sigc := make(chan os.Signal, 128)
	signal.CatchAll(sigc)
	go func() {
		for s := range sigc {
			if s == signal.SIGCHLD || s == signal.SIGPIPE {
				continue
			}

			// In go1.14+, the go runtime issues SIGURG as an interupt to support pre-emptable system calls on Linux.
			// Since we can't forward that along we'll check that here.
			if isRuntimeSig(s) {
				continue
			}
			var sig string
			for sigStr, sigN := range signal.SignalMap {
				if sigN == s {
					sig = sigStr
					break
				}
			}
			if sig == "" {
				fmt.Fprintf(cli.Err(), "Unsupported signal: %v. Discarding.\n", s)
				continue
			}

			if err := cli.Client().ContainerKill(ctx, cid, sig); err != nil {
				logrus.Debugf("Error sending signal: %s", err)
			}
		}
	}()
	return sigc
}
