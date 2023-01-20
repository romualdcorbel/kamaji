// Copyright 2022 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkProfileSpec defines the desired state of NetworkProfile.
type NetworkProfileSpec struct {
	// Address where API server of will be exposed.
	// In case of LoadBalancer Service, this can be empty in order to use the exposed IP provided by the cloud controller manager.
	Address string `json:"address,omitempty"`
	// AllowAddressAsExternalIP will include tenantControlPlane.Spec.NetworkProfile.Address in the section of
	// ExternalIPs of the Kubernetes Service (only ClusterIP or NodePort)
	AllowAddressAsExternalIP bool `json:"allowAddressAsExternalIP,omitempty"`
	// Port where API server of will be exposed
	// +kubebuilder:default=6443
	Port int32 `json:"port,omitempty"`
	// CertSANs sets extra Subject Alternative Names (SANs) for the API Server signing certificate.
	// Use this field to add additional hostnames when exposing the Tenant Control Plane with third solutions.
	CertSANs []string `json:"certSANs,omitempty"`
	// Kubernetes Service
	// +kubebuilder:default="10.96.0.0/16"
	ServiceCIDR string `json:"serviceCidr,omitempty"`
	// CIDR for Kubernetes Pods
	// +kubebuilder:default="10.244.0.0/16"
	PodCIDR string `json:"podCidr,omitempty"`
	// +kubebuilder:default={"10.96.0.10"}
	DNSServiceIPs []string `json:"dnsServiceIPs,omitempty"`
}

// +kubebuilder:validation:Enum=Hostname;InternalIP;ExternalIP;InternalDNS;ExternalDNS
type KubeletPreferredAddressType string

const (
	NodeHostName    KubeletPreferredAddressType = "Hostname"
	NodeInternalIP  KubeletPreferredAddressType = "InternalIP"
	NodeExternalIP  KubeletPreferredAddressType = "ExternalIP"
	NodeInternalDNS KubeletPreferredAddressType = "InternalDNS"
	NodeExternalDNS KubeletPreferredAddressType = "ExternalDNS"
)

type KubeletSpec struct {
	// Ordered list of the preferred NodeAddressTypes to use for kubelet connections.
	// Default to Hostname, InternalIP, ExternalIP.
	// +kubebuilder:default={"Hostname","InternalIP","ExternalIP"}
	// +kubebuilder:validation:MinItems=1
	PreferredAddressTypes []KubeletPreferredAddressType `json:"preferredAddressTypes,omitempty"`
	// CGroupFS defines the  cgroup driver for Kubelet
	// https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/configure-cgroup-driver/
	CGroupFS CGroupDriver `json:"cgroupfs,omitempty"`
}

// KubernetesSpec defines the desired state of Kubernetes.
type KubernetesSpec struct {
	// Kubernetes Version for the tenant control plane
	Version string      `json:"version"`
	Kubelet KubeletSpec `json:"kubelet"`

	// List of enabled Admission Controllers for the Tenant cluster.
	// Full reference available here: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers
	// +kubebuilder:default=CertificateApproval;CertificateSigning;CertificateSubjectRestriction;DefaultIngressClass;DefaultStorageClass;DefaultTolerationSeconds;LimitRanger;MutatingAdmissionWebhook;NamespaceLifecycle;PersistentVolumeClaimResize;Priority;ResourceQuota;RuntimeClass;ServiceAccount;StorageObjectInUseProtection;TaintNodesByCondition;ValidatingAdmissionWebhook
	AdmissionControllers AdmissionControllers `json:"admissionControllers,omitempty"`
}

// AdditionalMetadata defines which additional metadata, such as labels and annotations, must be attached to the created resource.
type AdditionalMetadata struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ControlPlane defines how the Tenant Control Plane Kubernetes resources must be created in the Admin Cluster,
// such as the number of Pod replicas, the Service resource, or the Ingress.
type ControlPlane struct {
	// Defining the options for the deployed Tenant Control Plane as Deployment resource.
	Deployment DeploymentSpec `json:"deployment,omitempty"`
	// Defining the options for the Tenant Control Plane Service resource.
	Service ServiceSpec `json:"service"`
	// Defining the options for an Optional Ingress which will expose API Server of the Tenant Control Plane
	Ingress *IngressSpec `json:"ingress,omitempty"`
}

// IngressSpec defines the options for the ingress which will expose API Server of the Tenant Control Plane.
type IngressSpec struct {
	AdditionalMetadata AdditionalMetadata `json:"additionalMetadata,omitempty"`
	IngressClassName   string             `json:"ingressClassName,omitempty"`
	// Hostname is an optional field which will be used as Ingress's Host. If it is not defined,
	// Ingress's host will be "<tenant>.<namespace>.<domain>", where domain is specified under NetworkProfileSpec
	Hostname string `json:"hostname,omitempty"`
}

// ComponentResourceRequirements describes the compute resource requirements.
type ComponentResourceRequirements struct {
	// Limits describes the maximum amount of compute resources allowed.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	Limits corev1.ResourceList `json:"limits,omitempty" protobuf:"bytes,1,rep,name=limits,casttype=ResourceList,castkey=ResourceName"`
	// Requests describes the minimum amount of compute resources required.
	// If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
	// otherwise to an implementation-defined value.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	Requests corev1.ResourceList `json:"requests,omitempty" protobuf:"bytes,2,rep,name=requests,casttype=ResourceList,castkey=ResourceName"`
}

type ControlPlaneComponentsResources struct {
	APIServer         *ComponentResourceRequirements `json:"apiServer,omitempty"`
	ControllerManager *ComponentResourceRequirements `json:"controllerManager,omitempty"`
	Scheduler         *ComponentResourceRequirements `json:"scheduler,omitempty"`
}

type DeploymentSpec struct {
	// +kubebuilder:default=2
	Replicas int32 `json:"replicas,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// RuntimeClassName refers to a RuntimeClass object in the node.k8s.io group, which should be used
	// to run the Tenant Control Plane pod. If no RuntimeClass resource matches the named class, the pod will not be run.
	// If unset or empty, the "legacy" RuntimeClass will be used, which is an implicit class with an
	// empty definition that uses the default runtime handler.
	// More info: https://git.k8s.io/enhancements/keps/sig-node/585-runtime-class
	RuntimeClassName string `json:"runtimeClassName,omitempty"`
	// Strategy describes how to replace existing pods with new ones for the given Tenant Control Plane.
	// Default value is set to Rolling Update, with a blue/green strategy.
	// +kubebuilder:default={type:"RollingUpdate",rollingUpdate:{maxUnavailable:0,maxSurge:"100%"}}
	Strategy appsv1.DeploymentStrategy `json:"strategy,omitempty"`
	// If specified, the Tenant Control Plane pod's tolerations.
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// If specified, the Tenant Control Plane pod's scheduling constraints.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes-using-node-affinity/
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// TopologySpreadConstraints describes how the Tenant Control Plane pods ought to spread across topology
	// domains. Scheduler will schedule pods in a way which abides by the constraints.
	// In case of nil underlying LabelSelector, the Kamaji one for the given Tenant Control Plane will be used.
	// All topologySpreadConstraints are ANDed.
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// Resources defines the amount of memory and CPU to allocate to each component of the Control Plane
	// (kube-apiserver, controller-manager, and scheduler).
	Resources *ControlPlaneComponentsResources `json:"resources,omitempty"`
	// ExtraArgs allows adding additional arguments to the Control Plane components,
	// such as kube-apiserver, controller-manager, and scheduler.
	ExtraArgs          *ControlPlaneExtraArgs `json:"extraArgs,omitempty"`
	AdditionalMetadata AdditionalMetadata     `json:"additionalMetadata,omitempty"`
}

// ControlPlaneExtraArgs allows specifying additional arguments to the Control Plane components.
type ControlPlaneExtraArgs struct {
	APIServer         []string `json:"apiServer,omitempty"`
	ControllerManager []string `json:"controllerManager,omitempty"`
	Scheduler         []string `json:"scheduler,omitempty"`
	// Available only if Kamaji is running using Kine as backing storage.
	Kine []string `json:"kine,omitempty"`
}

type ServiceSpec struct {
	AdditionalMetadata AdditionalMetadata `json:"additionalMetadata,omitempty"`
	// ServiceType allows specifying how to expose the Tenant Control Plane.
	ServiceType ServiceType `json:"serviceType"`
}

// AddonSpec defines the spec for every addon.
type AddonSpec struct {
	ImageOverrideTrait `json:",inline"`
}

type ImageOverrideTrait struct {
	// ImageRepository sets the container registry to pull images from.
	// if not set, the default ImageRepository will be used instead.
	ImageRepository string `json:"imageRepository,omitempty"`
	// ImageTag allows to specify a tag for the image.
	// In case this value is set, kubeadm does not change automatically the version of the above components during upgrades.
	ImageTag string `json:"imageTag,omitempty"`
}

// ExtraArgs allows adding additional arguments to said component.
type ExtraArgs []string

type KonnectivityServerSpec struct {
	// The port which Konnectivity server is listening to.
	Port int32 `json:"port"`
	// Container image version of the Konnectivity server.
	// +kubebuilder:default=v0.0.32
	Version string `json:"version,omitempty"`
	// Container image used by the Konnectivity server.
	// +kubebuilder:default=registry.k8s.io/kas-network-proxy/proxy-server
	Image string `json:"image,omitempty"`
	// Resources define the amount of CPU and memory to allocate to the Konnectivity server.
	Resources *ComponentResourceRequirements `json:"resources,omitempty"`
	ExtraArgs ExtraArgs                      `json:"extraArgs,omitempty"`
}

type KonnectivityAgentSpec struct {
	// AgentImage defines the container image for Konnectivity's agent.
	// +kubebuilder:default=registry.k8s.io/kas-network-proxy/proxy-agent
	Image string `json:"image,omitempty"`
	// Version for Konnectivity agent.
	// +kubebuilder:default=v0.0.32
	Version   string    `json:"version,omitempty"`
	ExtraArgs ExtraArgs `json:"extraArgs,omitempty"`
}

// KonnectivitySpec defines the spec for Konnectivity.
type KonnectivitySpec struct {
	// +kubebuilder:default={version:"v0.0.32",image:"registry.k8s.io/kas-network-proxy/proxy-server",port:8132}
	KonnectivityServerSpec KonnectivityServerSpec `json:"server,omitempty"`
	// +kubebuilder:default={version:"v0.0.32",image:"registry.k8s.io/kas-network-proxy/proxy-agent"}
	KonnectivityAgentSpec KonnectivityAgentSpec `json:"agent,omitempty"`
}

// AddonsSpec defines the enabled addons and their features.
type AddonsSpec struct {
	// Enables the DNS addon in the Tenant Cluster.
	// The registry and the tag are configurable, the image is hard-coded to `coredns`.
	CoreDNS *AddonSpec `json:"coreDNS,omitempty"`
	// Enables the Konnectivity addon in the Tenant Cluster, required if the worker nodes are in a different network.
	Konnectivity *KonnectivitySpec `json:"konnectivity,omitempty"`
	// Enables the kube-proxy addon in the Tenant Cluster.
	// The registry and the tag are configurable, the image is hard-coded to `kube-proxy`.
	KubeProxy *AddonSpec `json:"kubeProxy,omitempty"`
}

// TenantControlPlaneSpec defines the desired state of TenantControlPlane.
type TenantControlPlaneSpec struct {
	// DataStore allows to specify a DataStore that should be used to store the Kubernetes data for the given Tenant Control Plane.
	// This parameter is optional and acts as an override over the default one which is used by the Kamaji Operator.
	// Migration from a different DataStore to another one is not yet supported and the reconciliation will be blocked.
	DataStore    string       `json:"dataStore,omitempty"`
	ControlPlane ControlPlane `json:"controlPlane"`
	// Kubernetes specification for tenant control plane
	Kubernetes KubernetesSpec `json:"kubernetes"`
	// NetworkProfile specifies how the network is
	NetworkProfile NetworkProfileSpec `json:"networkProfile,omitempty"`
	// Addons contain which addons are enabled
	Addons AddonsSpec `json:"addons,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.controlPlane.deployment.replicas,statuspath=.status.kubernetesResources.deployment.replicas,selectorpath=.status.kubernetesResources.deployment.selector
// +kubebuilder:resource:shortName=tcp
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.kubernetes.version",description="Kubernetes version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.kubernetesResources.version.status",description="Status"
// +kubebuilder:printcolumn:name="Control-Plane endpoint",type="string",JSONPath=".status.controlPlaneEndpoint",description="Tenant Control Plane Endpoint (API server)"
// +kubebuilder:printcolumn:name="Kubeconfig",type="string",JSONPath=".status.kubeconfig.admin.secretName",description="Secret which contains admin kubeconfig"
//+kubebuilder:printcolumn:name="Datastore",type="string",JSONPath=".status.storage.dataStoreName",description="DataStore actually used"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age"

// TenantControlPlane is the Schema for the tenantcontrolplanes API.
type TenantControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantControlPlaneSpec   `json:"spec,omitempty"`
	Status TenantControlPlaneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TenantControlPlaneList contains a list of TenantControlPlane.
type TenantControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TenantControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TenantControlPlane{}, &TenantControlPlaneList{})
}
