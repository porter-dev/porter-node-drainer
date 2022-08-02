package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/drain"
)

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value bool   `json:"value"`
}

func DrainNode(clientset *kubernetes.Clientset, hostname string) error {
	drainer := &drain.Helper{
		Client:              clientset,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		DeleteEmptyDirData:  true,
		Out:                 os.Stdout,
		ErrOut:              os.Stderr,
	}

	// query for the relevant node
	nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("kubernetes.io/hostname=%s", hostname),
	})

	if err != nil {
		fmt.Println("could not list nodes for hostname", hostname, err)
		return err
	}

	if len(nodeList.Items) != 1 {
		fmt.Printf("got invalid number of nodes matching hostname %s: %d\n", hostname, len(nodeList.Items))
		return fmt.Errorf("invalid number of nodes matching hostname")
	}

	matchedNode := nodeList.Items[0]

	// cordon the matched node
	fmt.Println("cordoning node")

	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/unschedulable",
		Value: true,
	}}

	payloadBytes, _ := json.Marshal(payload)

	_, err = clientset.CoreV1().Nodes().Patch(context.Background(), matchedNode.Name, types.JSONPatchType, payloadBytes, v1.PatchOptions{})

	if err != nil {
		fmt.Println("could not cordon node", hostname, err)
		return err
	}

	// drain the matched node
	fmt.Println("draining node")

	podList, errList := drainer.GetPodsForDeletion(matchedNode.Name)

	if len(errList) != 0 {
		fmt.Println("could not get pods for deletion")

		for _, err := range errList {
			fmt.Println(err)
		}

		return errList[0]
	}

	fmt.Println("planning to delete pods:")

	for _, pod := range podList.Pods() {
		fmt.Printf("%s/%s\n", pod.Namespace, pod.Name)
	}

	err = drainer.DeleteOrEvictPods(podList.Pods())

	if err != nil {
		fmt.Println("could not delete or evict pods", err)
		return err
	}

	return nil
}
