package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/generate/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

const (
	gitlabSecretName       = "gitlab"
	gitlabRegistryHostname = "registry.gitlab.com"
	contourSecretName      = "password"
	contourRealmName       = "epicauth"
)

func init() {
	// Set up the user-namespace command and hook it into its parent,
	// the create command.
	createCmd.AddCommand(&cobra.Command{
		Use:     "user-namespace name registry-user registry-password ws-user ws-password",
		Short:   "Create User Namespace",
		Aliases: []string{"ns", "user-ns"},
		Long: `Create an EPIC User Namespace.

EPIC is a multi-tenant system. Groups of LoadBalancers used by a
team are configured and managed within a User Namespace.

This command creates a User Namespace. The name can contain only
alphanumeric characters and the dash "-". Contact Acnodal support
for your registry-user and registry-password.`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getEpicClient()
			if err != nil {
				panic(err.Error())
			}

			return createUserNamespace(context.Background(), client, args[0], args[1], args[2], args[3], args[4])
		},
	})
}

// createUserNamespace implements the behind-the-scenes work for the
// "user-namespace" command. Each new namespace requires some
// infratructure for various purposes. This sets up the minimal
// infrastructure that's always needed like an Account CR and the
// various secrets needed for Docker and Contour.
func createUserNamespace(ctx context.Context, cl client.Client, orgName string, registryUserName string, registryPassword string, wsUserName string, wsPassword string) error {

	nsName := epicv1.AccountNamespace(orgName)

	// Create the k8s Namespace
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nsName,
			Labels: epicv1.UserNSLabels,
		},
	}
	if err := cl.Create(ctx, &ns); err != nil {
		return err
	}

	// Every EPIC user namespace has an Account CR. Create that.
	acct := epicv1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName,
			Name:      orgName,
		},
		Spec: epicv1.AccountSpec{},
	}
	if err := cl.Create(ctx, &acct); err != nil {
		return err
	}

	// We use a private Docker registry for our proprietary images so
	// every user NS needs a k8s secret containing credentials to access
	// that registry.
	secret, err := dockerSecret(gitlabSecretName, nsName, gitlabRegistryHostname, registryUserName, registryPassword)
	if err != nil {
		return err
	}
	if err := cl.Create(ctx, &secret); err != nil {
		return err
	}

	// Contour (EPIC's web service authn proxy) looks for credentials in
	// a "password" secret in the user NS so we need to add this NS's WS
	// credentials to a secret in this NS.
	pwObj, err := contourSecret(contourSecretName, nsName, contourRealmName, wsUserName, wsPassword)
	if err != nil {
		return err
	}
	if err := cl.Create(ctx, &pwObj); err != nil {
		return err
	}

	return nil
}

// dockerSecret generates a k8s Secret to allow k8s to access our
// private registry on gitlab.
func dockerSecret(name string, ns string, host string, user string, password string) (v1.Secret, error) {
	secObj, err := versioned.SecretForDockerRegistryGeneratorV1{
		Name:     name,
		Server:   host,
		Username: user,
		Password: password,
	}.StructuredGenerate()
	if err != nil {
		return v1.Secret{}, err
	}
	secret := secObj.(*v1.Secret)
	secret.Namespace = ns

	return *secret, nil
}

// contourSecret generates a Contour-compatible k8s Secret.
func contourSecret(name string, ns string, realm string, user string, password string) (v1.Secret, error) {
	pwObj := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Annotations: map[string]string{
				"projectcontour.io/auth-type":  "basic",
				"projectcontour.io/auth-realm": realm,
			},
		},
	}
	pwBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return pwObj, err
	}
	pwObj.Data = map[string][]byte{
		"auth": []byte(fmt.Sprintf("%s:%s", user, string(pwBytes))),
	}

	return pwObj, nil
}
