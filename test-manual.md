# Manual Testing Guide for Window Control Feature

## Testing Steps

### 1. Install the new Finicky build
```bash
rm -rf /Applications/Finicky.app
cp -r apps/finicky/build/Finicky.app /Applications/
```

### 2. Set Finicky as the default browser
Open System Preferences → General → Default web browser → Finicky

### 3. Create a test config file at `~/.finicky.js`
```javascript
export default {
  defaultBrowser: "Google Chrome",
  options: {
    logRequests: true,
  },
  handlers: [
    {
      match: "*newwindow*",
      browser: {
        name: "Google Chrome",
        newWindow: true
      }
    },
    {
      match: "*incognito*",
      browser: {
        name: "Google Chrome",
        incognito: true
      }
    },
    {
      match: "*both*",
      browser: {
        name: "Google Chrome",
        newWindow: true,
        incognito: true
      }
    },
  ]
};
```

### 4. Test URLs

Test these URLs by clicking on them from various applications (Mail, Messages, Terminal, etc.):

- **New Window Test**: `https://example.com/newwindow`
  - Expected: Opens in a new Chrome window
  - Dry-run command should show: `open -a "Google Chrome" -n --args --new-window "https://example.com/newwindow"`

- **Incognito Test**: `https://example.com/incognito`
  - Expected: Opens in Chrome incognito mode (new window)
  - Dry-run command should show: `open -a "Google Chrome" --args --incognito "https://example.com/incognito"`

- **Combined Test**: `https://example.com/both`
  - Expected: Opens in new Chrome window with incognito
  - Dry-run command should show: `open -a "Google Chrome" -n --args --new-window --incognito "https://example.com/both"`

- **Default Test**: `https://example.com/normal`
  - Expected: Opens in Chrome normally (new tab in existing window)
  - Dry-run command should show: `open -a "Google Chrome" "https://example.com/normal"`

### 5. Check logs
```bash
tail -f ~/Library/Logs/Finicky/Finicky*.log
```

Look for:
- "Resolving Chromium window flags" debug messages
- "Run command" messages showing the actual `open` commands
- Confirm flags are correctly appended

### 6. Test with other Chromium browsers

Update config to test with:
- Brave Browser
- Microsoft Edge
- Vivaldi

Example:
```javascript
{
  match: "*brave*",
  browser: {
    name: "Brave Browser",
    newWindow: true
  }
}
```

### 7. Test with non-Chromium browser (should gracefully ignore flags)

```javascript
{
  match: "*firefox*",
  browser: {
    name: "Firefox",
    newWindow: true  // Should be ignored, no error
  }
}
```

Expected: Firefox opens normally without the newWindow flag
