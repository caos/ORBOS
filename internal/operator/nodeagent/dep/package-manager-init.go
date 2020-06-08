package dep

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func (p *PackageManager) debSpecificUpdatePackages() error {
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()
	cmd := exec.Command("apt-get", "--assume-yes", "update")
	cmd.Stderr = errBuf
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "updating deb packages failed with stderr %s", errBuf.String())
	}

	return p.debbasedInstall(
		&Software{Package: "apt-transport-https"},
		&Software{Package: "gnupg2"},
		&Software{Package: "software-properties-common"})
}

func (p *PackageManager) remSpecificUpdatePackages() error {

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("yum", "update", "-y")
	cmd.Stderr = errBuf
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "updating yum packages failed with stderr %s", errBuf.String())
	}

	if err := p.rembasedInstall(&Software{Package: "yum-utils"}); err != nil {
		return err
	}

	errBuf.Reset()
	cmd = exec.Command("package-cleanup", "--cleandupes", "-y")
	cmd.Stderr = errBuf
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "cleaning up duplicates failed with stderr %s", errBuf.String())
	}

	return p.rembasedInstall(
		&Software{Package: "yum-versionlock"},
		&Software{Package: "firewalld"},
	)
}
