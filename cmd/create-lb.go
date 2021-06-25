package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

func init() {
	createLBCmd := &cobra.Command{
		Use:     "load-balancer lb-name lb-ports",
		Aliases: []string{"lb"},
		Short:   "Create load balancer",
		Long: `Create a new ad-hoc EPIC load balancer.

Two arguments are required: the load balancer name,
and a list of ports on which the load balancer will
listen. Each port is a protocol and number, e.g.,
TCP/8080 or UDP/123. If the load balancer listens
on more than one port they must be separated with
commas, e.g., TCP/8080,TCP/8888`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getEpicClient()
			if err != nil {
				return err
			}

			account, serviceGroup, err := getAccountAndSG()
			if err != nil {
				return err
			}

			epPorts, err := parsePorts(args[1])
			if err != nil {
				return err
			}
			return createLB(rootCmd.Context(), cl, args[0], serviceGroup, account, epPorts)
		},
	}
	bindAccountAndSG(createLBCmd)
	createCmd.AddCommand(createLBCmd)
}

// createLB creates a LoadBalancer.
func createLB(ctx context.Context, cl client.Client, lbName string, serviceGroupName string, accountName string, ports []corev1.ServicePort) error {

	Debug("creating LB %s in service group %s\n", lbName, serviceGroupName)
	for _, port := range ports {
		Debug(" port: %s/%d\n", port.Protocol, port.Port)
	}

	lb := epicv1.LoadBalancer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: epicv1.AccountNamespace(accountName),
			Name:      epicv1.LoadBalancerName(serviceGroupName, lbName, true),
			Labels: map[string]string{
				epicv1.OwningLBServiceGroupLabel: serviceGroupName,
			},
		},
		Spec: epicv1.LoadBalancerSpec{
			DisplayName:      lbName,
			PublicPorts:      ports,
			UpstreamClusters: []string{clusterName},
			TrueIngress:      false,
		},
	}

	// get the owning lbsg which points to the service prefix from which
	// we'll allocate the address
	mgroup := epicv1.LBServiceGroup{}
	if err := cl.Get(ctx, client.ObjectKey{Namespace: epicv1.AccountNamespace(accountName), Name: serviceGroupName}, &mgroup); err != nil {
		return err
	}
	lb.Labels[epicv1.OwningServicePrefixLabel] = mgroup.Labels[epicv1.OwningServicePrefixLabel]

	if err := cl.Create(ctx, &lb); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "LB %s created\n", lbName)

	return nil
}
