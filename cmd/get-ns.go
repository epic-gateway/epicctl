package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "epic-gateway.org/resource-model/api/v1"
)

func init() {
	getCmd.AddCommand(&cobra.Command{
		Use:     "user-namespaces",
		Aliases: []string{"ns"},
		Short:   "Get user-namespaces",
		Long:    `Get user-namespaces`,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := getGoClientset()
			if err != nil {
				return err
			}
			cl, err := getCRClient()
			if err != nil {
				return err
			}

			return showUserNamespaces(rootCmd.Context(), cs, cl)
		},
	})
}

// showUserNamespaces extracts and prints the names of the user
// namespaces.
func showUserNamespaces(ctx context.Context, cs *kubernetes.Clientset, cl client.Client) error {
	nsPrefix := epicv1.ProductName + "-"

	// Fetch the namespaces that are part of EPIC.
	nsList, err := cs.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/part-of=" + epicv1.ProductName + ",app.kubernetes.io/component=user-namespace",
	})
	if err != nil {
		return err
	}

	// Sort by NS creation time (newest first).
	sort.Slice(nsList.Items, func(i, j int) bool {
		return nsList.Items[j].CreationTimestamp.Before(&nsList.Items[i].CreationTimestamp)
	})

	// Set up output table.
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"EPIC User NS", "Created At", "GWP Count"})
	table.SetAutoFormatHeaders(false)
	table.SetBorder(false)

	for _, ns := range nsList.Items {

		// Fetch the GWP count
		proxies := epicv1.GWProxyList{}
		err := cl.List(ctx, &proxies, &client.ListOptions{Namespace: ns.Name})
		if err != nil {
			return fmt.Errorf("user namespace %s not found", ns.Name)
		}

		// Add a row to the output table
		table.Append([]string{
			strings.TrimPrefix(ns.Name, nsPrefix),
			ns.CreationTimestamp.String(),
			strconv.Itoa(len(proxies.Items)),
		})
	}

	table.Render()

	return nil
}
