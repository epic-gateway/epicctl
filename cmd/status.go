package cmd

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "EPIC operational status",
		Long:  `Queries the EPIC cluster to determine its operational status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getEpicClient()
			if err != nil {
				panic(err.Error())
			}

			err = status(rootCmd.Context(), client)
			if err != nil {
				return err
			}

			return nil
		},
	})

}

// status determines the overall system status by calling more
// specific status functions, e.g., systemPodStatus.
func status(ctx context.Context, cl client.Client) error {
	var err error

	err = epicPodStatus(ctx, cl)
	if err != nil {
		return err
	}

	err = userPodStatus(ctx, cl)
	if err != nil {
		return err
	}

	fmt.Println("No problems found")

	return nil
}

// epicPodStatus checks the pods that run in the epic namespace.
func epicPodStatus(ctx context.Context, cl client.Client) error {
	nodes := v1.NodeList{}
	err := cl.List(ctx, &nodes)

	pods := v1.PodList{}
	err = cl.List(ctx, &pods, &client.ListOptions{Namespace: "epic"})
	if err != nil {
		return err
	}

	// Do we have the correct number of EPIC pods?
	if err = epicPodCount(&nodes, &pods); err != nil {
		return err
	}

	// Are all of the pods running?
	err = podStatus(ctx, cl, &nodes, &pods)
	if err != nil {
		return err
	}
	fmt.Println("All EPIC system pods are operational")

	return nil
}

// epicPodCount validates that the correct number of pods are running
// in the epic namespace.
func epicPodCount(nodes *v1.NodeList, pods *v1.PodList) error {
	// The count should be the sum of the agent daemonset + the web
	// service + the controller-manager
	desiredCount := len(nodes.Items) + 2

	if len(pods.Items) != desiredCount {
		return fmt.Errorf("incorrect system pod count: should be %d but is %d", desiredCount, len(pods.Items))
	}

	return nil
}

func userPodStatus(ctx context.Context, cl client.Client) error {

	return nil
}

// podStatus checks the pods that run in the epic namespace.
func podStatus(ctx context.Context, cl client.Client, nodes *v1.NodeList, pods *v1.PodList) error {
	// Are all of the pods running?
	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning {
			return fmt.Errorf("pod %s is not healthy: %+v", pod.Name, pod.Status.Phase)
		}
	}

	return nil
}
