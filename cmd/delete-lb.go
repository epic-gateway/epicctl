package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

func init() {
	deleteLBCmd := &cobra.Command{
		Use:     "load-balancer lb-name",
		Aliases: []string{"lb"},
		Short:   "Delete load balancer",
		Long:    `Delete an EPIC load balancer.`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cl, err := getEpicClient()
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}

			account, serviceGroup, err := getAccountAndSG()
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}

			if err := deleteLB(context.Background(), cl, args[0], serviceGroup, account, viper.GetBool("force")); err != nil {
				fmt.Printf("%s\n", err.Error())
			}
		},
	}
	bindAccountAndSG(deleteLBCmd)
	deleteLBCmd.Flags().Bool("force", false, "DANGER: Unconditionally delete the LB")
	viper.BindPFlag("force", deleteLBCmd.Flags().Lookup("force"))
	deleteCmd.AddCommand(deleteLBCmd)
}

// deleteLB deletes a LoadBalancer.
func deleteLB(ctx context.Context, cl client.Client, lbName string, serviceGroupName string, accountName string, force bool) error {
	var (
		err error
		lb  epicv1.LoadBalancer
	)

	// Need to get the LB first so we can remove our cluster from it
	if lb, err = getLB(ctx, cl, lbName, serviceGroupName, accountName); err != nil {
		return err
	}

	if force {
		// If we've been told to force-delete the LB then blow away all of
		// the upstream clusters
		lb.Spec.UpstreamClusters = []string{}
	} else {
		// attempt to delete just the cluster that this tool uses
		if err = lb.RemoveUpstream(clusterName); err != nil {
			return err
		}
	}
	if err = cl.Update(ctx, &lb); err != nil {
		return err
	}

	// Now we can try to delete the LB
	if err := cl.Delete(ctx, &lb); err != nil {
		return err
	}

	return nil
}
