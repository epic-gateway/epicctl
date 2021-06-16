package cmd

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spf13/cobra"
)

const (
	// A typical EPIC will have 3 pods running in the epic namespace:
	// api service, controller manager, and node agent.
	systemPodCount = 3
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

			err = status(context.Background(), client)
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

	err = systemPodStatus(ctx, cl)
	if err != nil {
		return err
	}

	err = userPodStatus(ctx, cl)
	if err != nil {
		return err
	}

	return nil
}

// systemPodStatus checks the pods that run in the epic namespace.
func systemPodStatus(ctx context.Context, cl client.Client) error {
	pods := v1.PodList{}
	err := cl.List(ctx, &pods, &client.ListOptions{Namespace: "epic"})
	if err != nil {
		return err
	}

	// Do we have the correct number of pods?
	if len(pods.Items) != systemPodCount {
		return fmt.Errorf("incorrect system pod count. Should be %d but is %d", systemPodCount, len(pods.Items))
	}

	// Are all of the pods running?
	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning {
			return fmt.Errorf("pod %s is not healthy: %+v", pod.Name, pod.Status.Phase)
		}
	}
	fmt.Println("All EPIC system pods are operational.")

	return nil
}

func userPodStatus(ctx context.Context, cl client.Client) error {

	return nil
}
