package client

// Config holds the client daemon configuration, populated from CLI flags.
type Config struct {
	Endpoint       string
	TokenPath      string
	KubeconfigPath string
	KubeAPIServer  string
}
