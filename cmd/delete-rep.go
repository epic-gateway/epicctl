package cmd

import (
	"context"
	"fmt"
	"net"

	"github.com/spf13/cobra"
	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	deleteRepCmd := &cobra.Command{
		Use:     "remoteendpoint lb-name rep-ip rep-port user-namespace service-group",
		Aliases: []string{"rep"},
		Short:   "Deletes ad-hoc remote endpoints",
		Long: `Deletes ad-hoc (i.e., non TrueIngress) remote endpoints.

Three arguments are required:
 - the load balancer name
 - the remote endpoint IP address
 - the remote endpoint port
   the port is a protocol and number, e.g., TCP/8080 or UDP/123
 - the user namespace
 - the service group to which the load balancer belongs.
`,
		Args: cobra.ExactArgs(5),
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

			account := args[3]
			serviceGroup := args[4]

			// Connect to the EPIC web service
			cl, err := getEpicClient()
			if err != nil {
				return err
			}

			return deleteRep(rootCmd.Context(), cl, args[0], serviceGroup, account, ip, port)
		},
	}
	deleteCmd.AddCommand(deleteRepCmd)
}

// deleteRep deletes an ad-hoc remote endpoint.
func deleteRep(ctx context.Context, cl client.Client, lbName string, serviceGroupName string, accountName string, address net.IP, port corev1.EndpointPort) error {
	list, err := getReps(ctx, cl, lbName, serviceGroupName, accountName)
	if err != nil {
		return err
	}

	// Each rep has a random suffix so similar ones don't collide (e.g,
	// two different lbsg's each with an lb with the same name and each
	// of those with a rep with the same params). Therefore we need to
	// scan the reps to find the one we want to delete.
	var toBeDeleted *epicv1.RemoteEndpoint = nil
	for _, rep := range list.Items {
		if rep.Spec.Address == address.String() && rep.Spec.Port == port {
			toBeDeleted = &rep
			break
		}
	}
	if toBeDeleted == nil {
		// We didn't find a rep that matches the arguments
		return fmt.Errorf("remote endpoint not found")
	}

	// We found the rep that the user specified so we can delete it
	if err := cl.Delete(ctx, toBeDeleted); err != nil {
		return err
	}

	return nil
}
