package defaultaddons

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"github.com/weaveworks/eksctl/pkg/kubernetes"
)

// LoadAsset return embedded manifest as a runtime.Object
func LoadAsset(name, ext string) (*metav1.List, error) {
	data, err := Asset(name + "." + ext)
	if err != nil {
		return nil, errors.Wrapf(err, "decoding embedded manifest for %q", name)
	}
	list, err := kubernetes.NewList(data)
	if err != nil {
		return nil, errors.Wrapf(err, "loading individual resources from manifest for %q", name)
	}
	return list, nil
}

// useRegionalImage updates the template spec to a region-aware container image.
func useRegionalImage(spec *corev1.PodTemplateSpec, prefixFormat, region, suffix string) error {
	image := spec.Spec.Containers[0].Image
	parts := strings.Split(image, ":")

	if len(parts) != 2 {
		return fmt.Errorf("unexpected image format %q", image)
	}
	prefix := fmt.Sprintf(prefixFormat, api.EKSResourceAccountID(region))
	if strings.HasSuffix(parts[0], suffix) {
		spec.Spec.Containers[0].Image = prefix + region + suffix + ":" + parts[1]
	}
	return nil
}
