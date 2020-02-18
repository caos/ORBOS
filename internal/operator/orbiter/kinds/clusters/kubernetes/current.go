package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/orbiter"
)

type CurrentCluster struct {
	Status   string
	Machines map[string]*Machine `yaml:"machines"`
}

type Current struct {
	Common  orbiter.Common `yaml:",inline"`
	Current CurrentCluster
}

type KubernetesStatus struct {
	Joined      bool
	Online      bool
	Maintaining bool
}

type NodeAgentStatus struct {
	Running bool
	Commit  string
}

type Machine struct {
	Kubernetes KubernetesStatus
	NodeAgent  NodeAgentStatus
	Metadata   MachineMetadata `yaml:",inline"`
}

type Versions struct {
	NodeAgent  string
	Kubernetes string
}

type MachineMetadata struct {
	Tier     Tier
	Provider string
	Pool     string
	Group    string `yaml:",omitempty"`
}

type Tier string

const (
	Controlplane Tier = "controlplane"
	Workers      Tier = "workers"
)
