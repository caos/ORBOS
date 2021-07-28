package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	orb2 "github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/internal/secret/operators"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/orb"
	"github.com/caos/orbos/pkg/secret"
)

var ErrNotInitialized = errors.New("cluster is probably not yet initialized")

func Init(
	monitor mntr.Monitor,
	orbConfig *orb.Orb,
	gitClient *git.Client,
	kubeconfig string,
	gitops bool,
	clone bool,
	omitK8sClient bool,
) (*kubernetes.Client, error) {

	var kc string
	orbConfigIsIncompleteErr := orb.IsComplete(orbConfig)
	if orbConfigIsIncompleteErr != nil && gitops {
		return nil, orbConfigIsIncompleteErr
	}

	if orbConfigIsIncompleteErr == nil && gitops {
		if clone {
			if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
				return nil, err
			}

			if err := gitClient.Clone(); err != nil {
				return nil, err
			}
		}
		if gitClient.Exists(git.OrbiterFile) && !omitK8sClient {
			orbTree, err := gitClient.ReadTree(git.OrbiterFile)
			if err != nil {
				return nil, fmt.Errorf("failed to get tree for %s: %w", git.OrbiterFile, err)
			}

			orbDef, err := orb2.ParseDesiredV0(orbTree)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", git.OrbiterFile, err)
			}

			for clustername, _ := range orbDef.Clusters {
				path := strings.Join([]string{"orbiter", clustername, "kubeconfig", "encrypted"}, ".")

				kc, err = secret.Read(
					nil,
					path,
					operators.GetAllSecretsFunc(monitor, false, gitops, gitClient, nil, orbConfig),
				)
				if err != nil || kc == "" {
					if err == nil {
						return nil, ErrNotInitialized
					}
					return nil, fmt.Errorf(err.Error()+": %w", ErrNotInitialized)
				}
			}
		}
	}

	if omitK8sClient {
		return nil, nil
	}

	if kc == "" {
		value, err := ioutil.ReadFile(kubeconfig)
		if err != nil {
			return nil, mntr.ToUserError(err)
		}
		kc = string(value)
	}

	return kubernetes.NewK8sClient(monitor, &kc)
}
