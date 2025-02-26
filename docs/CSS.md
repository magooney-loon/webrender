# CSS Architecture

Our dashboard uses a data-attribute utility approach for styling, providing a flexible and maintainable CSS architecture.

## Data-Attribute Utilities

Instead of traditional class-based utilities, we use data attributes for better semantics and scoping:

```html
<!-- Example component using data attributes -->
<div data-layout="flex" data-gap="4" data-p="4">
  <div data-bg="surface" data-rounded="lg" data-shadow="md">
    <h2 data-text="xl" data-font="bold">Dashboard</h2>
  </div>
</div>
```

## Core Utilities

### Layout
```css
[data-layout="flex"] { display: flex; }
[data-layout="grid"] { display: grid; }
[data-layout~="center"] { align-items: center; justify-content: center; }
[data-layout~="col"] { flex-direction: column; }

/* Gap utilities */
[data-gap="1"] { gap: 0.25rem; }
[data-gap="2"] { gap: 0.5rem; }
[data-gap="4"] { gap: 1rem; }
[data-gap="8"] { gap: 2rem; }
```

### Spacing
```css
/* Padding */
[data-p="1"] { padding: 0.25rem; }
[data-p="2"] { padding: 0.5rem; }
[data-p="4"] { padding: 1rem; }
[data-p="8"] { padding: 2rem; }

/* Margin */
[data-m="1"] { margin: 0.25rem; }
[data-m="2"] { margin: 0.5rem; }
[data-m="4"] { margin: 1rem; }
[data-m="8"] { margin: 2rem; }
```

### Typography
```css
[data-text="sm"] { font-size: 0.875rem; line-height: 1.25rem; }
[data-text="base"] { font-size: 1rem; line-height: 1.5rem; }
[data-text="lg"] { font-size: 1.125rem; line-height: 1.75rem; }
[data-text="xl"] { font-size: 1.25rem; line-height: 1.75rem; }

[data-font="normal"] { font-weight: 400; }
[data-font="medium"] { font-weight: 500; }
[data-font="bold"] { font-weight: 700; }
```

### Colors & Themes
```css
/* Auto-switching light/dark themes */
[data-bg="surface"] {
  background-color: var(--surface);
  color: var(--on-surface);
}

[data-bg="primary"] {
  background-color: var(--primary);
  color: var(--on-primary);
}

/* CSS variables for theming */
:root {
  --primary: #2563eb;
  --on-primary: #ffffff;
  --surface: #ffffff;
  --on-surface: #1f2937;
}

[data-theme="dark"] {
  --primary: #3b82f6;
  --on-primary: #ffffff;
  --surface: #1f2937;
  --on-surface: #f3f4f6;
}
```

## Component Examples

### Card Component
```html
<div data-component="card">
  <div data-layout="flex col" data-gap="2" data-p="4">
    <h3 data-text="lg" data-font="bold">Card Title</h3>
    <p data-text="base">Card content goes here.</p>
  </div>
</div>
```

### Dashboard Grid
```html
<div data-layout="grid" data-cols="3" data-gap="4">
  <div data-component="stat-card" data-bg="surface">
    <span data-text="sm" data-color="muted">Total Users</span>
    <span data-text="2xl" data-font="bold">1,234</span>
  </div>
</div>
```

## Benefits

1. **Semantic HTML**: Data attributes keep HTML semantic while providing styling hooks
2. **Scoped Styles**: Attributes provide natural scoping without deep selectors
3. **Theme Support**: Easy theme switching via CSS variables
4. **Performance**: Flat selectors are fast and efficient
5. **Maintainability**: Clear separation of styling from structure
6. **Developer Experience**: Intuitive and self-documenting

## Best Practices

1. **Composition Over Inheritance**
   ```html
   <!-- Good -->
   <div data-layout="flex" data-gap="4">
   
   <!-- Avoid -->
   <div class="flex-container large-gap">
   ```

2. **Single Responsibility**
   ```html
   <!-- Good -->
   <div data-layout="flex" data-p="4" data-bg="surface">
   
   <!-- Avoid -->
   <div data-style="flex-padded-surface">
   ```

3. **Responsive Design**
   ```css
   @media (min-width: 768px) {
     [data-cols="3"] {
       grid-template-columns: repeat(3, 1fr);
     }
   }
   ```

4. **State Management**
   ```css
   [data-state="loading"] {
     opacity: 0.7;
     pointer-events: none;
   }
   
   [data-state="error"] {
     border-color: var(--error);
   }
   ``` 