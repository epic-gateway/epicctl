package cmd

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

const (
	clusterName = "epicctl"
)

// parsePorts parses an array of strings and returns an array of
// corev1.ServicePort. Each string can either be "proto/port" or just
// "port". The proto must be either "TCP" or "UDP" and the portNum
// must fit in a 32-bit signed integer. If the proto is missing it
// defaults to "TCP". If error is non-nil than something has gone
// wrong and the ServicePort array is undefined.
func parsePorts(ports []string) ([]corev1.ServicePort, error) {
	epPorts := make([]corev1.ServicePort, len(ports))

	for i, port := range ports {
		parts := strings.Split(port, "/")
		if len(parts) == 2 {
			// if there's a "/" then it's a protocol and port num, e.g., "TCP/8080"
			portNum, err := strconv.ParseInt(parts[1], 10, 32)
			if err != nil {
				return epPorts, err
			}
			epPorts[i] = corev1.ServicePort{
				Protocol: corev1.Protocol(parts[0]),
				Port:     int32(portNum),
			}
		} else {
			// if there's no "/" then it's just a port num, e.g., "8080"
			portNum, err := strconv.ParseInt(parts[0], 10, 32)
			if err != nil {
				return epPorts, err
			}
			epPorts[i] = corev1.ServicePort{
				Protocol: corev1.ProtocolTCP,
				Port:     int32(portNum),
			}
		}
	}
	return epPorts, nil
}

func getLB(ctx context.Context, cl client.Client, lbName string, serviceGroupName string, accountName string) (epicv1.LoadBalancer, error) {
	var err error = nil
	lb := epicv1.LoadBalancer{}
	lb.Name = epicv1.LoadBalancerName(serviceGroupName, lbName, true)
	err = cl.Get(ctx, client.ObjectKey{Namespace: epicv1.AccountNamespace(accountName), Name: lb.Name}, &lb)

	return lb, err
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

// showLB gets a LoadBalancer from the cluster and dumps its contents
// to stdout.
func showLB(ctx context.Context, cl client.Client, lbName string, serviceGroupName string, accountName string) error {
	var (
		err error
		lb  epicv1.LoadBalancer
	)

	if lb, err = getLB(ctx, cl, lbName, serviceGroupName, accountName); err != nil {
		return err
	}

	fmt.Printf("Load Balancer - %s\n", lb.Spec.DisplayName)
	fmt.Printf("%+v\n", lb)

	return nil
}

// createLB creates a LoadBalancer.
func createLB(ctx context.Context, cl client.Client, lbName string, serviceGroupName string, accountName string, ports []corev1.ServicePort) error {

	fmt.Printf("creating LB %s in service group %s\n", lbName, serviceGroupName)
	for _, port := range ports {
		fmt.Printf(" port: %s/%d\n", port.Protocol, port.Port)
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

	fmt.Printf("LB %s created successfully\n", lbName)

	return nil
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

// addRep adds a RemoteEndpoint to a LoadBalancer.
func addRep(ctx context.Context, cl client.Client, lbName string, serviceGroupName string, accountName string, address net.IP, port corev1.EndpointPort) error {
	var (
		err error
		lb  epicv1.LoadBalancer
	)

	// Get the LB that will own this endpoint
	if lb, err = getLB(ctx, cl, lbName, serviceGroupName, accountName); err != nil {
		return err
	}

	// Create the endpoint
	rep := epicv1.RemoteEndpoint{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: lb.Namespace,
			Name:      epicv1.RemoteEndpointName(address, port.Port, port.Protocol),
			Labels: map[string]string{
				epicv1.OwningLoadBalancerLabel: lb.Name,
				epicv1.OwningClusterLabel:      clusterName,
			},
		},
		Spec: epicv1.RemoteEndpointSpec{
			Cluster:     clusterName,
			Address:     address.String(),
			NodeAddress: address.String(),
			Port:        port,
		},
	}

	if err := cl.Create(ctx, &rep); err != nil {
		return err
	}

	return nil
}

func init() {
	var (
		ctx          context.Context = context.Background()
		account      string
		serviceGroup string
		ports        []string
	)

	// lbCmd manages load balancers. This is an alternative to PureLB's
	// LB management workflow because there are use cases where clients
	// want to manage load balancers but the back-end isn't a PureLB
	// cluster. It might be a legacy cluster that doesn't run in
	// Kubernetes.
	lbCmd := &cobra.Command{
		Use:     "loadbalancer",
		Aliases: []string{"lb"},
		Short:   "Manage EPIC load balancers",
		Long:    `Load balancers can be created using PureLB or epicctl. This command command manages load balancers and their endpoints when there's no PureLB back end.`,
	}
	lbCmd.PersistentFlags().StringVarP(&account, "account", "a", "", "(required) account in which the LB lives")
	lbCmd.PersistentFlags().StringVarP(&serviceGroup, "service-group", "g", "", "(required) service group to which the LB belongs")
	lbCmd.MarkPersistentFlagRequired("account")
	lbCmd.MarkPersistentFlagRequired("service-group")
	rootCmd.AddCommand(lbCmd)

	lbListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List load balancers",
		Long:    `List EPIC load balancers in the provided account and service group.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getEpicClient()
			if err != nil {
				return err
			}

			return listLB(ctx, cl, serviceGroup, account)
		},
	}
	lbCmd.AddCommand(lbListCmd)

	lbShowCmd := &cobra.Command{
		Use:     "show lb-name",
		Aliases: []string{"s"},
		Short:   "Show load balancer",
		Long:    `Show an EPIC load balancer.`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getEpicClient()
			if err != nil {
				return err
			}

			return showLB(ctx, cl, args[0], serviceGroup, account)
		},
	}
	lbCmd.AddCommand(lbShowCmd)

	lbCreateCmd := &cobra.Command{
		Use:     "create lb-name lb-ports",
		Aliases: []string{"c"},
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

			epPorts, err := parsePorts(ports)
			if err != nil {
				return err
			}
			return createLB(ctx, cl, args[0], serviceGroup, account, epPorts)
		},
	}
	lbCmd.AddCommand(lbCreateCmd)

	lbDeleteCmd := &cobra.Command{
		Use:     "delete lb-name",
		Aliases: []string{"d", "del", "rm"},
		Short:   "Delete load balancer",
		Long:    `Delete an EPIC load balancer.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getEpicClient()
			if err != nil {
				return err
			}

			return deleteLB(ctx, cl, args[0], serviceGroup, account)
		},
	}
	lbCmd.AddCommand(lbDeleteCmd)

	lbAddRepCmd := &cobra.Command{
		Use:   "add-rep lb-name rep-ip rep-port",
		Short: "Add remote endpoint to LB",
		Long: `Add an ad-hoc (i.e., non TrueIngress) remote endpoint to an LB.

Three arguments are required:
 - the load balancer name
 - the remote endpoint IP address
 - the remote endpoint port
   the port is a protocol and number, e.g., TCP/8080 or UDP/123.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse the IP address
			ip := net.ParseIP(args[1])
			if ip == nil {
				return fmt.Errorf("%s cannot be parsed as an IP address", args[1])
			}

			// Parse the endpoint port
			servPorts, err := parsePorts([]string{args[2]})
			if err != nil {
				return err
			}
			port := corev1.EndpointPort{
				Port:     int32(servPorts[0].Port),
				Protocol: servPorts[0].Protocol,
			}

			// Connect to the EPIC web service
			cl, err := getEpicClient()
			if err != nil {
				return err
			}

			return addRep(ctx, cl, args[0], serviceGroup, account, ip, port)
		},
	}
	lbCmd.AddCommand(lbAddRepCmd)

}
