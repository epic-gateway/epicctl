package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

func init() {
	var (
		account      string
		serviceGroup string
	)

	lbDeleteCmd := &cobra.Command{
		Use:     "load-balancer lb-name",
		Aliases: []string{"lb"},
		Short:   "Delete load balancer",
		Long:    `Delete an EPIC load balancer.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getEpicClient()
			if err != nil {
				return err
			}

			return deleteLB(context.Background(), cl, args[0], serviceGroup, account)
		},
	}
	lbDeleteCmd.Flags().StringVarP(&account, "account", "a", "", "(required) account in which the LB lives")
	lbDeleteCmd.Flags().StringVarP(&serviceGroup, "service-group", "g", "", "(required) service group to which the LB belongs")
	lbDeleteCmd.MarkFlagRequired("account")
	lbDeleteCmd.MarkFlagRequired("service-group")
	deleteCmd.AddCommand(lbDeleteCmd)
}

// deleteLB deletes a LoadBalancer.
func deleteLB(ctx context.Context, cl client.Client, lbName string, serviceGroupName string, accountName string) error {
	var (
		err error
		lb  epicv1.LoadBalancer
	)

	// Need to get the LB first so we can remove our cluster from it
	if lb, err = getLB(ctx, cl, lbName, serviceGroupName, accountName); err != nil {
		return err
	}

	// Remove our upstream cluster from the LB
	if err = lb.RemoveUpstream(clusterName); err != nil {
		return err
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
