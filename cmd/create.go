/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var uuidoverride bool

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new resource",
	Long:  `Add new resources to the EPIC`,
}

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Create Organization",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) != 1 {
			fmt.Println("org name expected")

			return fmt.Errorf("org name expected")

		}

		createorg(args[0])

		return nil
	},
}

func createorg(orgname string) {

	var nsid string

	organn := make(map[string]string)

	organn["acnodal.io/org"] = orgname

	if uuidoverride {
		nsid = orgname
	} else {
		nsid = uuid.New().String()
	}

	ns := &v1.Namespace{

		ObjectMeta: metav1.ObjectMeta{
			Name:        nsid,
			Annotations: organn,
		},
	}

	clientset, err := getClient()
	if err != nil {
		panic(err.Error())
	}

	_, err = clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})

}

var apiacctCmd = &cobra.Command{
	Use:   "apiacct",
	Short: "Create k8s api account",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("apiacct called")
	},
}

var serviceprefixCmd = &cobra.Command{
	Use:   "serviceprefix",
	Short: "Create Service Prefix",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serviceprefix called")
	},
}
var lbgCmd = &cobra.Command{
	Use:   "lbg",
	Short: "Create Load Balancer Group",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("lbg called")
	},
}
var envoytemplateCmd = &cobra.Command{
	Use:   "envoytemplate",
	Short: "Create Envoy Template",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("envoytemplate called")
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(orgCmd)
	createCmd.AddCommand(apiacctCmd)
	createCmd.AddCommand(serviceprefixCmd)
	createCmd.AddCommand(lbgCmd)
	createCmd.AddCommand(envoytemplateCmd)

	orgCmd.Flags().BoolVarP(&uuidoverride, "uuidoverride", "", false, "Use Org name instead of UUID")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
