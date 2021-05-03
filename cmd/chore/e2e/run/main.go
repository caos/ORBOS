package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/pkg/orb"
)

func main() {

	var (
		unpublished bool
		orbconfig   string
		//		ghToken     string
		//		testcase    string
		graphiteURL string
		graphiteKey string
		from        int
		cleanup     bool
	)

	const (
		unpublishedDefault = false
		unpublishedUsage   = "Test all unpublished branches"
		orbDefault         = "~/.orb/config"
		orbUsage           = "Path to the orbconfig file which points to the orb the end-to-end testing should be performed on"
		//		githubTokenDefault  = ""
		//		githubTokenKeyUsage = "Personal access token with repo scope for github.com/caos/orbos"
		//		testcaseDefault     = ""
		//		testcaseUsage       = "Testcase identifier"
		graphiteURLDefault = ""
		graphiteURLUsage   = "https://<your-subdomain>.hosted-metrics.grafana.net/metrics"
		graphiteKeyDefault = ""
		graphiteKeyUsage   = "your api key from grafana.net -- should be editor role"
		fromDefault        = 1
		fromUsage          = "step to continue e2e tests from"
		cleanupDefault     = true
		cleanupUsage       = "destroy orb after tests are done"
	)

	flag.BoolVar(&unpublished, "unpublished", unpublishedDefault, unpublishedUsage)
	flag.BoolVar(&unpublished, "u", unpublishedDefault, unpublishedUsage+" (shorthand)")
	flag.StringVar(&orbconfig, "orbconfig", orbDefault, orbUsage)
	flag.StringVar(&orbconfig, "f", orbDefault, orbUsage+" (shorthand)")
	//	flag.StringVar(&ghToken, "github-access-token", githubTokenDefault, githubTokenKeyUsage)
	//	flag.StringVar(&ghToken, "t", githubTokenDefault, githubTokenKeyUsage+" (shorthand)")
	//	flag.StringVar(&testcase, "testcase", testcaseDefault, testcaseUsage)
	//	flag.StringVar(&testcase, "c", testcaseDefault, testcaseUsage+" (shorthand)")
	flag.StringVar(&graphiteURL, "graphiteurl", graphiteURLDefault, graphiteURLUsage)
	flag.StringVar(&graphiteURL, "g", graphiteURLDefault, graphiteURLUsage+" (shorthand)")
	flag.StringVar(&graphiteKey, "graphitekey", graphiteKeyDefault, graphiteKeyUsage)
	flag.StringVar(&graphiteKey, "k", graphiteKeyDefault, graphiteKeyUsage+" (shorthand)")
	flag.BoolVar(&cleanup, "cleanup", cleanupDefault, cleanupUsage)
	flag.BoolVar(&cleanup, "c", cleanupDefault, cleanupUsage+" (shorthand)")
	flag.IntVar(&from, "from", fromDefault, fromUsage)
	flag.IntVar(&from, "s", fromDefault, fromUsage)

	flag.Parse()

	fmt.Printf("unpublished=%t\n", unpublished)
	fmt.Printf("orbconfig=%s\n", orbconfig)
	fmt.Printf("graphiteurl=%s\n", graphiteURL)
	fmt.Printf("cleanup=%t\n", cleanup)
	fmt.Printf("from=%d\n", from)

	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		panic(err)
	}

	branch := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(string(out)), "refs/"), "heads/")

	testFunc := runFunc
	/*
		if ghToken != "" {
			testFunc = func(branch string) error {
				return github(trimBranch(branch), ghToken, strings.ToLower(testcase), runFunc)(orbconfig)
			}
		}
	*/
	if graphiteURL != "" {

		orbCfg, err := orb.ParseOrbConfig(helpers.PruneHome(orbconfig))
		if err != nil {
			panic(err)
		}

		if err := orb.IsComplete(orbCfg); err != nil {
			panic(err)
		}

		testFunc = graphite(
			strings.ToLower(strings.Split(strings.Split(orbCfg.URL, "/")[1], ".")[0]),
			graphiteURL,
			graphiteKey,
			trimBranch(branch),
			runFunc)
	}

	if err := testFunc(strings.ReplaceAll(strings.TrimPrefix(branch, "origin/"), ".", "-"), orbconfig, from, cleanup)(); err != nil {
		panic(err)
	}
	return
}

func trimBranch(ref string) string {
	return strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(ref), "refs/"), "heads/")
}
