package cmd

import (
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

func init() {
	// Set up the user-namespace command and hook it into its parent,
	// the create command.
	createCmd.AddCommand(&cobra.Command{
		Use:     "servicegroup-manifest epic-lbsg epic-host epic-user-ns purelb-group ws-user ws-password",
		Short:   "Create ServiceGroup Manifest",
		Aliases: []string{"sg"},
		Long: `Create a PureLB ServiceGroup manifest.

PureLB uses ServiceGroup custom resources to link to EPIC.

This command creates a PureLB ServiceGroup manifest that can be loaded
into the client Kubernetes cluster. It will enable the creation of Load
Balancers in this User Namespace.

Arguments:
 epic-lbsg    - the name of the EPIC LBServiceGroup that this PureLB
                ServiceGroup will reference
 epic-host    - the hostname of the EPIC web service
 epic-user-ns - the name of the EPIC User Namespace
 purelb-group - the name to use for the new ServiceGroup
 ws-user      - the EPIC web service username
 ws-password  - the EPIC web service password
`,
		Args: cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createServiceGroup(args[0], args[1], args[2], args[3], args[4], args[5])
		},
	})
}

// createServiceGroup implements the behind-the-scenes work for the
// "servicegroup-manifest" command. It's mostly just formatting a YAML
// fragment with the parameters provided and dumping it to stdout.
func createServiceGroup(lbsgName string, hostName string, orgName string, groupName string, wsUserName string, wsPassword string) error {

	tmpl, err := template.New("ServiceGroup").Parse(`---
apiVersion: purelb.io/v1
kind: ServiceGroup
metadata:
  name: {{.group}}
  namespace: purelb
spec:
  epic:
    api-service-hostname: {{.host}}
    api-service-username: {{.user}}
    api-service-password: {{.password}}
    user-namespace: {{.namespace}}
    lbservicegroup: {{.lbsg}}
`)
	if err != nil {
		return err
	}
	err = tmpl.Execute(
		os.Stdout,
		map[string]string{
			"group":     groupName,
			"lbsg":      lbsgName,
			"host":      hostName,
			"user":      wsUserName,
			"password":  wsPassword,
			"namespace": orgName,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
