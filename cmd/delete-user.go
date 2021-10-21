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
	deleteCmd.AddCommand(&cobra.Command{
		Use:     "api-user username user-namespace ",
		Aliases: []string{"api-user", "api-users"},
		Short:   "Create api-users",
		Long:    `Create api-user username in a specified user namespace`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getEpicClient()
			if err != nil {
				return err
			}

			return deleteAPIUser(rootCmd.Context(), cl, args[0], args[1])
		},
	})
}

// getAPIUser extracts and the api-usernames from the api-users secret it the
// user namespace

func deleteAPIUser(ctx context.Context, cl client.Client, apiUser string, accountName string) error {

	var newhttppasswd string

	secret := v1.Secret{}
	err := cl.Get(ctx, client.ObjectKey{Namespace: epicv1.AccountNamespace(accountName), Name: contourSecretName}, &secret)
	if err != nil {
		return fmt.Errorf("user namespace %s not found\n", accountName)
	}

	httppasswd := string(secret.Data["auth"])

	apiusernames := strings.Fields(httppasswd)

	for _, s := range apiusernames {
		user := s[:strings.IndexByte(s, ':')]
		if user != apiUser {
			newhttppasswd += s + "\n"
		}

	}

	secret.Data["auth"] = []byte(newhttppasswd)

	if err := cl.Update(ctx, &secret); err != nil {
		return err
	}

	fmt.Printf("api-user %s in user-namespace %s deleted \n", apiUser, accountName)

	return nil
}
