package orbiter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/caos/orbiter/internal/git"
)

func JoinPath(base string, append ...string) string {
	for _, item := range append {
		base = fmt.Sprintf("%s.%s", base, item)
	}
	return base
}

func ReadSecret(gitClient *git.Client, adapt AdaptFunc, path string) (string, error) {

	secret, _, err := findSecret(gitClient, adapt, path)
	if err != nil {
		return "", err
	}

	return secret.Value, nil
}

func WriteSecret(gitClient *git.Client, adapt AdaptFunc, path, value string) error {

	secret, tree, err := findSecret(gitClient, adapt, path)
	if err != nil {
		return err
	}

	secret.Value = value

	return pushSecretsFunc(gitClient, tree)()
}

func findSecret(gitClient *git.Client, adapt AdaptFunc, path string) (*Secret, *Tree, error) {
	treeDesired, treeSecrets, err := parse(gitClient)
	if err != nil {
		return nil, nil, err
	}

	_, _, secrets, err := adapt(treeDesired, treeSecrets, &Tree{})
	if err != nil {
		return nil, nil, err
	}

	if path != "" {
		sec, err := exactSecret(secrets, path)
		return sec, treeSecrets, err
	}

	items := make([]string, 0, len(secrets))
	for key := range secrets {
		items = append(items, key)
	}

	sort.Slice(items, func(i, j int) bool {
		iDots := strings.Count(items[i], ".")
		jDots := strings.Count(items[j], ".")
		return iDots < jDots || iDots == jDots && items[i] < items[j]
	})

	prompt := promptui.Select{
		Label: "Select Secret",
		Items: items,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return nil, nil, err
	}

	sec, err := exactSecret(secrets, result)
	return sec, treeSecrets, err
}

func exactSecret(secrets map[string]*Secret, path string) (*Secret, error) {
	secret, ok := secrets[path]
	if !ok {
		return nil, fmt.Errorf("Secret %s not found", path)
	}
	return secret, nil
}
