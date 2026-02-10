package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rophy/kube-imds/internal/client"
)

var Version = "dev"

func main() {
	endpoint := flag.String("endpoint", "", "kube-imds server URL (e.g. http://kube-imds:8080)")
	tokenPath := flag.String("token-path", "/var/run/kube-imds/token", "path to write the token file")
	kubeconfigPath := flag.String("kubeconfig-path", "/var/run/kube-imds/kubeconfig", "path to write the kubeconfig file")
	kubeAPIServer := flag.String("kube-apiserver", "", "Kubernetes API server URL for kubeconfig generation")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	if *endpoint == "" {
		log.Fatal("--endpoint is required")
	}
	if *kubeAPIServer == "" {
		log.Fatal("--kube-apiserver is required")
	}

	log.Printf("kube-imds-client version %s", Version)

	cfg := &client.Config{
		Endpoint:       *endpoint,
		TokenPath:      *tokenPath,
		KubeconfigPath: *kubeconfigPath,
		KubeAPIServer:  *kubeAPIServer,
	}

	if err := client.WriteKubeconfig(cfg.KubeconfigPath, cfg.KubeAPIServer, cfg.TokenPath); err != nil {
		log.Fatalf("failed to write kubeconfig: %v", err)
	}
	log.Printf("wrote kubeconfig to %s", cfg.KubeconfigPath)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client.Run(ctx, cfg)

	log.Println("shutting down")
}
