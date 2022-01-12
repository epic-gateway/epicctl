package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

func init() {
	describeLBCmd := &cobra.Command{
		Use:     "load-balancer lb-name user-namespace service-group",
		Aliases: []string{"lb"},
		Short:   "Describes load balancers",
		Long:    `Describes EPIC load balancers.`,
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getEpicClient()
			if err != nil {
				return err
			}

			account := args[1]
			serviceGroup := args[2]

			return describeLB(context.Background(), cl, args[0], serviceGroup, account)
		},
	}
	describeCmd.AddCommand(describeLBCmd)
}

// describeLB gets a LoadBalancer from the cluster and dumps its contents
// to stdout.
func describeLB(ctx context.Context, cl client.Client, lbName string, serviceGroupName string, accountName string) error {
	var (
		err error
		lb  epicv1.LoadBalancer
	)

	if lb, err = getLB(ctx, cl, lbName, serviceGroupName, accountName); err != nil {
		return err
	}

	// LB Info
	fmt.Printf("Load Balancer Name: %s\n", lb.Spec.DisplayName)
	fmt.Printf(" Public Address: %s\n", lb.Spec.PublicAddress)
	fmt.Printf(" TrueIngress: %t\n", lb.Spec.TrueIngress)
	fmt.Printf(" Public Ports: %s\n", lb.Spec.PublicAddress)
	for _, port := range lb.Spec.PublicPorts {
		fmt.Printf("  %d %s\n", port.Port, port.Protocol)
	}
	fmt.Printf(" Upstream Clusters:\n")
	for _, cluster := range lb.Spec.UpstreamClusters {
		fmt.Printf("  %s\n", cluster)
	}

	// Endpoints
	reps, err := getReps(ctx, cl, lbName, serviceGroupName, accountName)
	if err != nil {
		return err
	}
	fmt.Printf(" Upstream Endpoints:\n")
	for _, rep := range reps.Items {
		fmt.Printf("  %s %d %s %s\n", rep.Spec.Address, rep.Spec.Port.Port, rep.Spec.Port.Protocol, rep.Spec.Cluster)
	}
	Debug(" Raw CR Contents: %+v\n", lb)

	return nil
}
