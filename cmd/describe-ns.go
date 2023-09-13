package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "epic-gateway.org/resource-model/api/v1"
)

func init() {
	describeNSCmd := &cobra.Command{
		Use:     "user-namespace",
		Aliases: []string{"ns"},
		Short:   "Describes a user namespace",
		Long:    `Describes an EPIC user namespace.`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := getGoClientset()
			if err != nil {
				return err
			}
			cl, err := getCRClient()
			if err != nil {
				return err
			}

			account := args[0]

			return describeNS(context.Background(), cl, cs, account)
		},
	}
	describeCmd.AddCommand(describeNSCmd)
}

// describeNS gets a LoadBalancer from the cluster and dumps its contents
// to stdout.
func describeNS(ctx context.Context, cl client.Client, cs *kubernetes.Clientset, nsName string) error {
	var (
		err  error
		acct epicv1.Account
	)

	if acct, err = getAccount(ctx, cl, nsName); err != nil {
		return err
	}

	// NS Info
	Debug(" Raw CR Contents: %+v\n", acct)
	fmt.Printf("EPIC User Namespace %s\n\n", nsName)

	fmt.Printf("API Users\n")
	listAPIUsers(ctx, cl, nsName)

	fmt.Printf("Gateways\n")
	showProxies(ctx, cl, nsName)

	fmt.Printf("Web Service Activity\n")
	if err = showActivity(ctx, cs, nsName); err != nil {
		return err
	}

	return nil
}

func getAccount(ctx context.Context, cl client.Client, accountName string) (acct epicv1.Account, err error) {
	return acct, cl.Get(ctx, client.ObjectKey{Namespace: epicv1.AccountNamespace(accountName), Name: accountName}, &acct)
}

func showProxies(ctx context.Context, cl client.Client, accountName string) error {
	proxies := epicv1.GWProxyList{}
	err := cl.List(ctx, &proxies, &client.ListOptions{Namespace: epicv1.AccountNamespace(accountName)})
	if err != nil {
		return fmt.Errorf("user namespace %s not found", accountName)
	}

	for _, p := range proxies.Items {
		fmt.Printf("  %s\n", p.Name)
	}

	return nil
}

func showActivity(ctx context.Context, cs *kubernetes.Clientset, accountName string) error {
	podLogOptions := v1.PodLogOptions{
		SinceSeconds: pointer.Int64Ptr(300),
		Timestamps:   true,
	}

	wsPod, err := getWebServicePod(ctx, cs)
	if err != nil {
		return err
	}

	podLogRequest := cs.CoreV1().
		Pods(wsPod.Namespace).GetLogs(wsPod.Name, &podLogOptions)
	stream, err := podLogRequest.Stream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		if line := scanner.Text(); strings.Contains(line, accountName) {
			fmt.Println(line)
		}
	}

	return nil
}

func getWebServicePod(ctx context.Context, cs *kubernetes.Clientset) (pod v1.Pod, err error) {
	pods, err := cs.CoreV1().Pods("epic").List(ctx, metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=epic,app.kubernetes.io/component=web-service"})
	if err != nil {
		return pod, err
	}

	if len(pods.Items) < 1 {
		return pod, fmt.Errorf("web service pod not found")
	}

	return pods.Items[0], nil
}
