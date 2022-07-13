package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

var (
	scheme = runtime.NewScheme()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "epicctl",
	Short: "epicctl controls EPIC clusters",
	Long:  `epicctl controls EPIC clusters.`,
}

// Execute is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	Debug("Running epicctl\n")

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		if viper.GetBool("debug") {
			panic(err)
		}
		os.Exit(1)
	}
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(epicv1.AddToScheme(scheme))

	cobra.OnInitialize(readConfigFile)

	rootCmd.PersistentFlags().Bool("debug", false, "enable debug output")
	rootCmd.PersistentFlags().String("config", path.Join(homedir.HomeDir(), ".epicctl.yaml"), "epicctl config file")
	rootCmd.PersistentFlags().String(clientcmd.RecommendedConfigPathFlag, clientcmd.RecommendedHomeFile, "k8s config file")
	viper.BindPFlags(rootCmd.PersistentFlags())
	viper.BindPFlags(rootCmd.Flags())
	viper.SetEnvPrefix("EPICCTL")
	viper.AutomaticEnv()
	// This lets us name our command-line flags with dashes but have the
	// environment variables use underscores (e.g., "service-group" and
	// "SERVICE_GROUP")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}

// getCRClient creates a new Controller Runtime client.Client.
func getCRClient() (client.Client, error) {
	config, err := getClientConfig()
	if err != nil {
		return nil, err
	}

	return client.New(config, client.Options{
		Scheme: scheme,
	})
}

// getGoClientset creates a new client-go Clientset.
func getGoClientset() (*kubernetes.Clientset, error) {
	config, err := getClientConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func getClientConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if k8sConfig := viper.GetString(clientcmd.RecommendedConfigPathFlag); k8sConfig != "" {
		loadingRules.Precedence = append(loadingRules.Precedence, k8sConfig)
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil)

	// use the current context in kubeconfig
	return kubeConfig.ClientConfig()
}

// readConfigFile reads the config file.
func readConfigFile() {
	viper.SetConfigFile(viper.GetString("config"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		Debug("Using config file %s\n", viper.ConfigFileUsed())
	} else {
		Debug("Problem reading config file: %+v\n", err)
	}
}

// Debug is like fmt.Printf(os.Stderr...) except it only outputs if
// the debug flag is set.
func Debug(format string, kvs ...interface{}) {
	if viper.GetBool("debug") {
		fmt.Fprintf(os.Stderr, format, kvs...)
	}
}
