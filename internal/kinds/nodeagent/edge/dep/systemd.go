package dep

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/logging"
)

type SystemD struct {
	logger logging.Logger
}

func NewSystemD(logger logging.Logger) *SystemD {
	return &SystemD{logger}
}

func (s *SystemD) Disable(binary string) error {

	var errBuf bytes.Buffer
	cmd := exec.Command("systemctl", "stop", binary)
	cmd.Stderr = &errBuf
	if s.logger.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "stopping %s by systemd failed with stderr %s", binary, errBuf.String())
	}

	errBuf.Reset()
	cmd = exec.Command("systemctl", "disable", binary)
	cmd.Stderr = &errBuf
	if s.logger.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "disabling %s by systemd failed with stderr %s", binary, errBuf.String())
	}

	return nil
}

func (s *SystemD) Start(binary string) error {
	var errBuf bytes.Buffer
	cmd := exec.Command("systemctl", "restart", binary)
	cmd.Stderr = &errBuf
	return errors.Wrapf(cmd.Run(), "restarting %s from systemd failed with stderr %s", binary, errBuf.String())
}

func (s *SystemD) Enable(binary string) error {

	var errBuf bytes.Buffer
	cmd := exec.Command("systemctl", "daemon-reload")
	cmd.Stderr = &errBuf
	if s.logger.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "reloading systemd in order to use new %s failed with stderr %s", binary, errBuf.String())
	}

	errBuf.Reset()
	cmd = exec.Command("systemctl", "enable", binary)
	cmd.Stderr = &errBuf
	if s.logger.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	return errors.Wrapf(cmd.Run(), "configuring systemd to manage %s failed with stderr %s", binary, errBuf.String())
}
