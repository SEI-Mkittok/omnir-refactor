```markdown
# Design System Specification: The Rational Archive

## 1. Overview & Creative North Star
This design system is built upon the **"Rational Archive"** North Star. We are moving away from the ephemeral "vibe-coded" trends of the moment to embrace a permanent, utility-first aesthetic. It is a system designed for high-density information, where clarity is the highest form of luxury.

Unlike standard "Modern SaaS" templates that rely on heavy drop shadows and rounded bubbles, this system finds its identity in **Structured Intellectualism**. We use intentional asymmetry, disciplined typography, and a "Layered Document" metaphor. The UI should feel like a high-end architectural blueprint or a meticulously organized digital ledger—functional, quiet, and profoundly authoritative.

---

## 2. Colors & Surface Architecture
Our palette has been recalibrated to a sophisticated Slate and Indigo core (`#515F74`). This desaturated primary tone provides a professional "ink on paper" feel rather than a digital neon glow.

### The "No-Line" Rule
Standard 1px borders are strictly prohibited for primary sectioning. Global boundaries must be defined through **Background Color Shifts**. 
*   **The Technique:** Instead of a border, place a `surface-container-low` (`#f0f4f7`) section directly against the `surface` (`#f7f9fb`) background. 
*   **Result:** This creates a "flush" architectural look that feels integrated into the interface rather than floating on top of it.

### Surface Hierarchy & Nesting
Treat the UI as a physical stack of premium cardstock. Hierarchy is achieved through "Tonal Nesting":
1.  **Base Layer:** `surface` (`#f7f9fb`)
2.  **Structural Zones:** `surface-container` (`#e8eff3`)
3.  **Active Utility Zones:** `surface-container-high` (`#e1e9ee`)
4.  **Interactive Components:** `surface-container-lowest` (`#ffffff`)

### The Glass & Texture Rule
To prevent the "Rational" aesthetic from feeling cold, use **Glassmorphism** for floating overlays (modals, popovers). Use `surface_container_lowest` at 80% opacity with a `20px` backdrop blur. This allows the structural geometry of the underlying data to bleed through, maintaining a sense of place.

---

## 3. Typography: Functional Authority
The system utilizes **Inter**—a typeface designed for numerical and functional precision.

*   **Display & Headline:** Use `display-lg` (3.5rem) and `headline-lg` (2rem) with tight letter-spacing (-0.02em). These are your "Editorial Anchors." Position them with wide asymmetric margins to create a high-end magazine feel.
*   **Body:** `body-md` (0.875rem) is the workhorse. It must remain highly legible against the subtle surface shifts.
*   **Labels:** `label-sm` (0.6875rem) should be used for metadata and utility tags, often in `on_surface_variant` (`#566166`) to keep the visual noise low.

**Typography as Brand:** The identity is found in the *scale* difference. Pair a large `display-md` title with a very small `label-md` subtitle immediately adjacent to it. This "High-Low" pairing signals a sophisticated, data-driven environment.

---

## 4. Elevation & Depth: Tonal Layering
We reject the standard "shadow-heavy" UI. Depth is an environmental property, not a decorative one.

*   **The Layering Principle:** A `surface-container-lowest` card placed on a `surface-container-low` background creates a natural visual "lift" without a single pixel of shadow. This is the preferred method for cards and containers.
*   **Ambient Shadows:** If a floating element (like a context menu) requires a shadow, use a hyper-diffused style: `0 12px 40px rgba(42, 52, 57, 0.06)`. The shadow color must be derived from `on_surface` to look like natural light occlusion.
*   **The Ghost Border:** If accessibility requires a stroke, use `outline-variant` (`#a9b4b9`) at 15% opacity. It should be felt, not seen.

---

## 5. Components & Primitive Logic

### Buttons: The Utility Action
*   **Primary:** Background `primary` (`#515f74`), Text `on_primary` (`#f6f7ff`). Use `md` (0.375rem) corner radius.
*   **Secondary:** Background `primary_container` (`#d5e3fc`), Text `on_primary_container`.
*   **Tertiary/Ghost:** No background. Text `primary`. For use in dense data tables where visual clutter must be minimized.

### Input Fields: Rational Entries
*   **Default State:** Background `surface_container_lowest`, 1px Ghost Border (15% opacity).
*   **Focus State:** Border opacity increases to 100% using `primary` color. No "glow" or outer rings.
*   **Error State:** Text and underline using `error` (`#9f403d`).

### Cards & Lists: The Separation Rule
**Dividers are forbidden.** To separate list items or card sections:
1.  Use **Vertical White Space**: Apply `spacing-4` (0.9rem) or `spacing-5` (1.1rem) between items.
2.  Use **Subtle Fills**: Apply `surface_container_low` on hover to indicate interactivity.

### Data Tables (The Archive Component)
Tables should use `body-sm` for high density. Headers should use `label-md` in all-caps with `0.05em` letter spacing. This mimics a professional archival system.

---

## 6. Do's and Don'ts

### Do:
*   **Use Asymmetric Padding:** Try a `spacing-16` (3.5rem) left margin and a `spacing-8` (1.75rem) right margin for hero sections to create an editorial "Rational" look.
*   **Embrace the "Empty":** Allow `surface` colors to breathe. White space in this system is a functional tool to separate thoughts.
*   **Respect the Grid:** While elements can be asymmetric, they must always align to the `spacing-px` or `spacing-1` grid.

### Don't:
*   **Don't use 100% Black:** Always use `on_surface` (`#2a3439`) for text to maintain the "muted/professional" indigo-slate tone.
*   **Don't use "Vibe" Gradients:** Avoid colorful, mesh gradients. Only use functional gradients (e.g., `primary` to `primary_dim`) to add subtle depth to buttons.
*   **Don't use Rounded Corners > 0.75rem:** We are not building a consumer social app. Keep corners `sm` (0.125rem) to `md` (0.375rem) to maintain a structural, archival feel.