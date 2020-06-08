package main

import (
	"fmt"
	"github.com/caos/orbos/internal/utils/orbgit"
	"strings"

	"github.com/spf13/cobra"

	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
)

func TeardownCommand(rv RootValues) *cobra.Command {

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

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, monitor, orbConfig, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		gitClientConf := &orbgit.Config{
			Comitter:  "orbctl",
			Email:     "orbctl@caos.ch",
			OrbConfig: orbConfig,
			Action:    "takeoff",
		}

		gitClient, cleanUp, err := orbgit.NewGitClient(ctx, monitor, gitClientConf)
		defer cleanUp()
		if err != nil {
			return err
		}

		foundOrbiter, err := existsFileInGit(gitClient, "orbiter.yml")
		if err != nil {
			return err
		}
		if foundOrbiter {
			monitor.WithFields(map[string]interface{}{
				"version": version,
				"commit":  gitCommit,
				"repoURL": orbConfig.URL,
			}).Info("Destroying Orb")

			fmt.Println("Are you absolutely sure you want to destroy all clusters and providers in this Orb? [y/N]")
			var response string
			fmt.Scanln(&response)

			if !contains([]string{"y", "yes"}, strings.ToLower(response)) {
				monitor.Info("Not touching Orb")
				return nil
			}
			finishedChan := make(chan bool)
			return orbiter.Destroy(
				monitor,
				gitClient,
				orb.AdaptFunc(
					orbConfig,
					gitCommit,
					true,
					false),
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
