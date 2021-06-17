package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

var (
	debug     bool
	cfgFile   string
	k8sConfig string
	scheme    = runtime.NewScheme()
)

// rootCmd represents the base command when calledp without any subcommands
var rootCmd = &cobra.Command{
	Use:   "epicctl",
	Short: "epicctl controls the EPIC cluster",
	Long:  `epicctl controls the EPIC cluster.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		panic(err)
	}
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(epicv1.AddToScheme(scheme))

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "very verbose output")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", path.Join(homedir.HomeDir(), ".epicctl.yaml"), "epicctl config file")
	rootCmd.PersistentFlags().StringVar(&k8sConfig, clientcmd.RecommendedConfigPathFlag, clientcmd.RecommendedHomeFile, "k8s config file")
}

func getEpicClient() (client.Client, error) {
	config, err := getClientConfig()
	if err != nil {
		return nil, err
	}

	return client.New(config, client.Options{
		Scheme: scheme,
	})
}

func getClientConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if k8sConfig != "" {
		loadingRules.Precedence = append(loadingRules.Precedence, k8sConfig)
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil)

	// use the current context in kubeconfig
	return kubeConfig.ClientConfig()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv() // read in environment variables that match

	Debug("Running epicctl\n")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		Debug("Using config file: %s\n", viper.ConfigFileUsed())
	} else {
		Debug("Problem reading config file: %+v\n", err)
	}
}

// Debug is like fmt.Printf(os.Stderr...) except it only outputs if
// the debug command-line flag was set.
func Debug(format string, kvs ...interface{}) {
	if debug {
		fmt.Fprintf(os.Stderr, format, kvs...)
	}
}
