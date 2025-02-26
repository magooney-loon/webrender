# JavaScript Architecture

Our JavaScript architecture follows a modular pattern that integrates tightly with our data-attribute CSS approach. The codebase is organized into core modules that handle specific functionality.

## Core Modules

```js
// core.js - Core functionality and utilities
export const core = {
    theme: {
        toggle: () => document.documentElement.dataset.theme = 
            document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark',
        get: () => document.documentElement.dataset.theme,
        set: (theme) => document.documentElement.dataset.theme = theme
    },
    
    state: {
        set: (el, state) => el.dataset.state = state,
        remove: (el) => delete el.dataset.state
    },
    
    // Event delegation helper
    on: (selector, event, handler) => {
        document.addEventListener(event, e => {
            if (e.target.matches(selector)) handler(e)
        })
    }
}

// dashboard.js - Dashboard specific functionality
export class Dashboard {
    constructor() {
        this.stats = document.querySelector('[data-component="stats"]')
        this.charts = document.querySelector('[data-component="charts"]')
        this.setupRealtime()
    }

    async updateStats() {
        core.state.set(this.stats, 'loading')
        try {
            const data = await this.fetchStats()
            this.renderStats(data)
            core.state.remove(this.stats)
        } catch (err) {
            core.state.set(this.stats, 'error')
        }
    }
}

// settings.js - User settings and preferences
export class Settings {
    constructor() {
        this.bindThemeToggle()
        this.bindNotifications()
    }

    bindThemeToggle() {
        core.on('[data-action="toggle-theme"]', 'click', () => core.theme.toggle())
    }
}
```

## Data-Attribute Integration

We use data attributes for both styling and JavaScript behavior:

```html
<!-- Component with both styling and behavior attributes -->
<div data-component="chart" 
     data-layout="flex" 
     data-gap="4"
     data-update-interval="5000">
    <div data-chart="line"
         data-height="300"
         data-animate="true"></div>
</div>
```

### Behavior Attributes

```js
// Attribute-based behavior initialization
class ChartComponent {
    constructor(el) {
        this.el = el
        this.type = el.dataset.chart
        this.height = parseInt(el.dataset.height)
        this.shouldAnimate = el.dataset.animate === 'true'
        this.updateInterval = parseInt(el.dataset.updateInterval)
        
        this.init()
    }
    
    init() {
        // Initialize chart based on data attributes
    }
}

// Auto-initialize components
document.querySelectorAll('[data-component="chart"]')
    .forEach(el => new ChartComponent(el))
```

## State Management

We use data-state attributes for managing component states:

```js
// State management in components
class StatCard {
    async refresh() {
        this.el.dataset.state = 'loading'
        try {
            const data = await this.fetchData()
            this.render(data)
            delete this.el.dataset.state
        } catch (err) {
            this.el.dataset.state = 'error'
        }
    }
}
```

```css
/* Corresponding CSS */
[data-state="loading"] {
    opacity: 0.7;
    pointer-events: none;
}

[data-state="error"] {
    border-color: var(--error);
    background: var(--error-bg);
}
```

## Event Handling

We use data-action attributes for declarative event binding:

```html
<button data-action="save-settings"
        data-layout="flex"
        data-gap="2">
    <span data-icon="save"></span>
    Save Changes
</button>
```

```js
// Event delegation based on data-action
core.on('[data-action]', 'click', e => {
    const action = e.target.dataset.action
    switch (action) {
        case 'save-settings':
            settings.save()
            break
        case 'toggle-theme':
            core.theme.toggle()
            break
    }
})
```

## Real-time Updates

WebSocket integration with data attributes:

```js
class RealtimeComponent {
    constructor(el) {
        this.el = el
        this.channel = el.dataset.channel
        this.updateInterval = parseInt(el.dataset.updateInterval)
        
        this.connect()
    }
    
    connect() {
        this.ws = new WebSocket(WS_URL)
        this.ws.onmessage = (msg) => this.handleUpdate(msg)
        this.ws.onclose = () => {
            this.el.dataset.state = 'disconnected'
            setTimeout(() => this.connect(), 5000)
        }
    }
}
```

## Module Organization

Our JavaScript is organized into focused modules:

```
web/static/js/
├── core.js       # Core utilities and helpers
├── dashboard.js  # Dashboard-specific functionality
├── settings.js   # User settings management
├── telegram.js   # Telegram integration
├── navigation.js # Navigation and routing
└── login.js      # Authentication handling
```

## Best Practices

1. **Data Attribute Consistency**
   ```js
   // Good - Using data attributes for state
   element.dataset.state = 'loading'
   
   // Avoid - Using classes for state
   element.classList.add('loading')
   ```

2. **Event Delegation**
   ```js
   // Good - Single event listener with delegation
   core.on('[data-action]', 'click', handler)
   
   // Avoid - Multiple direct listeners
   elements.forEach(el => el.addEventListener('click', handler))
   ```

3. **State Management**
   ```js
   // Good - Clear state transitions
   async function updateComponent() {
       el.dataset.state = 'loading'
       try {
           await update()
           delete el.dataset.state
       } catch {
           el.dataset.state = 'error'
       }
   }
   ```

4. **Component Initialization**
   ```js
   // Good - Declarative component creation
   document.querySelectorAll('[data-component]')
       .forEach(el => {
           const type = el.dataset.component
           new Components[type](el)
       })
   ```

## Integration Examples

### Dashboard Stats Card
```html
<div data-component="stat-card"
     data-layout="flex col"
     data-gap="2"
     data-update-interval="30000">
    <h3 data-text="sm" data-color="muted">Active Users</h3>
    <p data-text="2xl" data-font="bold">0</p>
</div>
```

```js
class StatCard {
    constructor(el) {
        this.el = el
        this.interval = setInterval(
            () => this.update(),
            parseInt(el.dataset.updateInterval)
        )
    }
    
    async update() {
        this.el.dataset.state = 'loading'
        try {
            const data = await fetch('/api/stats')
            this.el.querySelector('p').textContent = data.value
            delete this.el.dataset.state
        } catch {
            this.el.dataset.state = 'error'
        }
    }
}
``` 