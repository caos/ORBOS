package centos

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/mntr"
)

func Ensurer(ctx context.Context, monitor mntr.Monitor, open []string) nodeagent.FirewallEnsurer {
	return nodeagent.FirewallEnsurerFunc(func(desired common.Firewall) (common.FirewallCurrent, func() error, error) {
		ensurers := make([]func() error, 0)
		current := make(common.FirewallCurrent, 0)

		if desired.Zones == nil {
			desired.Zones = make(map[string]*common.Zone, 0)
		}

		// Ensure that all runtime config made in the previous iteration becomes permanent.
		if _, err := runFirewallCommand(ctx, monitor, "--runtime-to-permanent"); err != nil {
			return current, nil, err
		}

		for name, _ := range desired.Zones {
			currentZone, ensureFunc, err := ensureZone(ctx, monitor, name, desired, open)
			if err != nil {
				return current, nil, err
			}
			current = append(current, currentZone)
			if ensureFunc != nil {
				ensurers = append(ensurers, ensureFunc)
			}
		}

		_, inactiveErr := runCommand(ctx, monitor, "systemctl", "is-active", "firewalld")
		if inactiveErr == nil && len(ensurers) == 0 {
			monitor.Debug("Not changing firewall")
			return current, nil, nil
		}

		current.Sort()

		return current, func() (err error) {
			monitor.Debug("Ensuring firewall")
			for _, ensurer := range ensurers {
				if err := ensurer(); err != nil {
					return err
				}
			}
			return nil
		}, nil
	})
}

func ensureZone(ctx context.Context, monitor mntr.Monitor, zoneName string, desired common.Firewall, open []string) (*common.ZoneDesc, func() error, error) {
	current := &common.ZoneDesc{
		Name:       zoneName,
		Interfaces: []string{},
		Services:   []*common.Service{},
		FW:         []*common.Allowed{},
	}

	ifaces, err := getInterfaces(ctx, monitor, zoneName)
	if err != nil {
		return current, nil, err
	}
	current.Interfaces = ifaces

	sources, err := getSources(ctx, monitor, zoneName)
	if err != nil {
		return current, nil, err
	}
	current.Sources = sources

	ensureMasquerade, err := getEnsureMasquerade(ctx, monitor, zoneName, current, desired)
	if err != nil {
		return current, nil, err
	}

	addPorts, removePorts, err := getAddAndRemovePorts(ctx, monitor, zoneName, current, desired.Ports(zoneName), open)
	if err != nil {
		return current, nil, err
	}

	ensureIfaces, removeIfaces, err := getEnsureAndRemoveInterfaces(zoneName, current, desired)
	if err != nil {
		return current, nil, err
	}

	addSources, removeSources, err := getAddAndRemoveSources(monitor, zoneName, current, desired)
	if err != nil {
		return current, nil, err
	}

	ensureTarget, err := getEnsureTarget(ctx, monitor, zoneName)
	if err != nil {
		return current, nil, err
	}

	monitor.WithFields(map[string]interface{}{
		"open":  strings.Join(addPorts, ";"),
		"close": strings.Join(removePorts, ";"),
	}).Debug("firewall changes determined")

	if len(addPorts) == 0 &&
		len(removePorts) == 0 &&
		len(addSources) == 0 &&
		len(removeSources) == 0 &&
		len(ensureIfaces) == 0 &&
		len(removeIfaces) == 0 &&
		len(ensureTarget) == 0 {
		return current, nil, nil
	}

	zoneNameCopy := zoneName
	return current, func() (err error) {

		if len(ensureTarget) > 0 {

			monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", ensureTarget, zoneNameCopy))
			if err := ensure(ctx, monitor, ensureTarget, zoneNameCopy); err != nil {
				return err
			}

			// this is the only property that needs a firewall reload
			_, err := runFirewallCommand(ctx, monitor, "--reload")
			return err
		}

		if ensureMasquerade != "" {
			monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", ensureMasquerade, zoneNameCopy))
			if err := ensure(ctx, monitor, []string{ensureMasquerade}, zoneNameCopy); err != nil {
				return err
			}
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", removeIfaces, zoneNameCopy))
		if err := ensure(ctx, monitor, removeIfaces, zoneNameCopy); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", ensureIfaces, zoneNameCopy))
		if err := ensure(ctx, monitor, ensureIfaces, zoneNameCopy); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", removeSources, zoneNameCopy))
		if err := ensure(ctx, monitor, removeSources, zoneNameCopy); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", addSources, zoneNameCopy))
		if err := ensure(ctx, monitor, addSources, zoneNameCopy); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", removePorts, zoneNameCopy))
		if err := ensure(ctx, monitor, removePorts, zoneNameCopy); err != nil {
			return err
		}

		monitor.Debug(fmt.Sprintf("Ensuring part of firewall with %s in zone %s", addPorts, zoneNameCopy))
		return ensure(ctx, monitor, addPorts, zoneNameCopy)
	}, nil
}

func ensure(ctx context.Context, monitor mntr.Monitor, changes []string, zone string) error {
	if changes == nil || len(changes) == 0 {
		return nil
	}

	if _, err := runCommand(ctx, monitor, "systemctl", "enable", "firewalld"); err != nil {
		return err
	}

	if _, err := runCommand(ctx, monitor, "systemctl", "start", "firewalld"); err != nil {
		return err
	}

	return changeFirewall(ctx, monitor, changes, zone)
}

func changeFirewall(ctx context.Context, monitor mntr.Monitor, changes []string, zone string) error {
	if len(changes) == 0 {
		return nil
	}

	_, err := runFirewallCommand(ctx, monitor.Verbose(), append([]string{"--zone", zone}, changes...)...)
	return err
}

func listFirewall(ctx context.Context, monitor mntr.Monitor, zone string, arg string) ([]string, error) {

	out, err := runFirewallCommand(ctx, monitor, "--zone", zone, arg)
	return strings.Fields(out), err
}

func runFirewallCommand(ctx context.Context, monitor mntr.Monitor, args ...string) (string, error) {
	return runCommand(ctx, monitor, "firewall-cmd", args...)
}

func runCommand(ctx context.Context, monitor mntr.Monitor, binary string, args ...string) (string, error) {

	outBuf := new(bytes.Buffer)
	defer outBuf.Reset()
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stderr = errBuf
	cmd.Stdout = outBuf

	fullCmd := fmt.Sprintf("'%s'", strings.Join(cmd.Args, "' '"))
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(`running %s failed with stderr %s: %w`, fullCmd, errBuf.String(), err)
	}

	stdout := outBuf.String()
	if monitor.IsVerbose() {
		fmt.Println(fullCmd)
		fmt.Println(stdout)
	}

	return strings.TrimSuffix(stdout, "\n"), nil
}
