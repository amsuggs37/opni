package gpuadapter_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/rancher/opni/apis/v1beta1"
	"github.com/rancher/opni/pkg/providers"
	testdata "github.com/rancher/opni/pkg/resources/gpuadapter/testdata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	imageSpecs = []v1beta1.ImagesSpec{
		{},
		{DriverManager: "foo:bar"},
		{DCGM: "foo:bar"},
		{DCGMExporter: "foo:bar"},
		{DevicePlugin: "foo:bar"},
		{GFD: "foo:bar"},
		{InitContainer: "foo:bar"},
		{Toolkit: "foo:bar"},
		{Validator: "foo:bar"},
		{MIGManager: "foo:bar"},
		{
			Driver:        "foo0:bar",
			DriverManager: "foo1:bar",
			DCGM:          "foo2:bar",
			DCGMExporter:  "foo3:bar",
			DevicePlugin:  "foo4:bar",
			GFD:           "foo5:bar",
			InitContainer: "foo6:bar",
			Toolkit:       "foo7:bar",
			Validator:     "foo8:bar",
			MIGManager:    "foo9:bar",
		},
	}
	manualProviders = []string{
		"auto",
		"k3s",
		"rke2",
		"rke",
		"none",
	}
	runtimes = []v1beta1.ContainerRuntime{
		v1beta1.Auto,
		v1beta1.Docker,
		v1beta1.Containerd,
		v1beta1.Crio,
	}
	discoveredProviders = []providers.Provider{
		providers.K3S,
		providers.RKE,
		providers.RKE2,
		providers.Unknown,
	}
	vgpuSpecs = []*v1beta1.VGPUSpec{
		nil,
		{
			LicenseConfigMap:  "foo",
			LicenseServerKind: "nls",
		},
		{
			LicenseConfigMap:  "bar",
			LicenseServerKind: "legacy",
		},
	}
)

var _ = Describe("GpuAdapter", Label("unit"), func() {
	It("should handle all permutations of GpuAdapter", func() {
		for _, vgpuSpec := range vgpuSpecs {
			for _, discoveredProvider := range discoveredProviders {
				for _, runtime := range runtimes {
					for _, manualProvider := range manualProviders {
						for _, imageSpec := range imageSpecs {
							testdata.Check(&v1beta1.GpuPolicyAdapter{
								TypeMeta: metav1.TypeMeta{
									APIVersion: v1beta1.GroupVersion.Identifier(),
									Kind:       "GpuPolicyAdapter",
								},
								ObjectMeta: metav1.ObjectMeta{
									Name:      "test",
									Namespace: "test",
								},
								Spec: v1beta1.GpuPolicyAdapterSpec{
									ContainerRuntime:   runtime,
									KubernetesProvider: manualProvider,
									Images:             imageSpec,
									VGPU:               vgpuSpec,
								},
							}, discoveredProvider)
						}
					}
				}
			}
		}
	})
})
