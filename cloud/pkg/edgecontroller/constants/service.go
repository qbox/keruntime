package constants

// Service level constants
const (
	ResourceNode = "node"

	// GroupResource Group
	GroupResource = "resource"

	// NvidiaGPUStatusAnnotationKey Nvidia Constants
	// NvidiaGPUStatusAnnotationKey is the key of the node annotation for GPU status
	NvidiaGPUStatusAnnotationKey = "huawei.com/gpu-status"
	// NvidiaGPUScalarResourceName is the device plugin resource name used for special handling
	NvidiaGPUScalarResourceName = "nvidia.com/gpu"

	// configmap label key, the value can be pod or native
	ConfigType = "configType"
	// configmap label key, the value is node name
	NodeName = "nodeName"
	// configmap label key, the value is app name
	AppName = "appName"
	// secret label key, the value is domain
	Domain = "domain"
	// configType: pod
	Pod = "pod"
	// configType: native
	Native = "native"
	// edge node role label
	EdgeNodeLabel = "node-role.kubernetes.io/edge"
	// node ready status
	Ready = "Ready"

	// default filter configmap name
	FilterConfig = "filter-config"
	// default filter namespace item name
	FilterPodNamespaces = "filterPodNamespaces"
	// default filter pod name prefix
	FilterPodNamePrefixs = "filterPodNamePrefixs"
)
