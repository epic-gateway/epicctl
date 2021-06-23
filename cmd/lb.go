package cmd

import (
	"context"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	epicv1 "gitlab.com/acnodal/epic/resource-model/api/v1"
)

const (
	clusterName = "epicctl"
)

// parsePorts parses an array of strings and returns an array of
// corev1.ServicePort. Each string can either be "proto/port" or just
// "port". The proto must be either "TCP" or "UDP" and the portNum
// must fit in a 32-bit signed integer. If the proto is missing it
// defaults to "TCP". If error is non-nil than something has gone
// wrong and the ServicePort array is undefined.
func parsePorts(ports string) ([]corev1.ServicePort, error) {
	epPorts := make([]corev1.ServicePort, strings.Count(ports, ",")+1)

	for i, port := range strings.Split(ports, ",") {
		parts := strings.Split(port, "/")
		if len(parts) == 2 {
			// if there's a "/" then it's a protocol and port num, e.g., "TCP/8080"
			portNum, err := strconv.ParseInt(parts[1], 10, 32)
			if err != nil {
				return epPorts, err
			}
			epPorts[i] = corev1.ServicePort{
				Protocol: corev1.Protocol(parts[0]),
				Port:     int32(portNum),
			}
		} else {
			// if there's no "/" then it's just a port num, e.g., "8080"
			portNum, err := strconv.ParseInt(parts[0], 10, 32)
			if err != nil {
				return epPorts, err
			}
			epPorts[i] = corev1.ServicePort{
				Protocol: corev1.ProtocolTCP,
				Port:     int32(portNum),
			}
		}
	}
	return epPorts, nil
}

func getLB(ctx context.Context, cl client.Client, lbName string, serviceGroupName string, accountName string) (epicv1.LoadBalancer, error) {
	var err error = nil
	lb := epicv1.LoadBalancer{}
	lb.Name = epicv1.LoadBalancerName(serviceGroupName, lbName, true)
	err = cl.Get(ctx, client.ObjectKey{Namespace: epicv1.AccountNamespace(accountName), Name: lb.Name}, &lb)

	return lb, err
}

func getReps(ctx context.Context, cl client.Client, lbName string, serviceGroupName string, accountName string) (epicv1.RemoteEndpointList, error) {
	listOps := client.ListOptions{
		Namespace:     epicv1.AccountNamespace(accountName),
		LabelSelector: labels.SelectorFromSet(map[string]string{epicv1.OwningLoadBalancerLabel: epicv1.LoadBalancerName(serviceGroupName, lbName, true)}),
	}
	list := epicv1.RemoteEndpointList{}
	err := cl.List(ctx, &list, &listOps)
	return list, err
}
