package token

import (
	"fmt"
	"time"

	k8sutil "github.com/Mirantis/mke/pkg/kubernetes"
	"github.com/Mirantis/mke/pkg/util"
	ers "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NewManager(kubeconfig string) (*Manager, error) {
	logrus.Debugf("loading kubeconfig from: %s", kubeconfig)
	client, err := k8sutil.Client(kubeconfig)
	if err != nil {
		return nil, err
	}
	// Create
	return &Manager{
		client: client,
	}, nil
}

type Manager struct {
	client *kubernetes.Clientset
}

// Create creates a new bootstrap token
func (m *Manager) Create(valid time.Duration, role string) (string, error) {
	err := m.ensureTokenRBAC()
	if err != nil {
		return "", ers.Wrapf(err, "failed to ensure presense of bootstrap token related rbac rules")
	}
	tokenID := util.RandomString(6)
	tokenSecret := util.RandomString(16)

	token := fmt.Sprintf("%s.%s", tokenID, tokenSecret)

	data := make(map[string]string)
	data["token-id"] = tokenID
	data["token-secret"] = tokenSecret
	if valid != 0 {
		data["expiration"] = time.Now().Add(valid).UTC().Format(time.RFC3339)
		logrus.Debugf("Set expiry to %s", data["expiration"])
	}

	if role == "worker" {
		data["description"] = "Worker bootstrap token generated by mke"
		data["usage-bootstrap-authentication"] = "true"
		data["usage-bootstrap-signing"] = "true"
	} else {
		data["description"] = "Controller bootstrap token generated by mke"
		data["usage-bootstrap-authentication"] = "false"
		data["usage-bootstrap-signing"] = "false"
		data["usage-controller-join"] = "true"
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("bootstrap-token-%s", tokenID),
			Namespace: "kube-system",
		},
		Type:       v1.SecretTypeBootstrapToken,
		StringData: data,
	}

	_, err = m.client.CoreV1().Secrets("kube-system").Create(secret)
	if err != nil {
		return "", err
	}

	return token, nil
}

/**

kubectl create clusterrolebinding kubelet-bootstrap \
  --clusterrole=system:node-bootstrapper \
  --group=system:bootstrappers

kubectl create clusterrolebinding node-autoapprove-bootstrap \
  --clusterrole=system:certificates.k8s.io:certificatesigningrequests:nodeclient \
  --group=system:bootstrappers

kubectl create clusterrolebinding node-autoapprove-certificate-rotation \
  --clusterrole=system:certificates.k8s.io:certificatesigningrequests:selfnodeclient \
  --group=system:nodes
*/
func (m *Manager) ensureTokenRBAC() error {
	crbs := make([]rbac.ClusterRoleBinding, 3)
	crbs = append(crbs, rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubelet-bootstrap",
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.SchemeGroupVersion.Group,
			Kind:     "ClusterRole",
			Name:     "system:node-bootstrapper",
		},
		Subjects: []rbac.Subject{
			rbac.Subject{
				APIGroup: rbac.SchemeGroupVersion.Group,
				Kind:     "Group",
				Name:     "system:bootstrappers",
			},
		},
	})

	crbs = append(crbs, rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-autoapprove-bootstrap",
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.SchemeGroupVersion.Group,
			Kind:     "ClusterRole",
			Name:     "system:certificates.k8s.io:certificatesigningrequests:nodeclient",
		},
		Subjects: []rbac.Subject{
			rbac.Subject{
				APIGroup: rbac.SchemeGroupVersion.Group,
				Kind:     "Group",
				Name:     "system:bootstrappers",
			},
		},
	})

	crbs = append(crbs, rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-autoapprove-certificate-rotation",
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.SchemeGroupVersion.Group,
			Kind:     "ClusterRole",
			Name:     "system:certificates.k8s.io:certificatesigningrequests:selfnodeclient",
		},
		Subjects: []rbac.Subject{
			rbac.Subject{
				APIGroup: rbac.SchemeGroupVersion.Group,
				Kind:     "Group",
				Name:     "system:nodes",
			},
		},
	})

	// Create everything, ignore if already existing
	for _, crb := range crbs {
		// Check if the CRB already exists
		if _, err := m.client.RbacV1().ClusterRoleBindings().Get(crb.ObjectMeta.Name, metav1.GetOptions{}); errors.IsNotFound(err) {
			_, err := m.client.RbacV1().ClusterRoleBindings().Create(&crb)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
