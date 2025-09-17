package serversideapply

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"testing"
	"time"

	catv1 "github.com/suinplayground/controller-runtime-playground/01-server-side-apply/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LoggingRoundTripper struct{ base http.RoundTripper }

func (l *LoggingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(r, true)
	fmt.Printf("=== Outgoing Request ===\n%s\n", dump)
	return l.base.RoundTrip(r)
}

func kubeconfigPath() (string, error) {
	if p := os.Getenv("KUBECONFIG"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kube", "config"), nil
}

func NewClient(t *testing.T) client.Client {
	t.Helper()

	kubeconfig, err := kubeconfigPath()
	if err != nil {
		t.Fatalf("resolve kubeconfig: %v", err)
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatalf("build config: %v", err)
	}

	cfg.ContentType = "application/json"
	cfg.AcceptContentTypes = "application/json"
	cfg.WrapTransport = func(rt http.RoundTripper) http.RoundTripper { return &LoggingRoundTripper{base: rt} }

	cfg.Timeout = 30 * time.Second

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}
	if err := catv1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	cl, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	return cl
}
