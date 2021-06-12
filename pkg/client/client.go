package client

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // This is required for client-go to know about the kubectl oidc authenticator, do not remove
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var (
	client *Client
)

type Client struct {
	clientset     kubernetes.Interface
	restConfig    *restclient.Config
	dynamicClient dynamic.Interface
	namespace     string
}

func GetRestConfig() (*restclient.Config, string, error) {

	getter := genericclioptions.NewConfigFlags(true)
	factory := cmdutil.NewFactory(getter)
	namespace, _, err := factory.ToRawKubeConfigLoader().Namespace()

	if err != nil {
		return nil, "", err
	}

	clientConfig := factory.ToRawKubeConfigLoader()
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, "", err
	}

	return restConfig, namespace, nil
}

func GetClient() (*Client, error) {
	if client != nil {
		return client, nil
	}

	restConfig, namespace, err := GetRestConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		namespace:     namespace,
		restConfig:    restConfig,
		clientset:     clientset,
		dynamicClient: dynamicClient,
	}, nil
}

func (c *Client) GetDynamicClient() dynamic.Interface {
	return c.dynamicClient
}

func (c *Client) GetClientset() kubernetes.Interface {
	return c.clientset
}

func (c *Client) GetRestConfig() *restclient.Config {
	return c.restConfig
}

func (c *Client) GetDefaultNamespace() string {
	return c.namespace
}

func (c *Client) SetCurrentNamespace(namespace string) {

}

func (c *Client) SetClientset(clientset kubernetes.Interface) {
	c.clientset = clientset
}

func (c *Client) SetDefaultNamespace(namespace string) error {
	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	config, err := configAccess.GetStartingConfig()

	if err != nil {
		return err
	}

	context := config.Contexts[config.CurrentContext]
	context.Namespace = namespace

	err = clientcmd.ModifyConfig(configAccess, *config, true)
	return err
}
