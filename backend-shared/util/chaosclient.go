package util

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	_ client.Client = &ChaosClient{}
)

type ChaosClient struct {
	InnerClient client.Client
}

func isEnvExist(key string) bool {
	if _, ok := os.LookupEnv(key); ok {
		return true
	}

	return false
}

// IfEnabledSimulateUnreliableClient will wrap a regular client.Client, and randomly simulate connection issues.
func IfEnabledSimulateUnreliableClient(kClient client.Client) client.Client {
	// set environment variable(ENABLE_UNRELIABLE_CLIENT) in your terminal before running e2e tests
	if os.Getenv("ENABLE_UNRELIABLE_CLIENT") == "true" {
		return &ChaosClient{InnerClient: kClient}
	} else {
		return kClient
	}
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
// obj must be a struct pointer so that obj can be updated with the response
// returned by the Server.
func (pc *ChaosClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {

	res := pc.InnerClient.Get(ctx, key, obj)

	if err := shouldSimulateFailure(); err != nil {
		return err
	}

	return res
}

// List retrieves list of objects for a given namespace and list options. On a
// successful call, Items field in the list will be populated with the
// result returned from the server.
func (pc *ChaosClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	res := pc.InnerClient.List(ctx, list, opts...)

	if err := shouldSimulateFailure(); err != nil {
		return err
	}

	return res
}

// Create saves the object obj in the Kubernetes cluster.
func (pc *ChaosClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	res := pc.InnerClient.Create(ctx, obj, opts...)

	if err := shouldSimulateFailure(); err != nil {
		return err
	}

	return res
}

// Delete deletes the given obj from Kubernetes cluster.
func (pc *ChaosClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	res := pc.InnerClient.Delete(ctx, obj, opts...)

	if err := shouldSimulateFailure(); err != nil {
		return err
	}

	return res
}

// Update updates the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (pc *ChaosClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	res := pc.InnerClient.Update(ctx, obj, opts...)

	if err := shouldSimulateFailure(); err != nil {
		return err
	}

	return res
}

// Patch patches the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (pc *ChaosClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	res := pc.InnerClient.Patch(ctx, obj, patch, opts...)

	if err := shouldSimulateFailure(); err != nil {
		return err
	}

	return res
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (pc *ChaosClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	res := pc.InnerClient.DeleteAllOf(ctx, obj, opts...)

	if err := shouldSimulateFailure(); err != nil {
		return err
	}

	return res
}

// StatusWriter knows how to update status subresource of a Kubernetes object.
func (pc *ChaosClient) Status() client.StatusWriter {
	res := pc.InnerClient.Status()
	return res
}

// Scheme returns the scheme this client is using.
func (pc *ChaosClient) Scheme() *runtime.Scheme {
	res := pc.InnerClient.Scheme()
	return res

}

// RESTMapper returns the rest this client is using.
func (pc *ChaosClient) RESTMapper() meta.RESTMapper {
	res := pc.InnerClient.RESTMapper()
	return res
}

type ChaosClientStatusWrapper struct {
	innerWriter *client.StatusWriter
	// parent      *ProxyClient
}

// Update updates the fields corresponding to the status subresource for the
// given obj. obj must be a struct pointer so that obj can be updated
// with the content returned by the Server.
func (pcsw *ChaosClientStatusWrapper) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {

	res := (*pcsw.innerWriter).Update(ctx, obj, opts...)

	if err := shouldSimulateFailure(); err != nil {
		return err
	}

	return res

}

// Patch patches the given object's subresource. obj must be a struct
// pointer so that obj can be updated with the content returned by the
// Server.
func (pcsw *ChaosClientStatusWrapper) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {

	if err := shouldSimulateFailure(); err != nil {
		return err
	}

	res := (*pcsw.innerWriter).Patch(ctx, obj, patch, opts...)

	return res
}

func shouldSimulateFailure() error {

	if !isEnvExist("UNRELIABLE_CLIENT_FAILURE_RATE") {
		return nil
	}

	// set environment variable(UNRELIABLE_CLIENT_FAILURE_RATE) in your terminal before running e2e tests
	unreliableClientFailureRate := os.Getenv("UNRELIABLE_CLIENT_FAILURE_RATE")
	unreliableClientFailureRateValue, err := strconv.Atoi(unreliableClientFailureRate)
	if err != nil {
		return err
	}

	// Convert it to a decimal point, e.g. 50% to 0.5
	unreliableClientFailureRateValue = unreliableClientFailureRateValue / 100

	// return an error randomly x% of the time (using math/rand package)
	if rand.Float64() < float64(unreliableClientFailureRateValue) {
		return fmt.Errorf("simulated K8s error")
	}

	return nil

}
