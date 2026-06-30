package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shivamshashank/CloudInferOps/internal/utils"
	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Manage and discover AI models",
	Long:  `Retrieve available model catalogs, pull local/remote LLMs, and query deployed inference endpoints.`,
}

type ModelInfo struct {
	Name     string
	Provider string
	Status   string
	Location string
}

var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available or deployed models",
	Long:  `Discovers active models via the inference gateway, local Ollama daemon, or falls back to local configuration defaults.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%sDiscovering active model registries...\n\n", utils.PrefixInfo)

		var models []ModelInfo
		discovered := false

		// 1. Try querying the deployed CloudInferOps Gateway (defaults)
		gatewayUrls := []string{"http://localhost:8000/models", "http://cloudinferops.local/models"}
		client := http.Client{Timeout: 500 * time.Millisecond}

		for _, url := range gatewayUrls {
			resp, err := client.Get(url)
			if err == nil && resp.StatusCode == http.StatusOK {
				var gatewayModels []struct {
					ID     string `json:"id"`
					Object string `json:"object"`
				}
				decodeErr := json.NewDecoder(resp.Body).Decode(&gatewayModels)
				_ = resp.Body.Close()
				if decodeErr == nil && len(gatewayModels) > 0 {
					for _, m := range gatewayModels {
						models = append(models, ModelInfo{
							Name:     m.ID,
							Provider: "gateway",
							Status:   "🟢 Active",
							Location: "Cloud Gateway",
						})
					}
					discovered = true
					break
				}
			}
		}

		// 2. If not discovered via gateway, try local Ollama daemon
		if !discovered {
			resp, err := client.Get("http://localhost:11434/api/tags")
			if err == nil && resp.StatusCode == http.StatusOK {
				var ollamaTags struct {
					Models []struct {
						Name string `json:"name"`
					} `json:"models"`
				}
				decodeErr := json.NewDecoder(resp.Body).Decode(&ollamaTags)
				_ = resp.Body.Close()
				if decodeErr == nil && len(ollamaTags.Models) > 0 {
					for _, m := range ollamaTags.Models {
						models = append(models, ModelInfo{
							Name:     m.Name,
							Provider: "ollama",
							Status:   "🟢 Downloaded",
							Location: "Local Daemon",
						})
					}
					discovered = true
				}
			}
		}

		// 3. Fallback to static catalog if nothing is active
		if !discovered {
			fmt.Printf("%sNo active gateway or local Ollama daemon detected. Showing static default catalog.\n\n", utils.PrefixWarn)
			models = []ModelInfo{
				{Name: "llama3", Provider: "ollama", Status: "⚪ Available", Location: "Local Registry"},
				{Name: "mistral", Provider: "ollama", Status: "⚪ Available", Location: "Local Registry"},
				{Name: "phi3", Provider: "ollama", Status: "⚪ Available", Location: "Local Registry"},
				{Name: "llama3-70b", Provider: "vllm", Status: "⚪ Available", Location: "Remote Registry"},
				{Name: "mixtral-8x7b", Provider: "vllm", Status: "⚪ Available", Location: "Remote Registry"},
			}
		}

		// Format and print the table output
		fmt.Printf("%s%-20s %-12s %-15s %-20s%s\n", utils.ColorBold, "MODEL", "PROVIDER", "STATUS", "LOCATION", utils.ColorReset)
		fmt.Println("----------------------------------------------------------------------")
		for _, m := range models {
			fmt.Printf("%-20s %-12s %-15s %-20s\n", m.Name, m.Provider, m.Status, m.Location)
		}
		fmt.Println()

		return nil
	},
}

var modelsPullCmd = &cobra.Command{
	Use:   "pull [model]",
	Short: "Pull a model to local model registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelName := args[0]
		fmt.Printf("%sPulling model %s%s%s...\n", utils.PrefixInfo, utils.ColorBold, modelName, utils.ColorReset)
		fmt.Println(utils.PrefixInfo + "Downloading model manifest and weights (Logic to be implemented in Phase 4)...")
		return nil
	},
}

func init() {
	modelsCmd.AddCommand(modelsListCmd)
	modelsCmd.AddCommand(modelsPullCmd)
	RootCmd.AddCommand(modelsCmd)
}
