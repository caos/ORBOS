package main

import (
	"errors"

	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"

	"github.com/spf13/cobra"
)

func RebootCommand(getRv GetRootValues) *cobra.Command {
	return &cobra.Command{
		Use:   "reboot",
		Short: "Gracefully reboot machines",
		Long:  "Pass machine ids as arguments, omit arguments for selecting machines interactively",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			node := ""
			if len(args) > 0 {
				node = args[0]
			}

			rv := getRv("reboot", "", map[string]interface{}{"node": node})
			defer rv.ErrFunc(err)

			if !rv.Gitops {
				return mntr.ToUserError(errors.New("reboot command is only supported with the --gitops flag and a committed orbiter.yml"))
			}

			return requireMachines(monitor, rv.GitClient, rv.OrbConfig, args, func(machine infra.Machine) (required bool, require func(), unrequire func()) {
				return machine.RebootRequired()
			})
		},
	}
}
