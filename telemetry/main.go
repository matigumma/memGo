package telemetry

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/posthog/posthog-go"
)

// AnonymousTelemetry - Corresponds to the Python AnonymousTelemetry class
type AnonymousTelemetry struct {
	client    posthog.Client
	userID    string
	setupCfg  func() error  // Function to call setupConfig
	getUserID func() string // Function to call getUserID
}

// NewAnonymousTelemetry initializes AnonymousTelemetry
func NewAnonymousTelemetry(projectAPIKey, host string, setupCfg func() error, getUserID func() string) (*AnonymousTelemetry, error) {
	// config := posthog.Config{Endpoint: host}
	// client := posthog.New(projectAPIKey)
	client, _ := posthog.NewWithConfig(
		// os.Getenv("POSTHOG_API_KEY"),
		projectAPIKey,
		posthog.Config{
			Endpoint: "https://us.i.posthog.com",
		},
	)

	// Ensure user_id is generated
	if err := setupCfg(); err != nil {
		return nil, fmt.Errorf("failed to run setup config: %w", err)
	}
	userID := getUserID()

	return &AnonymousTelemetry{
		client:    client,
		userID:    userID,
		setupCfg:  setupCfg,
		getUserID: getUserID,
	}, nil
}

// CaptureEvent - Corresponds to the Python capture_event method
func (t *AnonymousTelemetry) CaptureEvent(eventName string, properties map[string]interface{}) {
	if properties == nil {
		properties = make(map[string]interface{})
	}
	// Add system properties
	sysProps := t.getSystemProperties()
	for k, v := range sysProps {
		properties[k] = v
	}

	err := t.client.Enqueue(posthog.Capture{
		DistinctId: t.userID,
		Event:      eventName,
		Properties: properties,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error capturing event '%s': %v\n", eventName, err)
	}
}

// IdentifyUser - Corresponds to the Python identify_user method
func (t *AnonymousTelemetry) IdentifyUser(userID string, properties map[string]interface{}) {
	if properties == nil {
		properties = make(map[string]interface{})
	}
	err := t.client.Enqueue(posthog.Identify{
		DistinctId: userID,
		Properties: properties,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error identifying user '%s': %v\n", userID, err)
	}
}

// Close - Corresponds to the Python close method
func (t *AnonymousTelemetry) Close() {
	if t.client != nil {
		err := t.client.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error shutting down Posthog client: %v\n", err)
		}
	}
}

func (t *AnonymousTelemetry) getSystemProperties() map[string]interface{} {
	return map[string]interface{}{
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"num_cpu":    runtime.NumCPU(),
		// Add more if needed, Go doesn't have direct equivalents for all Python's platform info
		// You might need to use external commands for some (less recommended for simplicity)
		"os_version": t.getOSVersion(),
	}
}

func (t *AnonymousTelemetry) getOSVersion() string {
	var version string
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("uname", "-r")
		out, _ := cmd.CombinedOutput()
		version = strings.TrimSpace(string(out))
	case "darwin":
		cmd := exec.Command("sw_vers", "-productVersion")
		out, _ := cmd.CombinedOutput()
		version = strings.TrimSpace(string(out))
	// Add cases for "windows" if needed
	default:
		version = runtime.GOOS // Fallback to OS name
	}
	return version
}

// --- Global Telemetry Instance ---

var telemetryInstance *AnonymousTelemetry

// InitializeTelemetry initializes the global telemetry instance
func InitializeTelemetry(projectAPIKey, host string, setupCfg func() error, getUserID func() string) error {
	var err error
	telemetryInstance, err = NewAnonymousTelemetry("phc_eCRS68Q2koejazio0Umv93pwmGfwCH4uCa0dh1brRsI", host, setupCfg, getUserID)
	return err
}

// CloseTelemetry shuts down the global telemetry instance
func CloseTelemetry() {
	if telemetryInstance != nil {
		telemetryInstance.Close()
	}
}

// CaptureEvent - Global function to capture events
func CaptureEvent(eventName string, memoryInstance interface{}, additionalData map[string]interface{}) {
	if telemetryInstance == nil {
		fmt.Println("Telemetry not initialized")
		return
	}

	eventData := map[string]interface{}{
		// "collection":    memoryInstance.collectionName, // Accessing fields of an interface requires type assertion
		// "vector_size":   memoryInstance.embeddingModel.config.embedding_dims,
		"history_store": "sqlite",
		// "vector_store":    fmt.Sprintf("%T", memoryInstance.vectorStore),
		// "llm":             fmt.Sprintf("%T", memoryInstance.llm),
		// "embedding_model": fmt.Sprintf("%T", memoryInstance.embeddingModel),
		"function": fmt.Sprintf("%T", memoryInstance),
	}
	if additionalData != nil {
		for k, v := range additionalData {
			eventData[k] = v
		}
	}

	telemetryInstance.CaptureEvent(eventName, eventData)
}

// CaptureClientEvent - Global function to capture client events
func CaptureClientEvent(eventName string, instance interface{}, additionalData map[string]interface{}) {
	if telemetryInstance == nil {
		fmt.Println("Telemetry not initialized")
		return
	}

	eventData := map[string]interface{}{
		"function": fmt.Sprintf("%T", instance),
	}
	if additionalData != nil {
		for k, v := range additionalData {
			eventData[k] = v
		}
	}

	telemetryInstance.CaptureEvent(eventName, eventData)
}
