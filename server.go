package main

import (
	"crypto/rand"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

var memConsumed []byte

// TemplateData struct for HTML template
type TemplateData struct {
	Timestamp  string
	IP         string
	Pod        string
	Node       string
	Namespace  string
	PageHeader string
}

func main() {
	// Retrieve environment variables that may be set within the container.
	ip := os.Getenv("IP")               // IP address of the container
	pod := os.Getenv("POD")             // Name of the pod where the container is running
	node := os.Getenv("NODE")           // Name of the node where the pod is scheduled
	namespace := os.Getenv("NAMESPACE") // Namespace where the pod is deployed

	// Create a new Echo instance to handle HTTP requests.
	e := echo.New()

	// Define a template with the specified HTML content.
	tpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>{{.PageHeader}}</title>
		<style>
			body {
				background-color: #22c55e; /* Green background */
				font-family: Arial, sans-serif;
			}
			.container {
				margin: 20px;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>{{.PageHeader}}</h1>
			<p>Timestamp: {{.Timestamp}}</p>
			<p>IP: {{.IP}}</p>
			<p>Pod: {{.Pod}}</p>
			<p>Node: {{.Node}}</p>
			<p>Namespace: {{.Namespace}}</p>
		</div>
	</body>
	</html>
	`

	// Parse the HTML template.
	htmlTemplate := template.Must(template.New("htmlTemplate").Parse(tpl))

	// Define a route for handling incoming HTTP GET requests on all paths.
	e.GET("/*", func(c echo.Context) error {
		// Create a TemplateData instance with the information about the container and its environment.
		data := TemplateData{
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			IP:         ip,
			Pod:        pod,
			Node:       node,
			Namespace:  namespace,
			PageHeader: "Container Information V1",
		}

		// Render the HTML template with the provided data.
		return htmlTemplate.Execute(c.Response().Writer, data)
	})

	// Define a route for the "stress" endpoint.
	e.GET("/stress", func(c echo.Context) error {
		// Check if memory has already been consumed.
		if memConsumed != nil {
			return c.String(http.StatusOK, "RAM stress test is already running.")
		}

		// Define the target memory size in bytes (1 GB).
		const targetSize = 1024 * 1024 * 1024 // 1 GB

		// Allocate memory for the entire 1 GB.
		memConsumed = make([]byte, targetSize)

		// Fill the allocated memory with random data.
		rand.Read(memConsumed)

		// Return a response indicating that RAM has been consumed.
		return c.String(http.StatusOK, "RAM stress test initiated on Pod: "+pod)
	})

	// Define a route for the "kill" endpoint.
	e.GET("/kill", func(c echo.Context) error {
		// Return a message indicating that the server will be terminated, including the pod name.
		response := "Server in Pod: " + pod + " will be terminated."

		// Start a goroutine to allow the response to be sent before termination.
		go func() {
			time.Sleep(1 * time.Second) // Give some time for the response to be sent
			os.Exit(0)                  // Terminate the server after a delay
		}()

		return c.String(http.StatusOK, response)
	})

	// Start the Echo web server on port 8080 and handle incoming HTTP requests.
	e.Logger.Fatal(e.Start(":8080"))
}
