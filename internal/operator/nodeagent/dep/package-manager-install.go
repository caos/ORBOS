package dep

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/caos/orbos/mntr"

	"github.com/pkg/errors"
)

func (p *PackageManager) rembasedInstall(installVersion *Software, more ...*Software) error {

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	installPkgs := make([]string, 0)
	for _, sw := range append([]*Software{installVersion}, more...) {

		installedVersion, ok := p.installed[sw.Package]
		if ok && (sw.Version == "" || sw.Version == installedVersion) {
			continue
		}

		if sw.Version == "" {
			installPkgs = append(installPkgs, sw.Package)
			continue
		}

		installPkg := fmt.Sprintf("%s-%s", sw.Package, sw.Version)
		installPkgs = append(installPkgs, installPkg)
		cmd := exec.Command("yum", "versionlock", "delete", sw.Package)
		cmd.Stderr = errBuf
		if p.monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			cmd.Stdout = os.Stdout
		}
		err := cmd.Run()
		stderr := errBuf.String()
		if err != nil && !strings.Contains(stderr, "versionlock delete: no matches") {
			return errors.Wrapf(err, "unlocking package %s failed with stderr %s", sw.Package, stderr)
		}
		errBuf.Reset()

		cmd = exec.Command("yum", "versionlock", "add", "-y", installPkg)
		cmd.Stderr = errBuf
		if p.monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			cmd.Stdout = os.Stdout
		}
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "locking package %s at version %s failed with stderr %s", sw.Package, sw.Version, errBuf.String())
		}
		errBuf.Reset()
	}

	for _, pkg := range installPkgs {
		if err := rembasedInstallPkg(p.monitor, pkg); err != nil {
			return err
		}
	}
	return nil
}

func rembasedInstallPkg(monitor mntr.Monitor, pkg string) error {
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()
	cmd := exec.Command("yum", "install", "-y", pkg)
	cmd.Stderr = errBuf
	if monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	var errStr string
	err := cmd.Run()
	if err != nil {
		errStr = errBuf.String()
		if strings.Contains(errStr, "is already installed") {
			err = nil
		}
	}

	return errors.Wrapf(err, "installing yum package %s failed with stderr %s", pkg, errStr)
}

// TODO: Use lower level apt instead of apt-get?
func (p *PackageManager) debbasedInstall(installVersion *Software, more ...*Software) error {

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	pkgs := make([]string, len(more)+1)
	hold := make([]string, 0)
	for idx, sw := range append([]*Software{installVersion}, more...) {
		pkgs[idx] = sw.Package
		if sw.Version == "" {
			continue
		}
		pkgs[idx] = fmt.Sprintf("%s=%s", sw.Package, sw.Version)
		hold = append(hold, sw.Package)

		cmd := exec.Command("apt-mark", "unhold", sw.Package)
		cmd.Stderr = errBuf
		if p.monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			cmd.Stdout = os.Stdout
		}
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "unholding installed package failed with stderr %s", errBuf.String())
		}
		errBuf.Reset()
	}

	cmd := exec.Command("dpkg", "--configure", "-a")
	cmd.Stderr = errBuf
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "cleaning up dpkg failed with stderr %s", errBuf.String())
	}
	errBuf.Reset()

	cmd = exec.Command("apt-get", append(strings.Fields(
		"--assume-yes --allow-downgrades install -y"), pkgs...)...)
	cmd.Stderr = errBuf
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "installing package failed with stderr %s", errBuf.String())
	}
	errBuf.Reset()

	for _, pkg := range hold {
		cmd = exec.Command("apt-mark", "hold", pkg)
		cmd.Stderr = errBuf
		if p.monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			cmd.Stdout = os.Stdout
		}
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "holding package failed with stderr %s", errBuf.String())
		}
		errBuf.Reset()

		p.monitor.WithFields(map[string]interface{}{
			"package": installVersion.Package,
			"version": installVersion.Version,
		}).Debug("Installed package")
	}
	return nil
}
