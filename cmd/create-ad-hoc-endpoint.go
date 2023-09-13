package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "epic-gateway.org/resource-model/api/v1"
)

var (
	hostName       string
	hostAddress    net.IP
	hostPort       int32
	clusterName    string
	createAdhocCmd = cobra.Command{
		Use:     "ad-hoc-endpoint address port",
		Short:   "Create ad-hoc endpoint",
		Aliases: []string{"ad-hoc"},
		Long: `Create an ad-hoc EPIC endpoint.

EPIC can route traffic to ad-hoc Linux endpoints (i.e., endpoints that aren't managed by Kubernetes).

This command adds this node to a cluster of ad-hoc endpoints.

Arguments:
 address - the IP address to which EPIC will send traffic
 port - the port to which EPIC will send traffic
`,
		Args:    cobra.ExactArgs(2),
		PreRunE: parseInput,
		RunE: func(cmd *cobra.Command, args []string) error {
			// We'll need a Client to interact with the Epic cluster.
			cl, err := getCRClient()
			if err != nil {
				panic(err.Error())
			}

			// Create the GWEndpointSlice.
			if err := createAdHocEndpoint(rootCmd.Context(), cl, accountName, clusterName, hostName, hostAddress, hostPort); err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	// Figure out what hostName we'll use as a default.
	hostName, _ = os.Hostname()

	createAdhocCmd.Flags().StringVar(&hostName, "host-name", hostName, "node's hostname")
	createAdhocCmd.Flags().StringVar(&clusterName, "cluster-name", "linux-nodes", "name of the endpoint cluster to which this node will belong")
	// Set up this command and hook it into its parent, the create
	// command.
	createCmd.AddCommand(&createAdhocCmd)
}

func parseInput(cmd *cobra.Command, args []string) error {
	// We don't need to validate the hostAddress
	hostAddress = net.ParseIP(args[0])
	if hostAddress == nil {
		return fmt.Errorf("can't parse %s as an IP address", args[0])
	}

	// Parse the port argument.
	if port, err := strconv.ParseInt(args[1], 10, 32); err != nil {
		return fmt.Errorf("can't parse %s as a port value: %w", args[1], err)
	} else {
		hostPort = int32(port)
	}

	return nil
}

// createAdHocEndpoint implements the behind-the-scenes work for the
// "ad-hoc-endpoint" command. It's mostly just figuring out what we
// need from the node on which we're running, and then creating a
// GWEndpointSlice on Epic.
func createAdHocEndpoint(ctx context.Context, cl crclient.Client, groupName string, clusterName string, hostName string, ipAddress net.IP, port int32) error {
	proto := v1.ProtocolTCP
	slice := epicv1.GWEndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostName,
			Namespace: epicv1.AccountNamespace(groupName),
			Labels: map[string]string{
				epicv1.OwningAccountLabel: groupName,
			},
		},
		Spec: epicv1.GWEndpointSliceSpec{
			ParentRef: epicv1.ClientRef{
				UID: clusterName,
			},
			EndpointSlice: discoveryv1.EndpointSlice{
				AddressType: discoveryv1.AddressTypeIPv4,
				Endpoints: []discoveryv1.Endpoint{
					{
						NodeName:  pointer.StringPtr(hostName),
						Addresses: []string{hostAddress.String()},
					},
				},
				Ports: []discoveryv1.EndpointPort{{
					Port:     &port,
					Protocol: &proto,
				}},
			},
			NodeAddresses: map[string]string{
				hostName: hostAddress.String(),
			},
		},
	}

	err := cl.Create(ctx, &slice)
	if err != nil {
		return err
	}

	return nil
}
