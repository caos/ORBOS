package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/orb"
)

func readSecretCommand(rv rootValues) *cobra.Command {

	return &cobra.Command{
		Use:     "readsecret [path]",
		Short:   "Decrypt and print to stdout",
		Args:    cobra.MaximumNArgs(1),
		Example: `orbctl readsecret k8s.kubeconfig > ~/.kube/config`,
		RunE: func(cmd *cobra.Command, args []string) error {

			_, logger, gitClient, orbconfig, errFunc := rv()
			if errFunc != nil {
				return errFunc(cmd)
			}

			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			value, err := orbiter.ReadSecret(
				gitClient,
				orb.AdaptFunc(logger,
					orbconfig,
					gitCommit,
					false,
					false),
				path)
			if err != nil {
				panic(err)
			}
			if _, err := os.Stdout.Write([]byte(value)); err != nil {
				panic(err)
			}
			return nil
		},
	}
}
