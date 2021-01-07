package gce

import (
	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	uuid "github.com/satori/go.uuid"
)

func destroy(context *context, delegates map[string]interface{}) error {

	return helpers.Fanout([]func() error{
		func() error {
			destroyLB, err := queryLB(context, nil)
			if err != nil {
				return err
			}
			return destroyLB()
		},
		func() error {
			pools, err := context.machinesService.ListPools()
			if err != nil {
				return err
			}
			var delFuncs []func() error
			for _, pool := range pools {
				machines, err := context.machinesService.List(pool)
				if err != nil {
					return err
				}
				for _, machine := range machines {
					delFuncs = append(delFuncs, machine.Remove)
				}
			}
			if err := helpers.Fanout(delFuncs)(); err != nil {
				return err
			}

			return helpers.Fanout([]func() error{
				func() error {
					var deleteDisks []func() error

					deleteMonitor := context.monitor.WithField("type", "persistent disk")

					for kind, delegate := range delegates {
						volumes, ok := delegate.([]infra.Volume)
						if ok {
							for idx := range volumes {
								diskName := volumes[idx].Name
								deleteDisks = append(deleteDisks, deleteDiskFunc(context, deleteMonitor.WithField("id", diskName), kind, diskName))
							}
						}
					}
					return helpers.Fanout(deleteDisks)()
				},
				func() error {
					_, deleteFirewalls, err := queryFirewall(context, nil)
					if err != nil {
						return err
					}
					return destroyNetwork(context, deleteFirewalls)
				},
			})()
		},
	})()
}

func deleteDiskFunc(context *context, monitor mntr.Monitor, kind, id string) func() error {
	return func() error {
		return operateFunc(
			func() { monitor.Debug("Removing resource") },
			computeOpCall(context.client.Disks.Delete(context.projectID, context.desired.Zone, id).RequestId(uuid.NewV1().String()).Do),
			func() error { monitor.Info("Resource removed"); return nil },
		)()
	}
}
