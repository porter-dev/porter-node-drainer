package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/joeshaw/envdecode"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

type EKSCredentials struct {
	AccessKeyID   string `env:"EKS_AWS_ACCESS_KEY_ID"`
	SecretKey     string `env:"EKS_AWS_SECRET_ACCESS_KEY"`
	Region        string `env:"EKS_AWS_REGION"`
	ClusterID     string `env:"EKS_AWS_CLUSTER_ID"`
	ClusterServer string `env:"EKS_CLUSTER_SERVER"`
	CAData        string `env:"EKS_CA_DATA"`
}

func NewEKSCredentialsFromEnv() (*EKSCredentials, error) {
	var res EKSCredentials = EKSCredentials{}

	if err := envdecode.StrictDecode(&res); err != nil {
		return nil, fmt.Errorf("Failed to decode environment variable conf: %s", err)
	}

	return &res, nil
}

func (a *EKSCredentials) GetClientSet() (*kubernetes.Clientset, error) {
	apiConfig := &api.Config{}

	clusterMap := make(map[string]*api.Cluster)

	clusterMap[a.ClusterID] = &api.Cluster{
		Server:                   a.ClusterServer,
		CertificateAuthorityData: []byte(a.CAData),
	}

	tok, err := a.GetBearerToken()

	if err != nil {
		return nil, err
	}

	authInfoMap := make(map[string]*api.AuthInfo)

	// We add the AWS bearer token to the auth info. We are not concerned about token expiration
	// as the node drain timeout should be below 15 minutes. We are not concerned about AWS rate limits
	// as we only construct the clientset once.
	authInfoMap[a.ClusterID] = &api.AuthInfo{
		Token: tok,
	}

	contextMap := make(map[string]*api.Context)

	contextMap[a.ClusterID] = &api.Context{
		Cluster:  a.ClusterID,
		AuthInfo: a.ClusterID,
	}

	apiConfig.Clusters = clusterMap
	apiConfig.AuthInfos = authInfoMap
	apiConfig.Contexts = contextMap
	apiConfig.CurrentContext = a.ClusterID

	cmdConf := clientcmd.NewDefaultClientConfig(*apiConfig, nil)

	restConf, err := cmdConf.ClientConfig()

	if err != nil {
		return nil, err
	}

	rest.SetKubernetesDefaults(restConf)

	return kubernetes.NewForConfig(restConf)
}

// GetBearerToken retrieves a bearer token for an AWS account
func (a *EKSCredentials) GetBearerToken() (string, error) {
	generator, err := token.NewGenerator(false, false)

	if err != nil {
		return "", err
	}

	sess, err := a.GetSession()

	if err != nil {
		return "", err
	}

	tok, err := generator.GetWithOptions(&token.GetTokenOptions{
		Session:   sess,
		ClusterID: a.ClusterID,
	})

	if err != nil {
		return "", err
	}

	return tok.Token, nil
}

func (a *EKSCredentials) GetSession() (*session.Session, error) {
	awsConf := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			string(a.AccessKeyID),
			string(a.SecretKey),
			"",
		),
	}

	awsConf.Region = &a.Region

	return session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *awsConf,
	})
}
