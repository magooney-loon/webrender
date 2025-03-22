package template

import "html/template"

// Base template with common structure and WebSocket manager
const BaseTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="icon" href="/static/logo.svg" type="image/svg+xml">
    <title>{{.Title}}</title>
    <!-- Tailwind CSS -->
    <script src="https://cdn.tailwindcss.com"></script>
    <!-- Inter font for Vercel-like UI -->
    <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap">
    <!-- Fira Code for monospace elements -->
    <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Fira+Code:wght@400;500&display=swap">
    <script>
        tailwind.config = {
            darkMode: 'class',
            theme: {
                extend: {
                    colors: {
                        'vercel-bg': '#000',
                        'vercel-gray': {
                            50: '#f9fafb',
                            100: '#eaebed',
                            200: '#dcdee3',
                            300: '#aaacb2',
                            400: '#888c94',
                            500: '#666a73',
                            600: '#4d5058',
                            700: '#3d4046',
                            800: '#26272b',
                            900: '#1a1b1f',
                        },
                        'vercel-accent': {
                            400: '#0070f3',
                            500: '#0761d1',
                        },
                        'vercel-error': '#ff4444',
                        'vercel-success': '#00c781',
                        'vercel-warning': '#ffaa15',
                    },
                    fontFamily: {
                        'sans': ['Inter', 'ui-sans-serif', 'system-ui', '-apple-system', 'sans-serif'],
                        'mono': ['Fira Code', 'ui-monospace', 'monospace'],
                    },
                    boxShadow: {
                        'vercel': '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
                        'vercel-lg': '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
                    },
                },
            }
        }
    </script>
    <style>
        /* Base app styles */
        body {
            background: radial-gradient(circle at center top, #111, #000);
            min-height: 100vh;
            overflow-x: hidden;
        }
        .component-container {
            transition: all 0.2s ease;
        }
        .component-container:hover {
            transform: translateY(-2px);
        }
        .progress-bar {
            transition: width 0.5s ease-in-out;
        }
        .notification {
            animation: fadeIn 0.3s ease-in-out;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(-10px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .vercel-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 1.5rem;
        }
        .vercel-card {
            background: rgba(32, 32, 36, 0.5);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(63, 63, 70, 0.5);
            border-radius: 0.5rem;
            overflow: hidden;
            transition: all 0.2s ease-in-out;
        }
        .vercel-card:hover {
            border-color: rgba(63, 63, 70, 0.8);
            transform: translateY(-2px);
            box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.3);
        }
        .vercel-input {
            background: rgba(32, 32, 36, 0.5);
            border: 1px solid rgba(63, 63, 70, 0.5);
            border-radius: 0.375rem;
            color: #fff;
            font-size: 0.875rem;
            outline: none;
            padding: 0.625rem 0.75rem;
            transition: all 0.2s ease;
        }
        .vercel-input:focus {
            border-color: #0070f3;
            box-shadow: 0 0 0 2px rgba(0, 112, 243, 0.2);
        }
        .vercel-btn {
            align-items: center;
            border-radius: 0.375rem;
            display: inline-flex;
            font-weight: 500;
            justify-content: center;
            outline: none;
            padding: 0.5rem 1rem;
            position: relative;
            transition: all 0.2s ease;
            white-space: nowrap;
        }
        .vercel-btn-primary {
            background: #0070f3;
            color: #fff;
        }
        .vercel-btn-primary:hover {
            background: #0761d1;
        }
        .vercel-btn-secondary {
            background: rgba(32, 32, 36, 0.8);
            border: 1px solid rgba(63, 63, 70, 0.5);
            color: #fff;
        }
        .vercel-btn-secondary:hover {
            border-color: rgba(63, 63, 70, 0.8);
            background: rgba(32, 32, 36, 1);
        }
        
        /* Custom styles for the page */
        {{.Styles}}
    </style>
</head>
<body class="font-sans text-white leading-relaxed m-0 p-0">
    <div id="app" class="max-w-7xl mx-auto p-5">
        {{.Content}}
    </div>

    <!-- WebRender Core -->
    <script>
    {{.ClientJS}}
    </script>
    
    <!-- Initialize WebSocket -->
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            // Initialize WebSocket with auto-reconnect
            const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = wsProtocol + '//' + window.location.host + '/ws';
            WSManager.init(wsUrl);
            
            // Listen for state updates
            WSManager.on('state_update', function(data) {
                // Find component by ID
                const component = document.getElementById(data.component_id);
                if (!component) {
                    console.warn('Component not found:', data.component_id);
                    return;
                }
                
                // Update component state
                try {
                    let state = JSON.parse(component.getAttribute('data-state') || '{}');
                    
                    if (data.type === 'delete') {
                        delete state[data.key];
                    } else {
                        state[data.key] = data.value;
                    }
                    
                    // Update the data attribute
                    component.setAttribute('data-state', JSON.stringify(state));
                    
                    // Update any bound elements
                    const boundElements = component.querySelectorAll('[data-bind="' + data.key + '"]');
                    boundElements.forEach(el => {
                        el.textContent = data.value;
                    });
                    
                    // Dispatch a custom event for the component
                    component.dispatchEvent(new CustomEvent('state-changed', {
                        detail: {
                            key: data.key,
                            value: data.value,
                            type: data.type
                        }
                    }));
                } catch (error) {
                    console.error('Error updating component state:', error);
                }
            });
        });
    </script>

    <!-- Custom scripts for the page -->
    <script>{{.Scripts}}</script>
</body>
</html>
`

// PageData contains data for rendering a complete page
type PageData struct {
	Title    string
	Content  template.HTML
	Styles   template.CSS
	Scripts  template.JS
	ClientJS template.JS
}

// GetBaseTemplate returns a parsed base template
func GetBaseTemplate() *template.Template {
	return template.Must(template.New("base").Parse(BaseTemplate))
}
