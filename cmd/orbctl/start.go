package main

import (
	"log"
	"net/http"

	"github.com/caos/orbos/internal/ctrlcrd"
	"github.com/caos/orbos/internal/ctrlgitops"
	"github.com/caos/orbos/pkg/kubernetes"
	orbcfg "github.com/caos/orbos/pkg/orb"
	"github.com/spf13/cobra"
)

func StartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start an operator",
		Long:  `Pass exactly one of orbiter, boom or networking"`,
		Args:  cobra.MinimumNArgs(1),
	}
}

func StartOrbiter(getRv GetRootValues) *cobra.Command {
	var (
		verbose bool
		recur   bool
		deploy  bool
		pprof   bool
		cmd     = &cobra.Command{
			Use:   "orbiter",
			Short: "Launch an orbiter",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&recur, "recur", true, "Ensure the desired state continously")
	flags.BoolVar(&deploy, "deploy", true, "Ensure Orbiter deployment continously")
	flags.BoolVar(&pprof, "pprof", false, "Start pprof to analyse memory usage")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {

		rv := getRv("start", "orbiter", map[string]interface{}{"recur": recur, "depoy": deploy, "pprof": pprof})
		defer rv.ErrFunc(err)

		if err := orbcfg.IsComplete(rv.OrbConfig); err != nil {
			return err
		}

		if err := rv.GitClient.Configure(rv.OrbConfig.URL, []byte(rv.OrbConfig.Repokey)); err != nil {
			return err
		}

		orbiterConfig := &ctrlgitops.OrbiterConfig{
			Recur:         recur,
			Deploy:        deploy,
			Verbose:       verbose,
			Version:       version,
			OrbConfigPath: rv.OrbConfig.Path,
			GitCommit:     gitCommit,
		}

		if pprof {
			go func() {
				log.Println(http.ListenAndServe("localhost:6060", nil))
			}()
		}

		return ctrlgitops.Orbiter(rv.Ctx, monitor, orbiterConfig, rv.GitClient)
	}
	return cmd
}

func StartBoom(getRv GetRootValues) *cobra.Command {
	var (
		metricsAddr string
		cmd         = &cobra.Command{
			Use:   "boom",
			Short: "Launch a boom",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&metricsAddr, "metrics-addr", "", "The address the metric endpoint binds to.")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {

		rv := getRv("start", "boom", map[string]interface{}{"metrics-addr": metricsAddr != ""})
		defer rv.ErrFunc(err)

		monitor.Info("Takeoff Boom")

		if rv.Gitops {
			return ctrlgitops.Boom(monitor, rv.OrbConfig.Path, version)
		}
		return ctrlcrd.Start(monitor, version, "/boom", metricsAddr, "", ctrlcrd.Boom)
	}
	return cmd
}

func StartNetworking(getRv GetRootValues) *cobra.Command {
	var (
		metricsAddr string
		cmd         = &cobra.Command{
			Use:   "networking",
			Short: "Launch a networking operator",
			Long:  "Ensures a desired state of networking for an application",
		}
	)
	flags := cmd.Flags()
	flags.StringVar(&metricsAddr, "metrics-addr", "", "The address the metric endpoint binds to.")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		rv := getRv("start", "networking-operator", map[string]interface{}{"metrics-addr": metricsAddr != ""})
		defer rv.ErrFunc(err)

		monitor := rv.Monitor
		orbConfig := rv.OrbConfig

		monitor.Info("Takeoff Networking")

		if rv.Gitops {

			k8sClient, err := kubernetes.NewK8sClientWithPath(monitor, rv.Kubeconfig)
			if err != nil {
				return err
			}
			return ctrlgitops.Networking(monitor, orbConfig.Path, k8sClient, &version)
		}
		return ctrlcrd.Start(monitor, version, "/boom", metricsAddr, rv.Kubeconfig, ctrlcrd.Networking)
	}
	return cmd
}
