package defaultaddons

import (
	"github.com/kris-nova/logger"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/weaveworks/eksctl/pkg/kubernetes"
)

const (
	// AWSNode is the name of the aws-node addon
	AWSNode = "aws-node"

	awsNodeImagePrefixPTN = "%s.dkr.ecr."
	awsNodeImageSuffix    = ".amazonaws.com/amazon-k8s-cni"
)

// UpdateAWSNode will update the `aws-node` add-on and returns true
// if an update is available.
func UpdateAWSNode(rawClient kubernetes.RawClientInterface, region string, plan bool) (bool, error) {
	daemonSet, err := rawClient.ClientSet().AppsV1().DaemonSets(metav1.NamespaceSystem).Get(AWSNode, metav1.GetOptions{})
	if err != nil {
		if apierrs.IsNotFound(err) {
			logger.Warning("%q was not found", AWSNode)
			return false, nil
		}
		return false, errors.Wrapf(err, "getting %q", AWSNode)
	}

	// if DaemonSets is present, go through our list of assets
	list, err := LoadAsset(AWSNode, "yaml")
	if err != nil {
		return false, err
	}

	tagMismatch := true
	for _, rawObj := range list.Items {
		resource, err := rawClient.NewRawResource(rawObj.Object)
		if err != nil {
			return false, err
		}
		if resource.GVK.Kind == "DaemonSet" {
			podTemplate := resource.Info.Object.(*appsv1.DaemonSet).Spec.Template
			if err := useRegionalImage(&podTemplate, awsNodeImagePrefixPTN, region, awsNodeImageSuffix); err != nil {
				return false, errors.Wrapf(err, "update image for %q", AWSNode)
			}
			tagMismatch, err = imageTagsDiffer(
				podTemplate.Spec.Containers[0].Image,
				daemonSet.Spec.Template.Spec.Containers[0].Image,
			)
			if err != nil {
				return false, err
			}
		}

		if resource.GVK.Kind == "CustomResourceDefinition" && plan {
			// eniconfigs.crd.k8s.amazonaws.com CRD is only partially defined in the
			// manifest, and causes a range of issue in plan mode, we can skip it
			logger.Info(resource.LogAction(plan, "replaced"))
			continue
		}

		status, err := resource.CreateOrReplace(plan)
		if err != nil {
			return false, err
		}
		logger.Info(status)
	}

	if plan {
		if tagMismatch {
			logger.Critical("(plan) %q is not up-to-date", AWSNode)
			return true, nil
		}
		logger.Info("(plan) %q is already up-to-date", AWSNode)
		return false, nil
	}

	logger.Info("%q is now up-to-date", AWSNode)
	return false, nil
}
