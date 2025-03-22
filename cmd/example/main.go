package main

import (
	"fmt"
	"html/template"
	"log"

	"github.com/magooney-loon/webrender/pkg"
	"github.com/magooney-loon/webrender/pkg/components/example"
	"github.com/magooney-loon/webrender/pkg/components/testcomponent"
)

func main() {
	// Create a new WebRender instance with default configuration
	config := pkg.DefaultConfig()

	// Initialize WebRender
	webRender, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to initialize WebRender: %v", err)
	}

	// Register the counter component manually
	counter := example.NewCounter("counter-1")
	if err := webRender.RegisterComponent(counter); err != nil {
		log.Printf("Error registering counter component: %v", err)
	} else {
		log.Println("Counter component registered successfully with ID:", counter.ID)
	}

	// Home page with counter example using the simplified ComponentRoute API
	webRender.ComponentRoute("/", "WebRender Example", "counter-1",
		map[string]interface{}{"title": "Click Counter"},
		func() template.CSS { return template.CSS(example.GetStyles()) },
		func() template.JS { return template.JS(example.GetScripts()) },
	)

	// Alternative using RouteWithTemplate for more control
	webRender.RouteWithTemplate("/alt", "Alternative Example", func() (template.HTML, error) {
		// Render the counter component
		counterHTML, err := webRender.RenderComponent("counter-1", map[string]interface{}{"title": "Custom Counter"})
		if err != nil {
			return "", err
		}
		return template.HTML(counterHTML), nil
	},
		func() template.CSS { return template.CSS(example.GetStyles()) },
		func() template.JS { return template.JS(example.GetScripts()) })

	// Create and register the component first
	testcomponentComp := testcomponent.NewTestComponent("testcomponent-id")
	if err := webRender.RegisterComponent(testcomponentComp); err != nil {
		log.Printf("Error registering testcomponent component: %v", err)
	}

	webRender.ComponentRoute("/testcomponent", "TestComponent Example", "testcomponent-id",
		map[string]interface{}{
			"title":       "TestComponent Component",
			"description": "A custom TestComponent component",
		},
		func() template.CSS { return template.CSS(testcomponent.GetStyles()) },
		func() template.JS { return template.JS(testcomponent.GetScripts()) },
	)

	fmt.Println("To create new components, run the component generator: go run cmd/component/create.go")
	log.Fatal(webRender.Start(":8080"))
}
