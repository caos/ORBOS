package main

import (
	"errors"
	"fmt"

	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/pkg/cfg"

	"github.com/caos/orbos/pkg/kubernetes/cli"

	"github.com/caos/orbos/pkg/git"

	"github.com/caos/orbos/pkg/orb"
	"github.com/spf13/cobra"
)

func ConfigCommand(getRv GetRootValues) *cobra.Command {

	var (
		newMasterKey string
		newRepoURL   string
		cmd          = &cobra.Command{
			Use:     "configure",
			Short:   "Configures and reconfigures an orb",
			Long:    "Generates missing ssh keys and other secrets where it makes sense",
			Aliases: []string{"reconfigure", "config", "reconfig"},
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&newMasterKey, "masterkey", "", "Reencrypts all secrets")
	flags.StringVar(&newRepoURL, "repourl", "", "Configures the repository URL")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {

		rv := getRv("configure", "", map[string]interface{}{"masterkey": newMasterKey != "", "newRepoURL": newRepoURL})
		defer rv.ErrFunc(err)

		if !rv.Gitops {
			return mntr.ToUserError(errors.New("configure command is only supported with the --gitops flag"))
		}

		if err := orb.Reconfigure(rv.Ctx, rv.Monitor, rv.OrbConfig, newRepoURL, newMasterKey, rv.GitClient, githubClientID, githubClientSecret); err != nil {
			return err
		}

		var uninitialized bool
		k8sClient, err := cli.Init(rv.Monitor, rv.OrbConfig, rv.GitClient, rv.Kubeconfig, rv.Gitops, false, false)
		if err != nil {
			if !errors.Is(err, cli.ErrNotInitialized) {
				return err
			}
			uninitialized = true
			err = nil
		}

		unmanagedOperators := []git.DesiredFile{git.DatabaseFile, git.ZitadelFile}
		for i := range unmanagedOperators {
			operatorFile := unmanagedOperators[i]
			if rv.GitClient.Exists(operatorFile) {
				return mntr.ToUserError(fmt.Errorf("found %s in git repository. Please use zitadelctl's configure command", operatorFile))
			}
		}

		if !uninitialized {
			if err := cfg.ApplyOrbconfigSecret(
				rv.OrbConfig,
				k8sClient,
				rv.Monitor,
			); err != nil {
				return err
			}
		}

		return cfg.ConfigureOperators(
			rv.GitClient,
			rv.OrbConfig.Masterkey,
			cfg.ORBOSConfigurers(
				rv.Monitor,
				rv.OrbConfig,
				rv.GitClient,
			))
	}
	return cmd
}
