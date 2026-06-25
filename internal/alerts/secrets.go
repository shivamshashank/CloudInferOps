package alerts

import (
	"fmt"

	"github.com/shivamshashank/CloudInferOps/internal/utils"
)

// CreateSecret creates or updates a Kubernetes Secret inside the cluster in an idempotent manner
func CreateSecret(name, key, value, namespace string) error {
	// Constructing the idempotent apply pipeline
	cmdStr := fmt.Sprintf("kubectl create secret generic %s --from-literal=%s='%s' -n %s --dry-run=client -o yaml | kubectl apply -f -", name, key, value, namespace)

	_, stderr, err := utils.ExecCommand("", "sh", "-c", cmdStr)
	if err != nil {
		return fmt.Errorf("failed to provision Kubernetes secret '%s': %w (stderr: %s)", name, err, stderr)
	}

	return nil
}
