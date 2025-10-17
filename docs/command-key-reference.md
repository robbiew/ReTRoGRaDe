# Quick Reference: Command Keys

## Most Common Commands

Note, these are 100% lifted from Renegade BBS. They will need to be tweaked.

### Message Commands
| CmdKey | Name | Description |
|--------|------|-------------|
| MM | Read Mail | Read messages addressed to you |
| MP | Post Message | Post a new message to current base |
| MA | Change Message Base | Change to a different message base |
| MR | Read Messages | Read messages in current base |
| ML | List Messages | List message titles in current base |
| ME | Enter Message | Enter a new message to current base |
| MS | Scan Messages | Quick scan of message subjects |
| MN | New Message Scan | Scan all new messages |
| MZ | Global New Scan | Scan new messages across all bases |
| M# | Read Message Number | Read a specific message by number |

### File Commands
| CmdKey | Name | Description |
|--------|------|-------------|
| FA | Change File Base | Change to a different file area |
| FL | List Files | List files in current area |
| FD | Download File | Download a file from current area |
| FU | Upload File | Upload a file to current area |
| FF | Find File | Search for files across all areas |
| FN | New Files Scan | Scan for new files |
| FV | View File | View a file's description or contents |
| FZ | Global File Search | Search all file areas |
| FB | Batch Download | Add files to batch download queue |
| F@ | Your Uploads | List files you've uploaded |

### System Commands
| CmdKey | Name | Description |
|--------|------|-------------|
| G | Goodbye / Logoff | Log off the BBS |
| HC | Careful Logoff | Prompt before logging off |
| HI | Instant Logoff | Immediate logoff without prompt |
| HM | Main Menu | Return to main menu |

### User Commands
| CmdKey | Name | Description |
|--------|------|-------------|
| OA | Apply for Access | New user application |
| OB | Bulletins | Read system bulletins |
| OC | Page Sysop | Page the system operator |
| OE | User Editor | Edit your user settings |
| OF | Feedback | Send feedback to sysop |
| OL | Last Callers | View list of recent callers |
| ON | Node List | View active nodes/users |
| OP | Page User | Page another user |
| OS | System Information | View BBS system information |
| OU | User List | View list of users |

### Navigation Commands
| CmdKey | Name | Description |
|--------|------|-------------|
| -^ | Go to Menu | Navigate to a different menu |
| -/ | Gosub Menu | Go to menu and return |
| -" | Return from Menu | Return to previous menu |

### Batch Transfer Commands
| CmdKey | Name | Description |
|--------|------|-------------|
| BC | Clear Batch Queue | Clear your batch transfer queue |
| BD | Batch Download | Download all files in batch queue |
| BL | List Batch Queue | List files in your batch queue |
| BR | Remove from Batch | Remove a file from batch queue |
| BU | Batch Upload | Upload multiple files at once |
| B? | Batch Queue Status | Display number of files in batch queue |

### Offline Mail Commands
| CmdKey | Name | Description |
|--------|------|-------------|
| !D | Download QWK Packet | Download offline mail in .QWK format |
| !P | Set Message Pointers | Set message read pointers for offline mail |
| !U | Upload REP Packet | Upload offline mail replies in .REP format |

### Voting Commands
| CmdKey | Name | Description |
|--------|------|-------------|
| VA | Add Voting Question | Add a new voting question |
| VL | List Voting Questions | List all voting questions |
| VR | View Voting Results | View results of voting questions |
| VV | Vote on All | Vote on all unvoted questions |
| V# | Vote on Question | Vote on a specific question |

### Miscellaneous Commands
| CmdKey | Name | Description |
|--------|------|-------------|
| -! | Execute Program | Execute an external program |
| -& | Display File | Display a text file |
| -% | Display String | Display a text string |
| -$ | Prompt for Password | Prompt user for password |

## Command Categories

1. **Offline Mail** (3 commands) - !D, !P, !U
2. **Message** (15 commands) - MA, ME, MK, ML, MM, MN, MP, MR, MS, MU, MY, MZ, M#
3. **File** (11 commands) - FA, FB, FD, FF, FL, FN, FP, FS, FU, FV, FZ, F@, F#
4. **Batch** (6 commands) - BC, BD, BL, BR, BU, B?
5. **System** (4 commands) - G, HC, HI, HM
6. **User** (11 commands) - O1, OA, OB, OC, OE, OF, OG, OL, ON, OP, OR, OS, OU, OV
7. **Voting** (5 commands) - VA, VL, VR, VT, VU, VV, V#
8. **Navigation** (3 commands) - -^, -/, -"
9. **Misc** (4 commands) - -!, -&, -%, -$

## Implementation Status

✅ **Implemented**: G
⏸️ **Placeholder**: All others (return "not yet implemented" message)

## Usage in Menu Configuration

When creating menu commands in the TUI:

1. Navigate to: **Editors → Menus → Select Menu → Commands**
2. Press **I** to insert a new command or select existing and press **Enter**
3. Fill in the fields:
   - **Keys**: The key(s) user presses (e.g., "R", "P", "G")
   - **Short Description**: What appears on menu (e.g., "Read Mail")
   - **ACS Required**: Access control string (optional)
   - **CmdKeys**: Press Enter to select from list! (e.g., "MM")
   - **Options**: Additional parameters (optional)

## Example Menu Commands

### Main Menu
```
Keys: R  | Short Desc: Read Mail    | CmdKeys: MM
Keys: P  | Short Desc: Post Message | CmdKeys: MP
Keys: F  | Short Desc: Files        | CmdKeys: -^ | Options: FILES
Keys: G  | Short Desc: Goodbye      | CmdKeys: G
```

### Files Menu
```
Keys: L  | Short Desc: List Files   | CmdKeys: FL
Keys: D  | Short Desc: Download     | CmdKeys: FD
Keys: U  | Short Desc: Upload       | CmdKeys: FU
Keys: S  | Short Desc: Search       | CmdKeys: FF
Keys: Q  | Short Desc: Quit         | CmdKeys: -"
```

### Message Menu
```
Keys: R  | Short Desc: Read         | CmdKeys: MR
Keys: N  | Short Desc: New Scan     | CmdKeys: MN
Keys: P  | Short Desc: Post         | CmdKeys: MP
Keys: M  | Short Desc: Your Mail    | CmdKeys: MM
Keys: Q  | Short Desc: Quit         | CmdKeys: -"
```

## Tips

- **Navigation**: Use `-^` (go to menu), `-/` (gosub), `-"` (return) for menu flow
- **Testing**: Start with MM, MP, and G as they're already implemented
- **Linking**: You can chain multiple commands together
- **ACS**: Leave blank to allow all users
- **Options**: Vary by command (see Renegade docs for details)

## Further Reading

- Original Renegade documentation: RGV130.DOC (uploaded)
- GitHub repository: https://github.com/robbiew/retrograde
- Implementation files:
  - `/mnt/user-data/outputs/cmdkeys.go` - Full command registry
