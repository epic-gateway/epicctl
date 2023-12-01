package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	epicv1 "epic-gateway.org/resource-model/api/v1"
)

func init() {
	var serviceGroup string

	// Set up this command and hook it into its parent, the create
	// command.
	cmd := cobra.Command{
		Use:     "ad-hoc-gateway name port",
		Short:   "Create ad-hoc Gateway",
		Aliases: []string{"ad-hoc-gw"},
		Long: `Create an ad-hoc EPIC Gateway.

EPIC can route traffic to ad-hoc Linux endpoints (i.e., endpoints that aren't managed by Kubernetes).

This command creates an ad-hoc Gateway, which is the first step towards building a cluster of ad-hoc endpoints.

Arguments:
 name - the Gateway's name (must be unique within your account)
 port - the port on which the Gateway will receive traffic (32-bit int)
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// We'll need a Client to interact with the Epic cluster.
			cl, err := getCRClient()
			if err != nil {
				panic(err.Error())
			}

			// Parse the port argument.
			if port, err := strconv.ParseInt(args[1], 10, 32); err != nil {
				return fmt.Errorf("can't parse %s as a port value: %w", args[1], err)
			} else {
				// Create the GWEndpointSlice.
				if err := createAdHocGateway(rootCmd.Context(), cl, accountName, args[0], int32(port), serviceGroup); err != nil {
					return err
				}
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&serviceGroup, "service-group", "gatewayhttp", "the service group to which the Gateway will belong")
	createCmd.AddCommand(&cmd)
}

// createAdHocGateway implements the behind-the-scenes work for the
// "ad-hoc-gateway" command. It's mostly just figuring out what we
// need, and then creating a GWProxy on EPIC.
func createAdHocGateway(ctx context.Context, cl crclient.Client, account string, name string, port int32, serviceGroup string) error {
	portNum := v1alpha2.PortNumber(port)
	proxy := epicv1.GWProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: epicv1.AccountNamespace(account),
			Labels: map[string]string{
				epicv1.OwningAccountLabel:        account,
				epicv1.OwningLBServiceGroupLabel: serviceGroup,
			},
		},
		Spec: epicv1.GWProxySpec{
			ClientRef: epicv1.ClientRef{
				Namespace: "adhoc", // Used in the DNS name
			},
			DisplayName: name, // Used in the DNS name
			Gateway: v1alpha2.GatewaySpec{
				Listeners: []v1alpha2.Listener{{
					Protocol: v1alpha2.HTTPProtocolType,
					Port:     portNum,
					Name:     "http",
				}},
			},
		},
	}

	// Create the GWProxy.

	if err := cl.Create(ctx, &proxy); err != nil {
		return err
	}

	// Poll until the GWProxy object gets its external address/name.
	for len(proxy.Spec.Endpoints) == 0 {
		time.Sleep(1 * time.Second)
		err := cl.Get(ctx, crclient.ObjectKeyFromObject(&proxy), &proxy)
		if err != nil {
			continue
		}
	}
	fmt.Printf("IP address: %s\n", proxy.Spec.Endpoints[0].Targets[0])
	fmt.Printf("DNS name: %s\n", proxy.Spec.Endpoints[0].DNSName)

	route := epicv1.GWRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: epicv1.AccountNamespace(account),
			Labels: map[string]string{
				epicv1.OwningAccountLabel: account,
			},
		},
		Spec: epicv1.GWRouteSpec{
			HTTP: &v1alpha2.HTTPRouteSpec{
				CommonRouteSpec: v1alpha2.CommonRouteSpec{
					ParentRefs: []v1alpha2.ParentReference{{
						Name: v1alpha2.ObjectName(name), // Link the Route to the Proxy
					}},
				},
				Rules: []v1alpha2.HTTPRouteRule{{
					BackendRefs: []v1alpha2.HTTPBackendRef{{
						BackendRef: v1alpha2.BackendRef{
							BackendObjectReference: v1alpha2.BackendObjectReference{
								Name: v1alpha2.ObjectName("linux-nodes"), // Link the Route to the backend cluster
								Port: &portNum,
							},
						},
					}},
				}},
			},
		},
	}

	// Create the GWRoute.
	if err := cl.Create(ctx, &route); err != nil {
		return err
	}

	return nil
}
