package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func runFunc(ctx context.Context, logger promtail.Client, orb, branch, orbconfig string, from uint8, cleanup bool) func() error {
	return func() (err error) {

		newOrbctl, err := buildOrbctl(ctx, logger, orbconfig)
		if err != nil {
			return err
		}

		if cleanup {
			defer func() {
				// context is probably cancelled as program is terminating so we create a new one here
				if cuErr := destroyTestFunc(context.Background(), logger)(newOrbctl, nil); cuErr != nil {

					original := ""
					if err != nil {
						original = fmt.Sprintf(": %s", err.Error())
					}

					err = fmt.Errorf("cleaning up after end-to-end test failed: %w%s", cuErr, original)
				}
			}()
		}

		kubeconfig, err := ioutil.TempFile("", "kubeconfig-*")
		if err != nil {
			return err
		}
		if err := kubeconfig.Close(); err != nil {
			return err
		}

		readKubeconfig, deleteKubeconfig := readKubeconfigFunc(ctx, logger, orb, kubeconfig.Name())
		defer deleteKubeconfig()

		branchParts := strings.Split(branch, "/")
		branch = branchParts[len(branchParts)-1:][0]

		return seq(logger, newOrbctl, configureKubectl(kubeconfig.Name()), from, 5, readKubeconfig,
			/*  1 */ retry(3, writeInitialDesiredStateTest(ctx, logger, orb, branch)),
			/*  2 */ retry(3, destroyTestFunc(ctx, logger)),
			/*  3 */ retry(3, bootstrapTestFunc(ctx, logger, orb, 4)),
			/*  4 */ ensureORBITERTest(ctx, logger, orb, 5, 10*time.Minute, isEnsured(orb, 3, 3, "v1.18.8")),
			/*  5 */ retry(3, patchTestFunc(ctx, logger, fmt.Sprintf("clusters.%s.spec.controlplane.nodes", orb), "1")),
			/*  6 */ retry(3, patchTestFunc(ctx, logger, fmt.Sprintf("clusters.%s.spec.workers.0.nodes", orb), "2")),
			/*  7 */ ensureORBITERTest(ctx, logger, orb, 8, 5*time.Minute, isEnsured(orb, 1, 2, "v1.18.8")),
			/*  8 */ retry(3, patchTestFunc(ctx, logger, fmt.Sprintf("clusters.%s.spec.versions.kubernetes", orb), "v1.21.0")),
			/*  9 */ ensureORBITERTest(ctx, logger, orb, 10, 45*time.Minute, isEnsured(orb, 1, 2, "v1.21.0")),
		)
	}
}

type testFunc func(newOrbctlCommandFunc, newKubectlCommandFunc) error

func seq(logger promtail.Client, orbctl newOrbctlCommandFunc, kubectl newKubectlCommandFunc, from, readKubeconfigFrom uint8, readKubeconfigFunc func(orbctl newOrbctlCommandFunc) (err error), fns ...testFunc) error {

	var at uint8
	for _, fn := range fns {
		at++
		if at < from {
			logger.Infof("\033[1;32mSkipping step %d\033[0m\n", at)
			continue
		}

		if at >= readKubeconfigFrom {
			if err := readKubeconfigFunc(orbctl); err != nil {
				return err
			}
		}

		fnName := fmt.Sprintf("%s (%d. in stack)", runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(), at)

		if err := fn(orbctl, kubectl); err != nil {
			return fmt.Errorf("%s failed: %w", fnName, err)
		}
		logger.Infof("\033[1;32m%s succeeded\033[0m\n", fnName)
	}
	return nil
}

func retry(count uint8, fn func(newOrbctlCommandFunc, newKubectlCommandFunc) error) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(newOrbctl newOrbctlCommandFunc, newKubectl newKubectlCommandFunc) error {
		return try(count, newOrbctl, newKubectl, fn)
	}
}

func try(count uint8, newOrbctl newOrbctlCommandFunc, newKubectl newKubectlCommandFunc, fn func(newOrbctl newOrbctlCommandFunc, newKubectl newKubectlCommandFunc) error) error {
	err := fn(newOrbctl, newKubectl)
	exitErr := &exec.ExitError{}
	if err != nil &&
		// Don't retry if context timed out or was cancelled
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, context.DeadlineExceeded) &&
		// Don't retry if commands context timed out or was cancelled
		(!errors.As(err, &exitErr) ||
			!exitErr.Sys().(syscall.WaitStatus).Signaled()) &&
		count > 0 {
		return try(count-1, newOrbctl, newKubectl, fn)
	}
	return err
}
