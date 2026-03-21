# Design System Specification: The Rational Archive

## 1. Overview & Creative North Star
This design system is anchored by the **"The Rational Archive"**—a creative North Star that treats digital interfaces with the reverence of high-end editorial archives. While inspired by the utility-first nature of Notion and the precision of Linear, this system transcends "standard" SaaS patterns through intentional asymmetry, extreme typographic hierarchy, and architectural breathing room.

The goal is to create a "Quiet Authority." We achieve this by favoring **Tonal Depth** over structural lines. By breaking the rigid 1:1 grid and utilizing the provided spacing scale to create intentional "pockets" of white space, we guide the user’s eye not through boxes, but through visual weight and mathematical precision.

## 2. Colors & Surface Architecture
The palette is a sophisticated interplay of cool neutrals and a singular, high-integrity Indigo (`primary: #4d44e3`). 

### The "No-Line" Rule
Standard UI relies on gray borders to define space; this design system prohibits them for primary sectioning. Boundaries must be defined through **Background Color Shifts**. To separate a sidebar from a main content area, place a `surface_container_low` (#f2f3ff) panel against a `background` (#faf8ff) canvas. 

### Surface Hierarchy & Nesting
Treat the UI as a series of physical layers—stacked sheets of fine, heavy-stock paper. Use the surface-container tiers to create nested depth:
*   **Base:** `background` (#faf8ff)
*   **Secondary Sections:** `surface_container_low` (#f2f3ff)
*   **Primary Interaction Containers (Cards/Modals):** `surface_container_lowest` (#ffffff)
*   **Utility Elements:** `surface_container_high` (#e2e7ff)

### Signature Textures
To avoid a "flat" feel without resorting to trendy glows, utilize **Tonal Gradients**. For primary CTAs or hero headers, use a subtle linear transition from `primary` (#4d44e3) to `primary_dim` (#4034d7). This creates a "latte" effect—a soft, professional polish that feels like high-quality ink on paper.

## 3. Typography
We utilize **Inter** across all scales to maintain a utilitarian, Swiss-inspired aesthetic. 

*   **The Display Scale:** Use `display-lg` (3.5rem) with `on_background` (#113069) for entry points. This creates a high-contrast editorial "moment" that commands attention.
*   **The Utility Scale:** Navigational elements and metadata should exclusively use `label-md` and `label-sm`. The smaller the text, the more generous the letter-spacing should be to maintain the "Linear-esque" utility feel.
*   **Rhythmic Hierarchy:** Ensure `headline-sm` is used to categorize content blocks, always followed by a `spacing-4` (1.4rem) gap before `body-md` text.

## 4. Elevation & Depth
In this system, "Elevation" is a measure of light and density, not shadow.

### Tonal Layering
Depth is achieved by "stacking" the surface-container tiers. Place a `surface_container_lowest` card on top of a `surface_container_low` section to create a soft, natural lift. 

### Ambient Shadows
When a floating element (like a context menu) is required, use **Ambient Shadows**. Specify a blur of 24px-48px with an opacity of 4%-6%. The shadow color must be a tinted version of `on_surface` (#113069), never pure black. This mimics natural light passing through a translucent object.

### The "Ghost Border"
If a border is required for accessibility (e.g., input fields), use a **Ghost Border**: the `outline_variant` (#98b1f2) at 20% opacity. This provides a structural hint without cluttering the visual field with hard lines.

## 5. Components

### Buttons
*   **Primary:** Background: `primary` (#4d44e3). Text: `on_primary` (#faf6ff). Border-radius: `md` (0.375rem).
*   **Secondary:** Background: `secondary_container` (#d3e4fe). Text: `on_secondary_container` (#435368).
*   **Tertiary:** No background. Text: `primary`. Used for low-priority utility actions.

### Input Fields
*   **Surface:** `surface_container_lowest` (#ffffff).
*   **Border:** Ghost Border (20% `outline_variant`). 
*   **Padding:** Use `spacing-3` (1rem) for internal horizontal padding to feel generous and modern.
*   **States:** On focus, transition the border to `primary` (#4d44e3) and the background to `surface_bright`.

### Navigation (Utility-First)
*   **Structure:** Icon-heavy, left-aligned. 
*   **Icons:** Use `outline` (#6079b7) for inactive states. 
*   **Active State:** `primary` (#4d44e3) icon with a `surface_container_high` (#e2e7ff) background pill.

### Cards & Lists
*   **Strict Rule:** No divider lines. Separate list items using `spacing-2` (0.7rem) of vertical white space or a subtle background hover state shift to `surface_container_low`.

## 6. Do's and Don'ts

### Do:
*   **Use Asymmetric Padding:** On large screens, allow for wider margins on the left to create an editorial "gutter" effect.
*   **Respect the Spacing Scale:** Stick strictly to the increments (e.g., use `spacing-12` for section breaks).
*   **Prioritize Iconography:** Use icons to reduce cognitive load in utility-dense areas, but always pair them with `label-sm` text for clarity.

### Don't:
*   **Don't use 100% Opaque Borders:** This creates a "boxed-in" feeling that contradicts the minimalist North Star.
*   **Don't use Pure Black (#000000):** All "dark" elements must use `on_surface` (#113069) or `secondary` (#506076) to maintain the cool, indigo-tinted atmosphere.
*   **Don't Over-Round:** Stick to `DEFAULT` (0.25rem) or `md` (0.375rem) for most components. Reserves `full` (9999px) only for chips and tags.