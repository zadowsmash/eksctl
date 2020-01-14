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

// imageTag extracts the container image's tag.
func imageTag(image string) (string, error) {
	parts := strings.Split(image, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected image format %q", image)
	}

	return parts[1], nil
}

// imageTagsDiffer returns true if the image tags are not the same
// while ignoring the image name.
func imageTagsDiffer(image1, image2 string) (bool, error) {
	tag1, err := imageTag(image1)
	if err != nil {
		return false, err
	}
	tag2, err := imageTag(image2)
	if err != nil {
		return false, err
	}
	return tag1 != tag2, nil
}
