package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

func init() {
	createCmd.AddCommand(&cobra.Command{
		Use:     "api-user username user-namespace ",
		Aliases: []string{"api-user", "api-users"},
		Short:   "Create api-users",
		Long:    `Create api-user username in a specified user namespace`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := getCRClient()
			if err != nil {
				return err
			}

			return createAPIUser(rootCmd.Context(), cl, args[0], args[1])
		},
	})
}

// getAPIUser extracts and the api-usernames from the api-users secret it the
// user namespace

func createAPIUser(ctx context.Context, cl client.Client, apiUser string, accountName string) error {

	secret := v1.Secret{}
	err := cl.Get(ctx, client.ObjectKey{Namespace: epicv1.AccountNamespace(accountName), Name: contourSecretName}, &secret)
	if err != nil {
		return fmt.Errorf("user namespace %s not found\n", accountName)
	}

	httppasswd := string(secret.Data["auth"])

	apiusernames := strings.Fields(httppasswd)

	for _, s := range apiusernames {
		user := s[:strings.IndexByte(s, ':')]
		if user == apiUser {
			return fmt.Errorf("api-user %s exists", apiUser)
		}

	}

	fmt.Print("New Password:  ")
	pass1, err := readPassword()
	if err != nil {
		panic(err.Error())
	}

	fmt.Print("\nRetype New Password:  ")
	pass2, err := readPassword()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("")

	if len(pass1) < 6 {
		fmt.Println("minimum password length 6 characters")
		os.Exit(1)
	}

	if pass1 != pass2 {
		fmt.Println("passwords don't match")
		os.Exit(1)
	}

	pwBytes, _ := bcrypt.GenerateFromPassword([]byte(pass2), bcrypt.DefaultCost)

	newapiuser := fmt.Sprintf("%s:%s\n", apiUser, string(pwBytes))

	newhttppasswd := httppasswd + newapiuser

	secret.Data["auth"] = []byte(newhttppasswd)

	if err := cl.Update(ctx, &secret); err != nil {
		return err
	}

	fmt.Printf("api-user %s in user-namespace %s created\n", apiUser, accountName)

	return nil
}

// Reads a password from stdin. If stdin is a terminal then it uses
// term.ReadPassword so the terminal doesn't echo the password.  If
// stdin isn't a terminal (e.g., stdin has been redirected) then it
// just scans.
func readPassword() (string, error) {
	var (
		pw    string
		err   error
		stdin int = int(syscall.Stdin)
	)

	if term.IsTerminal(stdin) {
		var bytepw []byte
		bytepw, err = term.ReadPassword(stdin)
		pw = string(bytepw)
	} else {
		_, err = fmt.Scanln(&pw)
	}

	return pw, err
}
