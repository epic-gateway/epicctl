package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

func init() {
	getCmd.AddCommand(&cobra.Command{
		Use:     "api-user user-namespace ",
		Aliases: []string{"api-user", "api-users"},
		Short:   "Get api-users",
		Long:    `Get api-users in a specified user namespace`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getEpicClient()
			if err != nil {
				return err
			}

			return getAPIUser(rootCmd.Context(), cl, args[0])
		},
	})
}

// getAPIUser extracts and the api-usernames from the api-users secret it the
// user namespace

func getAPIUser(ctx context.Context, cl client.Client, accountName string) error {

	secret := v1.Secret{}
	err := cl.Get(ctx, client.ObjectKey{Namespace: epicv1.AccountNamespace(accountName), Name: contourSecretName}, &secret)
	if err != nil {
		return fmt.Errorf("user namespace %s not found", accountName)
	}

	fmt.Printf("EPIC API Users in User Namespace %s\n", accountName)

	httppasswd := string(secret.Data["auth"])

	apiusernames := strings.Fields(httppasswd)

	for _, s := range apiusernames {
		user := s[:strings.IndexByte(s, ':')]
		fmt.Printf("  %s\n", user)
	}

	return nil
}