/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provisioning

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	metal3iov1alpha1 "github.com/openshift/cluster-baremetal-operator/api/v1alpha1"
)

func TestBuildEnvVar(t *testing.T) {
	tCases := []struct {
		name           string
		configName     string
		spec           *metal3iov1alpha1.ProvisioningSpec
		expectedEnvVar corev1.EnvVar
	}{
		{
			name:       "Managed ProvisioningIPCIDR",
			configName: provisioningIP,
			spec:       managedProvisioning().build(),
			expectedEnvVar: corev1.EnvVar{
				Name:  provisioningIP,
				Value: "172.30.20.3/24",
			},
		},
		{
			name:       "Unmanaged ProvisioningInterface",
			configName: provisioningInterface,
			spec:       unmanagedProvisioning().build(),
			expectedEnvVar: corev1.EnvVar{
				Name:  provisioningInterface,
				Value: "ensp0",
			},
		},
		{
			name:       "Disabled MachineOsUrl",
			configName: machineImageUrl,
			spec:       disabledProvisioning().build(),
			expectedEnvVar: corev1.EnvVar{
				Name:  machineImageUrl,
				Value: "http://172.22.0.1/images/rhcos-44.81.202001171431.0-openstack.x86_64.qcow2.gz?sha256=e98f83a2b9d4043719664a2be75fe8134dc6ca1fdbde807996622f8cc7ecd234",
			},
		},
		{
			name:       "Disabled ProvisioningInterface",
			configName: provisioningInterface,
			spec:       disabledProvisioning().build(),
			expectedEnvVar: corev1.EnvVar{
				Name:  provisioningInterface,
				Value: "",
			},
		},
	}
	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing tc : %s", tc.name)
			actualEnvVar := buildEnvVar(tc.configName, tc.spec)
			assert.Equal(t, tc.expectedEnvVar, actualEnvVar, fmt.Sprintf("%s : Expected : %s Actual : %s", tc.configName, tc.expectedEnvVar, actualEnvVar))
			return
		})
	}

}

func TestNewMetal3InitContainers(t *testing.T) {
	images := Images{
		BaremetalOperator:   expectedBaremetalOperator,
		Ironic:              expectedIronic,
		IronicInspector:     expectedIronicInspector,
		IpaDownloader:       expectedIronicIpaDownloader,
		MachineOsDownloader: expectedMachineOsDownloader,
		StaticIpManager:     expectedIronicStaticIpManager,
	}
	tCases := []struct {
		name               string
		config             *metal3iov1alpha1.ProvisioningSpec
		expectedContainers []corev1.Container
	}{
		{
			name:   "valid config",
			config: managedProvisioning().build(),
			expectedContainers: []corev1.Container{
				{
					Name:  "metal3-ipa-downloader",
					Image: images.IpaDownloader,
				},
				{
					Name:  "metal3-machine-os-downloader",
					Image: images.MachineOsDownloader,
				},
				{
					Name:  "metal3-static-ip-set",
					Image: images.StaticIpManager,
				},
			},
		},
		{
			name:   "disabled without provisioning ip",
			config: disabledProvisioning().ProvisioningIP("").ProvisioningNetworkCIDR("").build(),
			expectedContainers: []corev1.Container{
				{
					Name:  "metal3-ipa-downloader",
					Image: images.IpaDownloader,
				},
				{
					Name:  "metal3-machine-os-downloader",
					Image: images.MachineOsDownloader,
				},
			},
		},
	}
	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing tc : %s", tc.name)
			actualContainers := newMetal3InitContainers(&images, tc.config)
			assert.Equal(t, len(tc.expectedContainers), len(actualContainers), fmt.Sprintf("%s : Expected number of Init Containers : %d Actual number of Init Containers : %d", tc.name, len(tc.expectedContainers), len(actualContainers)))
		})
	}

}

func TestNewMetal3Containers(t *testing.T) {
	images := Images{
		BaremetalOperator:   expectedBaremetalOperator,
		Ironic:              expectedIronic,
		IronicInspector:     expectedIronicInspector,
		IpaDownloader:       expectedIronicIpaDownloader,
		MachineOsDownloader: expectedMachineOsDownloader,
		StaticIpManager:     expectedIronicStaticIpManager,
	}
	tCases := []struct {
		name               string
		config             *metal3iov1alpha1.ProvisioningSpec
		expectedContainers int
	}{
		{
			name:               "ManagedSpec",
			config:             managedProvisioning().build(),
			expectedContainers: 10,
		},
		{
			name:               "UnmanagedSpec",
			config:             unmanagedProvisioning().build(),
			expectedContainers: 10,
		},
		{
			name:               "DisabledSpec",
			config:             disabledProvisioning().build(),
			expectedContainers: 9,
		},
		{
			name:               "DisabledSpecWithoutProvisioningIP",
			config:             disabledProvisioning().ProvisioningIP("").ProvisioningNetworkCIDR("").build(),
			expectedContainers: 8,
		},
	}
	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing tc : %s", tc.name)
			actualContainers := newMetal3Containers(&images, tc.config)
			assert.Equal(t, tc.expectedContainers, len(actualContainers), fmt.Sprintf("%s : Expected number of Containers : %d Actual number of Containers : %d", tc.name, tc.expectedContainers, len(actualContainers)))
		})
	}
}
