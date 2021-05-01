package chore

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Run(cmd *exec.Cmd) error {
	cmd.Stderr = os.Stderr
	cmd.Env = append(cmd.Env, os.Environ()...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("executing %s failed: %s", strings.Join(cmd.Args, " "), err.Error())
	}
	return nil
}

func Orbctl(debug, skipRebuild bool) (func() *exec.Cmd, error) {

	noop := func() *exec.Cmd { return nil }

	if skipRebuild {
		return runOrbctlCmd(debug), nil
	}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	outBuf := new(bytes.Buffer)
	cmd.Stdout = outBuf
	if err := Run(cmd); err != nil {
		return noop, err
	}

	version := strings.TrimSpace(strings.Replace(outBuf.String(), "heads/", "", 1)) + "-dev"

	cmd = exec.Command("git", "rev-parse", "HEAD")
	outBuf = new(bytes.Buffer)
	cmd.Stdout = outBuf
	if err := Run(cmd); err != nil {
		return noop, err
	}

	commit := strings.TrimSpace(outBuf.String())

	files, err := filepath.Glob("./cmd/chore/gen-executables/*.go")
	if err != nil {
		panic(err)
	}
	args := []string{"run", "-race"}
	args = append(args, files...)
	args = append(args,
		"--version", version,
		"--commit", commit,
		"--githubclientid", os.Getenv("GITHUBOAUTHCLIENTID"),
		"--githubclientsecret", os.Getenv("GITHUBOAUTHCLIENTSECRET"),
		"--orbctl", "./artifacts",
		"--dev",
	)
	if debug {
		args = append(args, "--debug")
	}
	cmd = exec.Command("go", args...)
	cmd.Stdout = os.Stderr
	// gen-executables
	if err := Run(cmd); err != nil {
		return noop, err
	}

	files, err = filepath.Glob("./cmd/chore/gen-charts/*.go")
	if err != nil {
		panic(err)
	}
	args = []string{"build", "-o", "./artifacts/gen-charts"}
	args = append(args, files...)
	cmd = exec.Command("go", args...)
	cmd.Stdout = os.Stderr
	cmd.Env = []string{"CGO_ENABLED=0", "GOOS=linux"}
	// gen-charts
	if err := Run(cmd); err != nil {
		return noop, err
	}

	return runOrbctlCmd(debug), nil
}

func runOrbctlCmd(debug bool) func() *exec.Cmd {
	bin := "./artifacts/orbctl-Linux-x86_64"
	return func() *exec.Cmd {
		if debug {
			return exec.Command("dlv", "exec", "--api-version", "2", "--headless", "--listen", "127.0.0.1:2345", bin, "--")
		}
		return exec.Command(bin)
	}
}
