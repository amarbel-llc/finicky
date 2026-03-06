# Browser Window Control Implementation Summary

## Overview
Successfully implemented support for launching Chromium browsers with command-line arguments to control window behavior, enabling URLs to open in new Chrome windows instead of tabs.

## Changes Made

### 1. TypeScript Config Schema (`packages/config-api/src/configSchema.ts`)
Added three optional boolean fields to both `BrowserConfigSchema` and `BrowserConfigStrictSchema`:
- `newWindow?: boolean` - Opens URL in a new browser window
- `incognito?: boolean` - Opens URL in incognito/private mode
- `newTab?: boolean` - Opens URL in new tab (documented for completeness, default behavior)

### 2. TypeScript Config Processing (`packages/config-api/src/index.ts`)
Updated `createBrowserConfig()` to include defaults for the new fields:
```typescript
const defaults = {
  // ... existing fields
  newWindow: undefined,
  incognito: undefined,
  newTab: undefined,
};
```

### 3. Go Browser Config Struct (`apps/finicky/src/browser/launcher.go`)
Added corresponding fields to the `BrowserConfig` struct:
```go
type BrowserConfig struct {
  // ... existing fields
  NewWindow *bool `json:"newWindow"`
  Incognito *bool `json:"incognito"`
  NewTab    *bool `json:"newTab"`
}
```

### 4. Go Window Flags Resolver (`apps/finicky/src/browser/launcher.go`)
Implemented `resolveChromiumWindowFlags()` function that:
- Checks if browser is Chromium-based via `browsers.json`
- Converts convenience fields to Chrome command-line flags:
  - `newWindow: true` → `--new-window`
  - `incognito: true` → `--incognito`
- Returns flags array and success boolean
- Gracefully handles non-Chromium browsers (returns nil, false)

### 5. Go Browser Launcher Integration (`apps/finicky/src/browser/launcher.go`)
Updated `LaunchBrowser()` function to:
- Call `resolveChromiumWindowFlags()` to get window control flags
- Add `-n` flag when newWindow is set (ensures fresh browser instance)
- Insert window flags in proper order: profile args → window flags → custom args → URL
- Command structure: `open -a "Browser" [-n] --args [profile] [window-flags] [custom-args] [URL]`

### 6. Unit Tests

**TypeScript Tests** (`packages/config-api/src/utils.test.ts`):
- Added 4 new tests for `resolveBrowser()` function
- Tests cover: newWindow flag, incognito flag, newTab flag, and combined usage with profile
- All 83 tests pass ✅

**Go Tests** (`apps/finicky/src/browser/launcher_test.go`):
- Added comprehensive tests for `resolveChromiumWindowFlags()`
- Tests cover: Chrome with various flags, non-Chromium browsers, unknown browsers
- All 7 test cases pass ✅

## Example Usage

### Basic New Window
```javascript
{
  match: "*.workplace.com/*",
  browser: {
    name: "Google Chrome",
    newWindow: true
  }
}
```

### Incognito Mode
```javascript
{
  match: ["*.bank.com/*", "*.financial.com/*"],
  browser: {
    name: "Brave Browser",
    incognito: true
  }
}
```

### Combined with Profile
```javascript
{
  match: "*.work-app.com/*",
  browser: {
    name: "Google Chrome",
    profile: "Work",
    newWindow: true
  }
}
```

## Browser Compatibility

**Supported (Chromium-based):**
- Google Chrome (stable, beta, canary)
- Brave Browser
- Microsoft Edge
- Vivaldi
- Chromium
- Opera / Opera GX
- Wavebox
- Other browsers detected as type="Chromium" in `browsers.json`

**Unsupported (gracefully ignored):**
- Safari
- Firefox
- Arc
- Any non-Chromium browser

## Generated Commands

### New Window Example
```bash
open -a "Google Chrome" -n --args --new-window "https://example.com"
```

### Incognito Example
```bash
open -a "Google Chrome" --args --incognito "https://example.com"
```

### Combined Example (profile + newWindow + incognito)
```bash
open -a "Google Chrome" -n --args --profile-directory="Work" --new-window --incognito "https://example.com"
```

### Default (no flags)
```bash
open -a "Google Chrome" "https://example.com"
```

## Testing Results

### Unit Tests
- ✅ TypeScript: 83/83 tests passing
- ✅ Go: 7/7 tests passing

### Test Coverage
- Schema validation ✅
- Config resolution ✅
- Chromium browser detection ✅
- Flag generation ✅
- Non-Chromium browser handling ✅
- Combined features (profile + flags) ✅

## Files Modified

1. `packages/config-api/src/configSchema.ts` - Schema definitions
2. `packages/config-api/src/index.ts` - Config processing
3. `apps/finicky/src/browser/launcher.go` - Core implementation
4. `packages/config-api/src/utils.test.ts` - TypeScript tests
5. `apps/finicky/src/browser/launcher_test.go` - Go tests (new file)

## Build Status

- ✅ TypeScript packages built successfully
- ✅ Go binary compiled successfully
- ✅ All tests passing
- ⚠️  Application ready for installation and manual testing

## Next Steps

1. Install the built application: `cp -r apps/finicky/build/Finicky.app /Applications/`
2. Set Finicky as default browser in System Preferences
3. Create test config with window control handlers
4. Test with various Chromium browsers
5. Verify logs show correct flags being applied

See `test-manual.md` for detailed manual testing instructions.

## Architecture Notes

**Design Decision**: Window flag conversion happens in Go (not TypeScript) because:
- Keeps TypeScript layer declarative
- Centralizes platform-specific logic
- Mirrors existing profile support pattern
- Easier to extend for other browser types

**Edge Cases Handled**:
- Non-Chromium browsers: flags silently ignored
- Unknown browsers: flags not applied
- Nil vs false values: only true adds flags
- Conflicting user args: both passed, Chrome resolves precedence
- Profile + newWindow: -n flag automatically added
