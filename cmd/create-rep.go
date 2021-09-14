package cmd

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

func init() {
	createRepCmd := &cobra.Command{
		Use:     "remoteendpoint lb-name rep-ip rep-port",
		Aliases: []string{"rep"},
		Short:   "Adds remote endpoints to load balancers",
		Long: `Adds ad-hoc (i.e., non TrueIngress) remote endpoints to load balancers.

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
			servPorts, err := parsePorts(args[2])
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

			account, serviceGroup, err := getAccountAndSG()
			if err != nil {
				return err
			}

			return addRep(rootCmd.Context(), cl, args[0], serviceGroup, account, ip, port)
		},
	}
	bindAccountAndSG(createRepCmd)
	createCmd.AddCommand(createRepCmd)
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
			},
		},
		Spec: epicv1.RemoteEndpointSpec{
			Cluster: clusterName,
			Address: address.String(),
			Port:    port,
			// NodeAddress not set means that this is an ad-hoc endpoint
		},
	}

	if err := cl.Create(ctx, &rep); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Endpoint created\n")

	return nil
}
