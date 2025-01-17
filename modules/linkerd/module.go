package linkerd

import (
	"context"

	"github.com/rancher/rio/modules/linkerd/feature"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/types"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	if err := installLinkerd(rContext); err != nil {
		return err
	}
	return feature.Register(ctx, rContext)
}

func installLinkerd(rContext *types.Context) error {
	cmClient := rContext.Core.Core().V1().ConfigMap()
	linkerdUpgrade := ""
	if _, err := cmClient.Get("linkerd", "linkerd-config", metav1.GetOptions{}); err == nil {
		linkerdUpgrade = "TRUE"
	}
	if constants.DevMode && linkerdUpgrade == "TRUE" {
		return nil
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    rContext.Namespace,
			GenerateName: "linkerd-install-",
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &[]int32{120}[0],
			BackoffLimit:            &[]int32{1}[0],
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					ServiceAccountName: "rio-controller-serviceaccount",
					RestartPolicy:      v1.RestartPolicyNever,
					Containers: []v1.Container{
						{
							Name:            "linkerd-install",
							Image:           constants.LinkerdInstallImage,
							ImagePullPolicy: v1.PullAlways,
							Env: []v1.EnvVar{
								{
									Name:  "LINKERD_UPGRADE",
									Value: linkerdUpgrade,
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := rContext.Batch.Batch().V1().Job().Create(job)
	return err
}
