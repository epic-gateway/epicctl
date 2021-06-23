package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

func init() {
	var (
		account      string
		serviceGroup string
	)

	listLBCmd := &cobra.Command{
		Use:     "load-balancer",
		Aliases: []string{"lb", "lbs"},
		Short:   "List load balancers",
		Long:    `List EPIC load balancers in the provided account and service group.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getEpicClient()
			if err != nil {
				return err
			}

			return listLB(rootCmd.Context(), cl, serviceGroup, account)
		},
	}
	listLBCmd.Flags().StringVarP(&account, "account", "a", "", "(required) account in which the LB lives")
	listLBCmd.Flags().StringVarP(&serviceGroup, "service-group", "g", "", "(required) service group to which the LB belongs")
	listLBCmd.MarkFlagRequired("account")
	listLBCmd.MarkFlagRequired("service-group")
	listCmd.AddCommand(listLBCmd)
}

// listLB lists the load balancers in the provided account and service
// group.
func listLB(ctx context.Context, cl client.Client, serviceGroupName string, accountName string) error {
	// List all of the LBs
	listOps := client.ListOptions{Namespace: epicv1.AccountNamespace(accountName)}
	if serviceGroupName != "" {
		labelSelector := labels.SelectorFromSet(map[string]string{epicv1.OwningLBServiceGroupLabel: serviceGroupName})
		listOps.LabelSelector = labelSelector
	}
	list := epicv1.LoadBalancerList{}
	if err := cl.List(ctx, &list, &listOps); err != nil {
		return err
	}

	// For each LB, print summary
	for _, lb := range list.Items {
		fmt.Printf("Name: %s\n", lb.Spec.DisplayName)
	}

	return nil
}
