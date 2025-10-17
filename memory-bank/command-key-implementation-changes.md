# CmdKeys Selection Feature - Implementation Summary

## What This Changes

Instead of typing command keys like "MM", "MP", or "-/" manually in the CmdKeys field, users will see a beautiful, organized selection list with categories and descriptions.

## Visual Flow

```
BEFORE:
┌─────────────────────────────────────┐
│ Edit Command: R                     │
├─────────────────────────────────────┤
│ Keys:                [R________]    │
│ Short Description:   [Read Mail_]  │
│ ACS Required:        [_________]    │
│ CmdKeys:            [MM_______]  ← User types "MM"
│ Options:             [_________]    │
└─────────────────────────────────────┘

AFTER:
┌─────────────────────────────────────┐
│ Edit Command: R                     │
├─────────────────────────────────────┤
│ Keys:                [R________]    │
│ Short Description:   [Read Mail_]  │
│ ACS Required:        [_________]    │
│ CmdKeys:             [MM]       ← Shows current value
│   Press Enter to select...          │
│ Options:             [_________]    │
└─────────────────────────────────────┘
        ↓ User presses Enter
┌──────────────────────────────────────────────────────────┐
│ Select CmdKeys                                           │
├──────────────────────────────────────────────────────────┤
│ — Offline Mail —                                         │
│  [!D] Download QWK Packet    Download offline mail...   │
│  [!P] Set Message Pointers   Set message read pointers  │
│  [!U] Upload REP Packet      Upload offline mail replie │
│                                                          │
│ — Message —                                              │
│  [MA] Change Message Base    Change to different base   │
│  [ME] Enter Message          Enter a new message        │
│  [ML] List Messages          List message titles        │
│ ▶[MM] Read Mail              Read messages addressed..  │← Selected
│  [MN] New Message Scan       Scan all new messages      │
│  [MP] Post Message           Post a new message         │
│  [MR] Read Messages          Read messages in base      │
│                                                          │
│ — File —                                                 │
│  [FA] Change File Base       Change to different area   │
│  [FB] Batch Download         Add files to batch queue   │
│                                                          │
│ 8-15 of 67                                              │
│ ↑↓ Navigate | Enter Select | Esc Cancel                │
└──────────────────────────────────────────────────────────┘
        ↓ User presses Enter
┌─────────────────────────────────────┐
│ Edit Command: R                     │
├─────────────────────────────────────┤
│ Keys:                [R________]    │
│ Short Description:   [Read Mail_]  │
│ ACS Required:        [_________]    │
│ CmdKeys:             [MM]       ← Value updated
│ Options:             [_________]    │
│                                     │
│ ✓ CmdKey set to: MM (Read Mail)   │
└─────────────────────────────────────┘
```

## Files Created

### 1. `/mnt/user-data/outputs/cmdkeys.go`
Enhanced version of `internal/menu/cmdkeys.go` with:
- `CmdKeyDefinition` struct (CmdKey, Name, Description, Category, Handler)
- Comprehensive registration of all Renegade command keys
- Methods to get definitions by category
- ~60+ command keys defined with proper metadata

Key features:
```go
type CmdKeyDefinition struct {
    CmdKey      string        // "MM", "MP", "G", etc.
    Name        string        // "Read Mail", "Post Message"
    Description string        // Full description
    Category    string        // "Message", "File", "System", etc.
    Handler     CmdKeyHandler // Function to execute
}
```

### 2. `/mnt/user-data/outputs/tui-cmdkey-select-implementation.md`
Complete step-by-step implementation guide with:
- Code changes needed for each file
- New SelectValue type addition
- SelectOption struct definition  
- Navigation mode handling
- View rendering code
- Full working code examples

## Categories Defined

Commands are organized into logical categories:

1. **Offline Mail** - !D, !P, !U (QWK/REP packet handling)
2. **Message** - MA, ME, ML, MM, MP, MR, MS, etc. (15+ commands)
3. **File** - FA, FB, FD, FF, FL, FN, FU, FV, FZ, etc. (11+ commands)
4. **Batch** - BC, BD, BL, BR, BU, B? (batch transfers)
5. **System** - G, HC, HI, HM (logoff, system functions)
6. **User** - O1, OA, OB, OC, OE, OF, OG, OL, ON, etc. (user info/interaction)
7. **Voting** - VA, VL, VR, VV, V# (voting questions)
8. **Navigation** - -^, -/, -" (menu navigation)
9. **Misc** - -!, -&, -%, -$ (special functions)

## Implementation Steps (High Level)

1. **Replace** `internal/menu/cmdkeys.go` with the new version
2. **Add** `SelectValue` to ValueType enum in `internal/tui/editor.go`
3. **Add** `SelectOption` struct and `SelectOptions` field to MenuItem
4. **Create** helper function `getCmdKeySelectOptions()` 
5. **Update** `setupMenuEditCommandModal()` to use SelectValue
6. **Add** `SelectingValue` navigation mode
7. **Implement** `handleSelectingValue()` input handler
8. **Implement** `renderSelectingValueView()` view function
9. **Update** main Update() and View() methods to handle new mode
10. **Test** the full flow

## Benefits for Users

✅ **No memorization** - Browse all available commands visually
✅ **Self-documenting** - See what each command does before selecting
✅ **Organized** - Commands grouped by category (Message, File, etc.)
✅ **Error-proof** - Can't type invalid command keys
✅ **Modern UX** - Smooth navigation with keyboard shortcuts
✅ **Discoverable** - Learn about commands you didn't know existed

## Technical Benefits

✅ **Type-safe** - SelectValue type ensures consistency
✅ **Extensible** - Easy to add new commands to registry
✅ **Maintainable** - Single source of truth for command definitions
✅ **Reusable** - Selection UI pattern can be used elsewhere
✅ **Testable** - Command registry is separate from UI

## Example Command Definitions

```go
r.Register(&CmdKeyDefinition{
    CmdKey:      "MM",
    Name:        "Read Mail",
    Description: "Read messages addressed to you",
    Category:    "Message",
    Handler:     handleReadMail,
})

r.Register(&CmdKeyDefinition{
    CmdKey:      "-/",
    Name:        "Gosub Menu",
    Description: "Go to menu and return",
    Category:    "Navigation",
    Handler:     handleNotImplemented,
})

r.Register(&CmdKeyDefinition{
    CmdKey:      "FD",
    Name:        "Download File",
    Description: "Download a file from current area",
    Category:    "File",
    Handler:     handleNotImplemented,
})
```

## Next Steps

After implementation, you can:
1. Add search/filter functionality (press '/' to filter)
2. Color-code implemented vs. unimplemented commands
3. Show which commands are already used in the current menu
4. Add tooltips showing the command's options format
5. Reuse this pattern for other selection fields (themes, menus, etc.)

## GitHub Reference

All code references the official repository:
https://github.com/robbiew/retrograde

Specifically:
- `internal/menu/cmdkeys.go` - Command key registry
- `internal/tui/editor.go` - TUI core types
- `internal/tui/editor_view.go` - Menu command editing
- `internal/tui/editor_update.go` - Input handling