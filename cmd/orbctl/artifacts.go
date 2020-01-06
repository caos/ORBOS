package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	mach "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/logging"
)

func ensureArtifacts(logger logging.Logger, secrets *operator.Secrets, orb *Orb, oneoff bool, secretsNamespace, orbiterversion string, boomversion string) error {

	l := logger.WithFields(map[string]interface{}{
		"cluster": secretsNamespace,
	})

	logger.WithFields(map[string]interface{}{
		"orbiter": orbiterversion,
		"boom":    boomversion,
	}).Debug("Ensuring artifacts")

	if orbiterversion == "" && boomversion == "" {
		return nil
	}

	kc, err := secrets.Read(secretsNamespace + "_kubeconfig")
	if err != nil {
		return nil
	}

	kcStr := string(kc)
	client := k8s.New(logger, &kcStr)

	orbfile, err := yaml.Marshal(orb)
	if err != nil {
		return err
	}

	if err := client.ApplyNamespace(&core.Namespace{
		ObjectMeta: mach.ObjectMeta{
			Name: "caos-system",
			Labels: map[string]string{
				"name": "caos-system",
			},
		},
	}); err != nil {
		return err
	}

	if err := client.ApplySecret(&core.Secret{
		ObjectMeta: mach.ObjectMeta{
			Name:      "caos",
			Namespace: "caos-system",
		},
		StringData: map[string]string{
			"orbconfig": string(orbfile),
		},
	}); err != nil {
		return err
	}

	if err := client.ApplySecret(&core.Secret{
		ObjectMeta: mach.ObjectMeta{
			Name:      "public-github-packages",
			Namespace: "caos-system",
		},
		Type: core.SecretTypeDockerConfigJson,
		StringData: map[string]string{
			core.DockerConfigJsonKey: `{
		"auths": {
				"docker.pkg.github.com": {
						"auth": "aW1ncHVsbGVyOmU2NTAxMWI3NDk1OGMzOGIzMzcwYzM5Zjg5MDlkNDE5OGEzODBkMmM="
				}
		}
}`,
		},
	}); err != nil {
		return err
	}
	if orbiterversion != "" {
		if err := client.ApplyDeployment(&apps.Deployment{
			ObjectMeta: mach.ObjectMeta{
				Name:      "orbiter",
				Namespace: "caos-system",
			},
			Spec: apps.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &mach.LabelSelector{
					MatchLabels: map[string]string{
						"app": "orbiter",
					},
				},
				Template: core.PodTemplateSpec{
					ObjectMeta: mach.ObjectMeta{
						Labels: map[string]string{
							"app": "orbiter",
						},
					},
					Spec: core.PodSpec{
						NodeSelector: map[string]string{
							"node-role.kubernetes.io/master": "",
						},
						Tolerations: []core.Toleration{{
							Key:      "node-role.kubernetes.io/master",
							Operator: "Equal",
							Value:    "",
							Effect:   "NoSchedule",
						}},
						ImagePullSecrets: []core.LocalObjectReference{{
							Name: "public-github-packages",
						}},
						Containers: []core.Container{{
							Name:            "orbiter",
							ImagePullPolicy: core.PullIfNotPresent,
							Image:           "docker.pkg.github.com/caos/orbiter/orbiter:" + orbiterversion,
							Command:         []string{"/orbctl", "--orbconfig", "/etc/orbiter/orbconfig", "takeoff", "--recur"},
							VolumeMounts: []core.VolumeMount{{
								Name:      "keys",
								ReadOnly:  true,
								MountPath: "/etc/orbiter",
							}},
						}},
						Volumes: []core.Volume{{
							Name: "keys",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: "caos",
									Optional:   boolPtr(false),
								},
							},
						}},
					},
				},
			},
		}); err != nil {
			return err
		}
		l.WithFields(map[string]interface{}{
			"version": orbiterversion,
		}).Debug("Orbiter deployment ensured")

		if oneoff {
			os.Exit(0)
		}
	}

	if boomversion == "" {
		return nil
	}

	if err := client.ApplyRole(&rbac.Role{
		ObjectMeta: mach.ObjectMeta{
			Name:      "boom-leader-election-role",
			Namespace: "caos-system",
		},
		Rules: []rbac.PolicyRule{{
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
			Verbs: []string{
				"get",
				"list",
				"watch",
				"create",
				"update",
				"patch",
				"delete",
			},
		}, {
			APIGroups: []string{""},
			Resources: []string{"configmaps/status"},
			Verbs: []string{
				"get",
				"update",
				"patch",
			},
		}, {
			APIGroups: []string{""},
			Resources: []string{"events"},
			Verbs:     []string{"create"},
		}},
	}); err != nil {
		return err
	}

	if err := client.ApplyClusterRole(&rbac.ClusterRole{
		ObjectMeta: mach.ObjectMeta{
			Name: "boom-manager-role",
		},
		Rules: []rbac.PolicyRule{{
			APIGroups: []string{""},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}, {
			APIGroups: []string{"admissionregistration.k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}, {
			APIGroups: []string{"apiextensions.k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}, {
			APIGroups: []string{"apiregistration.k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}, {
			APIGroups: []string{"apps"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}, {
			APIGroups: []string{"batch"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}, {
			APIGroups: []string{"extensions"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}, {
			APIGroups: []string{"logging.banzaicloud.io"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}, {
			APIGroups: []string{"monitoring.coreos.com"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}, {
			APIGroups: []string{"policy"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}, {
			APIGroups: []string{"rbac.authorization.k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}, {
			APIGroups: []string{"toolsets.boom.caos.ch"},
			Resources: []string{"toolsets"},
			Verbs: []string{
				"create",
				"delete",
				"get",
				"list",
				"patch",
				"update",
				"watch",
			},
		}, {
			APIGroups: []string{"toolsets.boom.caos.ch"},
			Resources: []string{"toolsets/status"},
			Verbs: []string{
				"get",
				"patch",
				"update",
			},
		}},
	}); err != nil {
		return err
	}

	if err := client.ApplyClusterRole(&rbac.ClusterRole{
		ObjectMeta: mach.ObjectMeta{
			Name: "boom-proxy-role",
		},
		Rules: []rbac.PolicyRule{{
			APIGroups: []string{"authentication.k8s.io"},
			Resources: []string{"tokenreviews"},
			Verbs:     []string{"create"},
		}, {
			APIGroups: []string{"authorization.k8s.io"},
			Resources: []string{"subjectaccessreviews"},
			Verbs:     []string{"create"},
		}},
	}); err != nil {
		return err
	}

	if err := client.ApplyRoleBinding(&rbac.RoleBinding{
		ObjectMeta: mach.ObjectMeta{
			Namespace: "caos-system",
			Name:      "boom-leader-election-rolebinding",
		},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "boom-leader-election-role",
		},
		Subjects: []rbac.Subject{{
			Kind:      "ServiceAccount",
			Name:      "default",
			Namespace: "caos-system",
		}},
	}); err != nil {
		return err
	}
	if err := client.ApplyClusterRoleBinding(&rbac.ClusterRoleBinding{
		ObjectMeta: mach.ObjectMeta{
			Name: "boom-manager-rolebinding",
		},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "boom-manager-role",
		},
		Subjects: []rbac.Subject{{
			Kind:      "ServiceAccount",
			Name:      "default",
			Namespace: "caos-system",
		}},
	}); err != nil {
		return err
	}
	if err := client.ApplyClusterRoleBinding(&rbac.ClusterRoleBinding{
		ObjectMeta: mach.ObjectMeta{
			Name: "boom-proxy-rolebinding",
		},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "boom-proxy-role",
		},
		Subjects: []rbac.Subject{{
			Kind:      "ServiceAccount",
			Name:      "default",
			Namespace: "caos-system",
		}},
	}); err != nil {
		return err
	}

	err = client.ApplyDeployment(&apps.Deployment{
		ObjectMeta: mach.ObjectMeta{
			Name:      "boom",
			Namespace: "caos-system",
			Labels: map[string]string{
				"app": "boom",
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &mach.LabelSelector{
				MatchLabels: map[string]string{
					"app": "boom",
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: mach.ObjectMeta{
					Labels: map[string]string{
						"app": "boom",
					},
				},
				Spec: core.PodSpec{
					ImagePullSecrets: []core.LocalObjectReference{{
						Name: "public-github-packages",
					}},
					Containers: []core.Container{{
						Name:            "boom",
						ImagePullPolicy: core.PullIfNotPresent,
						Image:           fmt.Sprintf("docker.pkg.github.com/caos/boom/boom:%s", boomversion),
						Command:         []string{"/boom"},
						Args: []string{
							"--metrics-addr", "127.0.0.1:8080",
							"--enable-leader-election",
							"--git-orbconfig", "/secrets/tools-secret/orbconfig",
							"--git-crd-path", "crd/boom.yml",
						},
						VolumeMounts: []core.VolumeMount{{
							Name:      "tools-secret",
							ReadOnly:  true,
							MountPath: "/secrets/tools-secret",
						}},
					}},
					TerminationGracePeriodSeconds: int64Ptr(10),
					Volumes: []core.Volume{{
						Name: "tools-secret",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: "caos",
							},
						},
					}},
				},
			},
		},
	})
	if err == nil {
		l.WithFields(map[string]interface{}{
			"version": boomversion,
		}).Debug("Boom deployment ensured")
	}
	return err
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }
