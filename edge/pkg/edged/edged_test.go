package edged

import (
	"encoding/json"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelettypes "k8s.io/kubernetes/pkg/kubelet/types"
)

func TestHandlePodListFromMetaManager(t *testing.T) {
	deletionTime := metav1.Now()
	active := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "active"}, Spec: v1.PodSpec{NodeName: "edge-node"}}
	terminating := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "terminating", DeletionTimestamp: &deletionTime}, Spec: v1.PodSpec{NodeName: "edge-node"}}
	other := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "other"}, Spec: v1.PodSpec{NodeName: "other-node"}}
	tests := []struct {
		name    string
		pods    []*v1.Pod
		want    []*v1.Pod
		content []byte
		wantErr bool
	}{
		{name: "active pod", pods: []*v1.Pod{active}, want: []*v1.Pod{active}},
		{name: "complete snapshot", pods: []*v1.Pod{active, terminating, other}, want: []*v1.Pod{active, terminating}},
		{name: "other node", pods: []*v1.Pod{other}},
		{name: "empty snapshot"},
		{name: "invalid list", content: []byte(`{`), wantErr: true},
		{name: "invalid pod", content: []byte(`["{"]`), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := test.content
			if content == nil {
				lists := make([]string, 0, len(test.pods))
				for _, pod := range test.pods {
					data, err := json.Marshal(pod)
					if err != nil {
						t.Fatal(err)
					}
					lists = append(lists, string(data))
				}
				var err error
				content, err = json.Marshal(lists)
				if err != nil {
					t.Fatal(err)
				}
			}
			updates := make(chan interface{}, 1)
			edge := &edged{nodeName: "edge-node"}
			err := edge.handlePodListFromMetaManager(content, updates)
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected error: %v, wantErr: %v", err, test.wantErr)
			}
			if test.wantErr {
				if len(updates) != 0 {
					t.Fatal("invalid input emitted a pod update")
				}
				return
			}
			select {
			case rawUpdate := <-updates:
				update, ok := rawUpdate.(kubelettypes.PodUpdate)
				if !ok {
					t.Fatalf("unexpected update type: %T", rawUpdate)
				}
				want := kubelettypes.PodUpdate{Op: kubelettypes.SET, Pods: test.want, Source: kubelettypes.ApiserverSource}
				wantJSON, _ := json.Marshal(want)
				gotJSON, _ := json.Marshal(update)
				if !reflect.DeepEqual(gotJSON, wantJSON) {
					t.Fatalf("unexpected update: %s, want: %s", gotJSON, wantJSON)
				}
			default:
				t.Fatal("missing pod update")
			}
		})
	}
}
