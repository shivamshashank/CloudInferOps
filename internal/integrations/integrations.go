package integrations

import "fmt"

func ConfigureSlack(webhookURL string) {
	fmt.Printf("[Integrations] Configuring Slack with webhook URL: %s\n", webhookURL)
}

func ConfigurePagerDuty(routingKey string) {
	fmt.Printf("[Integrations] Configuring PagerDuty with routing key: %s\n", routingKey)
}
