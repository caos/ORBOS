package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/caos/orbos/pkg/kubernetes/cli"

	"github.com/caos/orbos/pkg/git"

	"github.com/caos/orbos/pkg/labels"

	"github.com/spf13/cobra"

	"github.com/caos/orbos/internal/operator/orbiter"
	orbadapter "github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
)

func TeardownCommand(getRv GetRootValues) *cobra.Command {

	var (
		cmd = &cobra.Command{
			Use:   "teardown",
			Short: "Tear down an Orb",
			Long:  "Destroys a whole Orb",
			Aliases: []string{
				"shoot",
				"destroy",
				"devastate",
				"annihilate",
				"crush",
				"bulldoze",
				"total",
				"smash",
				"decimate",
				"kill",
				"trash",
				"wipe-off-the-map",
				"pulverize",
				"take-apart",
				"destruct",
				"obliterate",
				"disassemble",
				"explode",
				"blow-up",
			},
		}
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {

		rv := getRv("teardown", "", nil)
		defer rv.ErrFunc(err)

		if !rv.Gitops {
			return errors.New("teardown command is only supported with the --gitops flag and a committed orbiter.yml")
		}

		if _, err := cli.Init(monitor, rv.OrbConfig, rv.GitClient, rv.Kubeconfig, rv.Gitops, true, false); err != nil {
			return err
		}

		if rv.GitClient.Exists(git.OrbiterFile) {
			monitor.WithFields(map[string]interface{}{
				"version": version,
				"commit":  gitCommit,
				"repoURL": rv.OrbConfig.URL,
			}).Info("Destroying Orb")

			fmt.Println("Are you absolutely sure you want to destroy all clusters and providers in this Orb? [y/N]")
			var response string
			fmt.Scanln(&response)

			if !contains([]string{"y", "yes"}, strings.ToLower(response)) {
				monitor.Info("Not touching Orb")
				return nil
			}
			finishedChan := make(chan struct{})
			return orbiter.Destroy(
				monitor,
				rv.GitClient,
				orbadapter.AdaptFunc(
					labels.NoopOperator("ORBOS"),
					rv.OrbConfig,
					gitCommit,
					true,
					false,
					rv.GitClient,
				),
				finishedChan,
			)
		}
		return nil
	}
	return cmd
}

func contains(slice []string, elem string) bool {
	for _, item := range slice {
		if elem == item {
			return true
		}
	}
	return false
}
