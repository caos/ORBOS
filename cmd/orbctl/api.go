package main

import (
	"errors"

	"github.com/caos/orbos/pkg/kubernetes/cli"

	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/pkg/git"

	boomapi "github.com/caos/orbos/internal/operator/boom/api"
	"github.com/caos/orbos/internal/operator/orbiter"
	orbadapter "github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/pkg/labels"
	"github.com/spf13/cobra"
)

func APICommand(getRv GetRootValues) *cobra.Command {
	var (
		cmd = &cobra.Command{
			Use:   "api",
			Short: "Upgrade the yml-files to the newest version",
			Long:  "Upgrade the yml-files to the newest version",
		}
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {

		rv := getRv("api", "", nil)
		defer rv.ErrFunc(err)

		if !rv.Gitops {
			return mntr.ToUserError(errors.New("api command is only supported with the --gitops flag"))
		}

		if _, err := cli.Init(rv.Monitor, rv.OrbConfig, rv.GitClient, rv.Kubeconfig, rv.Gitops, true, true); err != nil {
			return err
		}

		var desireds []git.GitDesiredState
		if rv.GitClient.Exists(git.OrbiterFile) {
			_, _, _, migrate, desired, _, _, err := orbiter.Adapt(rv.GitClient, monitor, make(chan struct{}), orbadapter.AdaptFunc(
				labels.NoopOperator("ORBOS"),
				rv.OrbConfig,
				gitCommit,
				true,
				false,
				rv.GitClient,
			))
			if err != nil {
				return err
			}

			if migrate {
				desireds = append(desireds, git.GitDesiredState{
					Desired: desired,
					Path:    git.OrbiterFile,
				})
			}
		}
		if rv.GitClient.Exists(git.BoomFile) {

			desired, err := rv.GitClient.ReadTree(git.BoomFile)
			if err != nil {
				return err
			}

			toolset, migrate, _, _, err := boomapi.ParseToolset(desired)
			if err != nil {
				return err
			}
			if migrate {
				desired.Parsed = toolset
				desireds = append(desireds, git.GitDesiredState{
					Desired: desired,
					Path:    git.BoomFile,
				})
			}
		}
		if len(desireds) > 0 {
			return rv.GitClient.PushGitDesiredStates(monitor, "migrate apis", desireds)
		}
		return nil
	}
	return cmd
}
