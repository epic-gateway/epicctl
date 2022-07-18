package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
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

			fmt.Printf("EPIC User Namespaces\n")
			return showUserNamespaces(rootCmd.Context(), cs)
		},
	})
}

// showUserNamespaces extracts and prints the names of the user
// namespaces.
func showUserNamespaces(ctx context.Context, cs *kubernetes.Clientset) error {
	nsPrefix := epicv1.ProductName + "-"

	nsList, err := cs.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/part-of=" + epicv1.ProductName + ",app.kubernetes.io/component=user-namespace",
	})
	if err != nil {
		return err
	}

	for _, ns := range nsList.Items {
		fmt.Printf(" %s\n", strings.TrimPrefix(ns.Name, nsPrefix))
	}
	return nil
}
