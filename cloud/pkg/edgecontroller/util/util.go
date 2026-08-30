package util

import (
	"strings"

	v1 "k8s.io/api/core/v1"

	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
)

func IsEdgePod(config *v1alpha1.EdgeController, pod *v1.Pod) bool {
	var namespaces []string
	var podNamePrefixs []string

	if len(config.FilterPodNamespaces) != 0 {
		namespaces = strings.Split(config.FilterPodNamespaces, ",")
	}

	if len(config.FilterPodNamePrefixs) != 0 {
		podNamePrefixs = strings.Split(config.FilterPodNamePrefixs, ",")
	}

	if InArray(namespaces, pod.Namespace) {
		return false
	}

	if MatchPrefixs(podNamePrefixs, pod.Name) {
		return false
	}

	return true
}

func InArray(arr []string, target string) bool {
	if len(arr) == 0 {
		return false
	}
	for _, elem := range arr {
		if elem == target {
			return true
		}
	}
	return false
}

func MatchPrefixs(arr []string, target string) bool {
	if len(arr) == 0 {
		return false
	}
	for _, elem := range arr {
		if strings.HasPrefix(target, elem) {
			return true
		}
	}
	return false
}
