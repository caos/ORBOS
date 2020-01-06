package cri

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/middleware"
	"github.com/caos/orbiter/logging"
)

type Installer interface {
	isCRI()
	adapter.Installer
}

// TODO: Add support for containerd, cri-o, ...
type criDep struct {
	logger                    logging.Logger
	os                        dep.OperatingSystemMajor
	manager                   *dep.PackageManager
	dockerVersionPrunerRegexp *regexp.Regexp
	systemd                   *dep.SystemD
}

// New returns a dependency that implements the kubernetes container runtime interface
func New(logger logging.Logger, os dep.OperatingSystemMajor, manager *dep.PackageManager, systemd *dep.SystemD) Installer {
	return &criDep{logger, os, manager, regexp.MustCompile(`\d+\.\d+\.\d+`), systemd}
}

func (criDep) Is(other adapter.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (c criDep) isCRI() {}

func (c criDep) String() string { return "Container Runtime" }

func (s *criDep) Equals(other adapter.Installer) bool {
	_, ok := other.(*criDep)
	return ok
}

func (c *criDep) Current() (pkg operator.Package, err error) {
	installed, err := c.manager.CurrentVersions("docker-ce")
	if err != nil {
		return pkg, err
	}
	version := ""
	for _, pkg := range installed {
		version = fmt.Sprintf("%s %s %s", version, pkg.Package, "v"+c.dockerVersionPrunerRegexp.FindString(pkg.Version))
	}
	pkg.Version = strings.TrimSpace(version)
	return pkg, nil
}

func (c *criDep) Ensure(uninstall operator.Package, install operator.Package) (bool, error) {

	fields := strings.Fields(install.Version)
	if len(fields) != 2 {
		return false, errors.Errorf("Container runtime must have the form [runtime] [version], but got %s", install)
	}

	if fields[0] != "docker-ce" {
		return false, errors.New("Only docker-ce is supported yet")
	}

	version := strings.TrimLeft(fields[1], "v")

	switch c.os.OperatingSystem {
	case dep.Ubuntu:
		return c.ensureUbuntu(fields[0], version)
	case dep.CentOS:
		return c.ensureCentOS(fields[0], version)
	}
	return false, errors.Errorf("Operating %s system is not supported", c.os)
}
