# QA Report: Saved Views UI (OMN-315)

**Issue**: OMN-333
**Feature**: Phase 8 Saved Searches + Views Frontend UI
**Branch**: `feature/OMN-315-saved-views-ui`
**Commit**: 58d03f9
**QA Date**: 2026-03-18
**QA Engineer**: Skadi

## Test Environment

- **Frontend**: Vite dev server (http://localhost:5173)
- **API**: http://localhost:8080
- **Database**: PostgreSQL with `saved_views` migration applied

## Implementation Review

### Files Changed
- `web/src/api/types.ts` - Added SavedView, ViewFilters, and CRUD request types
- `web/src/api/views.ts` - Views API module (list, get, create, update, delete, pin)
- `web/src/hooks/useViews.ts` - React Query hooks with cache invalidation
- `web/src/stores/viewFilters.ts` - Zustand store with localStorage persistence
- `web/src/components/ui/Tabs.tsx` - Radix-based tab wrapper component
- `web/src/components/omnir/SaveViewModal.tsx` - Dialog for creating views
- `web/src/components/omnir/ViewManagerPanel.tsx` - Side panel for managing views
- `web/src/components/omnir/ViewPinBar.tsx` - Horizontal tab bar with drag-to-reorder
- Integrated into: `LeadsPage`, `ContactsPage`, `AccountsPage`, `DealsPage`

## Test Results

### 1. View Pin Bar

#### 1.1 Pin bar renders above list tables
- [ ] **Contacts page** - Pin bar visible above table
- [ ] **Accounts page** - Pin bar visible above table
- [ ] **Deals page** - Pin bar visible above table
- [ ] **Leads page** - Pin bar visible above table

**Status**: ⏳ Pending
**Notes**:

#### 1.2 Pinned views appear as tabs
- [ ] Pinned views render as tab buttons
- [ ] Active view highlighted in indigo color
- [ ] Tab labels show view names correctly
- [ ] Shared views show 👥 indicator

**Status**: ⏳ Pending
**Notes**:

#### 1.3 Tab interaction
- [ ] Clicking tab applies view filters to list
- [ ] Clicking active tab clears the view
- [ ] View selection updates URL/state correctly

**Status**: ⏳ Pending
**Notes**:

#### 1.4 UI buttons present
- [ ] "+ Save view" button visible at end of bar
- [ ] Gear icon (Settings) button visible
- [ ] Buttons are clickable and responsive

**Status**: ⏳ Pending
**Notes**:

### 2. Drag-to-Reorder

#### 2.1 Drag functionality
- [ ] Pinned view tabs can be grabbed/dragged
- [ ] Visual feedback during drag (opacity, cursor)
- [ ] Drop zones work correctly
- [ ] Tabs reorder visually on drop

**Status**: ⏳ Pending
**Notes**:

#### 2.2 Persistence
- [ ] New order persists after page reload
- [ ] Backend PIN order saved correctly (check Network tab)
- [ ] Order syncs across tabs/sessions

**Status**: ⏳ Pending
**Notes**:

### 3. Save View Modal

#### 3.1 Modal opening
- [ ] "+ Save view" button opens modal
- [ ] Modal renders with correct title
- [ ] Form fields are visible and functional

**Status**: ⏳ Pending
**Notes**:

#### 3.2 Form validation
- [ ] View name field is required
- [ ] Empty name disables Save button
- [ ] Valid name enables Save button
- [ ] Name placeholder is helpful

**Status**: ⏳ Pending
**Notes**:

#### 3.3 Toggles
- [ ] "Share with team" checkbox works
- [ ] "Pin to view bar" checkbox works (default: checked)
- [ ] Toggle states persist in form

**Status**: ⏳ Pending
**Notes**:

#### 3.4 Saving
- [ ] Save button triggers API call
- [ ] Loading state shows during save
- [ ] Success: view appears in pin bar (if pinned)
- [ ] Success: modal closes
- [ ] Error handling shows user-friendly message

**Status**: ⏳ Pending
**Notes**:

### 4. View Manager Panel

#### 4.1 Panel opening
- [ ] Gear icon opens right-side panel
- [ ] Panel slides in from right
- [ ] Close button/overlay closes panel

**Status**: ⏳ Pending
**Notes**:

#### 4.2 View lists
- [ ] Own views section shows user's views
- [ ] Shared views section shows team views
- [ ] Sections are clearly labeled
- [ ] Empty states show when no views

**Status**: ⏳ Pending
**Notes**:

#### 4.3 View indicators
- [ ] Shared views show 👥 indicator
- [ ] Active view is highlighted
- [ ] Visual distinction between own/shared views

**Status**: ⏳ Pending
**Notes**:

#### 4.4 View selection
- [ ] Clicking view row applies filters
- [ ] List updates to reflect view filters
- [ ] Panel closes after selection
- [ ] Active view state updates

**Status**: ⏳ Pending
**Notes**:

#### 4.5 Inline rename
- [ ] Pencil icon triggers edit mode
- [ ] Input field replaces view name
- [ ] Enter key confirms rename
- [ ] Escape key cancels rename
- [ ] API PATCH call sends update
- [ ] View name updates in all locations

**Status**: ⏳ Pending
**Notes**:

#### 4.6 Pin/Unpin
- [ ] Pin button toggles pin state
- [ ] Pinning adds view to pin bar
- [ ] Unpinning removes view from pin bar
- [ ] Button state updates immediately
- [ ] Changes persist after reload

**Status**: ⏳ Pending
**Notes**:

#### 4.7 Delete
- [ ] Trash icon triggers delete
- [ ] Loading state shows during deletion
- [ ] View removed from list on success
- [ ] If active view deleted, state clears
- [ ] Error handling for delete failures

**Status**: ⏳ Pending
**Notes**:

### 5. Filter Persistence + Unsaved Changes

#### 5.1 Applying saved view
- [ ] Selecting view applies saved filters
- [ ] Search filter applied correctly
- [ ] Sort order applied correctly
- [ ] Entity-specific filters applied (stage, source, etc.)
- [ ] List updates to show filtered results

**Status**: ⏳ Pending
**Notes**:

#### 5.2 Unsaved changes detection
- [ ] Changing search triggers unsaved state
- [ ] Changing sort triggers unsaved state
- [ ] Changing filters triggers unsaved state
- [ ] Amber banner appears on filter change

**Status**: ⏳ Pending
**Notes**:

#### 5.3 Unsaved changes banner
- [ ] Banner shows view name and "unsaved changes" text
- [ ] "Save as new" button opens Save modal
- [ ] "Update view" button updates current view
- [ ] Update button saves new filters to backend
- [ ] Banner disappears after save

**Status**: ⏳ Pending
**Notes**:

#### 5.4 Page title
- [ ] Active view name shown as page title
- [ ] Title clears when view deselected
- [ ] Title updates when switching views

**Status**: ⏳ Pending
**Notes**:

### 6. Entity Coverage

#### 6.1 Contacts Page
- [ ] ViewPinBar renders
- [ ] All features work on Contacts
- [ ] Contact-specific filters (stage) work

**Status**: ⏳ Pending
**Notes**:

#### 6.2 Accounts Page
- [ ] ViewPinBar renders
- [ ] All features work on Accounts
- [ ] Account-specific filters work

**Status**: ⏳ Pending
**Notes**:

#### 6.3 Deals Page
- [ ] ViewPinBar renders
- [ ] All features work on Deals
- [ ] Deal-specific filters (stage, pipeline) work

**Status**: ⏳ Pending
**Notes**:

#### 6.4 Leads Page
- [ ] ViewPinBar renders
- [ ] All features work on Leads
- [ ] Lead-specific filters (source, score) work

**Status**: ⏳ Pending
**Notes**:

## Code Quality Observations

### Positive
- ✓ Clean component structure with proper separation of concerns
- ✓ React Query hooks for data fetching with proper cache invalidation
- ✓ Zustand store with localStorage persistence
- ✓ Drag-and-drop using @dnd-kit library
- ✓ Proper TypeScript typing throughout
- ✓ Loading states and error handling patterns in place

### Concerns/Notes
- N/A (to be filled during testing)

## API Endpoints Tested

- [ ] `GET /api/v1/views?entity_type=contacts` - List views
- [ ] `POST /api/v1/views` - Create view
- [ ] `PATCH /api/v1/views/{id}` - Update view
- [ ] `DELETE /api/v1/views/{id}` - Delete view
- [ ] `POST /api/v1/views/{id}/pin` - Update pin order

**API Response Samples**: (to be added during testing)

## Browser Compatibility

- [ ] Chrome/Chromium (latest)
- [ ] Firefox (latest)
- [ ] Safari (if available)

## Mobile Responsiveness

- [ ] 375px (iPhone SE) - Pin bar stacks/adapts appropriately
- [ ] 768px (iPad) - Tabs layout works
- [ ] 1024px+ (Desktop) - Full layout

## Performance Observations

- Initial load time: TBD
- View switching time: TBD
- Drag-and-drop responsiveness: TBD

## Bugs Found

_None yet - testing in progress_

## Overall Assessment

**Status**: 🔄 Testing in Progress
**Recommendation**: TBD

---

## Testing Notes

_Detailed observations and steps taken during manual testing will be added here._
