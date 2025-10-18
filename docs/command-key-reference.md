# Command Key Reference

This document consolidates every menu command key described in the Renegade
v1.30 documentation (Chapter 11) and maps them into Retrograde's menu editor.
Use it as the canonical list when wiring menus or deciding which legacy
commands we still need to implement.

- **Function** comes directly from the original manual.
- **Option(s)** reflects the syntax shown in the manual. Angle brackets (`<>`)
  denote optional values, square brackets (`[]`) mean required literal input,
  and braces (`{}`) indicate mutually exclusive choices.
- **Implemented** uses `✅` for commands Retrograde already handles; everything
  else still behaves as a placeholder.

## Working with command keys

1. In the TUI, open **Editors → Menus → Commands** and edit/insert a menu entry.
2. Use the **CmdKeys** selector to assign one of the keys below.
3. Provide any Options required by the command. If a command lists `None`, leave
   the Options field empty.
4. Many commands expect supporting files (bulletins, door batch files, etc.); be
   sure those resources exist under your configured paths.

## Full command list (Renegade v1.30)

The tables are grouped exactly how the original manual organizes them.

### Offline Mail (`!` commands)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `!D` | Download .QWK packet | None | No |
| `!P` | Set Message Pointers | None | No |
| `!U` | Upload .REP packet | None | No |

### Timebank (`$` commands)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `$D` | Deposit Time into Timebank | <Maxperday;Max Size of bank> | No |
| `$W` | Withdraw Time from Timebank | <Maxperday> | No |

### Sysop Functions (`*` commands)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `*B` | Enter the message base editor | None | No |
| `*C` | Change to a different user's account | None | No |
| `*D` | Enter the Mini-DOS environment | None | No |
| `*E` | Enter the event editor | None | No |
| `*F` | Enter the file base editor | None | No |
| `*L` | Show SysOp Log for certain day | None | No |
| `*N` | Edit a text file | None | No |
| `*P` | Enter the system configuration editor | None | No |
| `*R` | Enter Conference Editor | None | No |
| `*U` | Enter user editor | None | No |
| `*V` | Enter the voting editor | None | No |
| `*X` | Enter the protocol editor | None | No |
| `*Z` | Displays system activity log | None | No |
| `*1` | Edit file(s) in current file base | None | No |
| `*2` | Sort files in all file bases by name | None | No |
| `*3` | Read all users' private mail | None | No |
| `*4` | Download a file from anywhere on your computer | <filespec> | No |
| `*5` | Recheck files in current or all directories for size and online | None | No |
| `*6` | Upload file(s) not in file lists | None | No |
| `*7` | Validate files | None | No |
| `*8` | Add specs to all *.GIF files in current file base | None | No |
| `*9` | Pack the message bases | None | No |
| `*#` | Enter the menu editor | None | No |
| `*$` | Gives a long DOS directory of the current file base | None | No |
| `*%` | Gives a condensed DOS directory of the current file base | None | No |

### Navigation, Display, and Flow (`-`, `/`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `-C` | Display message on SysOp Window | <string> | No |
| `-F` | Display a text file | [filename] <.ext> | No |
| `/F` | Display a text file | [filename] <.ext> | No |
| `-L` | Display a line of text | [string] | ✅ |
| `-N` | Shows question, displays quote if Y is pressed, and continues | [question;quote] | No |
| `-Q` | Read an Infoform questionnaire file (answers in .ASW) | <Infoform questionnaire filename> | No |
| `-R` | Read an Infoform questionnaire answer file | <Infoform questionnaire filename> | No |
| `-S` | Append line to SysOp log file | [string] | No |
| `-Y` | Shows question, displays quote if N is pressed, and continues | [question;quote] | No |
| `-;` | Execute macro | [macro] | No |
| `-$` | Prompt for password | [password] < <[;prompt]> [;bad-message] > | No |
| `-^` | Goto menu | [menu file] | ✅ |
| `-/` | Gosub menu | [menu file] | No |
| `-\` | Return from menu | None | No |

### Archive Management (`A*`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `AA` | Add files to archive | None | No |
| `AC` | Convert between archive formats | None | No |
| `AE` | Extract files from archive | None | No |
| `AG` | Manipulate files extracted from archives | None | No |
| `AM` | Modify comment fields in archive | None | No |
| `AR` | Re-archive archived files using same format | None | No |
| `AT` | Run integrity test on archive file | None | No |

### Batch File Transfer (`B*`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `BC` | Clear batch queue | <U> | No |
| `BD` | Download batch queue | None | No |
| `BL` | List batch queue | <U> | No |
| `BR` | Remove single file from batch queue | <U> | No |
| `BU` | Upload batch queue | None | No |
| `B?` | Display number of files left in batch download queue | None | No |

### Dropfile / Door Launch (`D*`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `DC` | Create CHAIN.TXT (WWIV door) and execute Option | [command to execute] | No |
| `DD` | Create DORINFO1.DEF (RBBS door) and execute Option | [command to execute] | No |
| `DG` | Create DOOR.SYS (GAP door) and execute Option | [command to execute] | No |
| `DP` | Create PCBOARD.SYS (PCBoard door) and execute Option | [command to execute] | No |
| `DS` | Create SFDOORS.DAT (Spitfire door) and execute Option | [command to execute] | No |
| `DW` | Create CALLINFO.BBS (Wildcat! door) and execute Option | [command to execute] | No |
| `D-` | Execute Option without creating a door information file | [command to execute] | No |

### File System (`F*`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `FA` | Change file bases | <base#> or {+/-} or <L> | No |
| `FB` | Add file to Batch Download List | < Filename > | No |
| `FD` | Download file on BBS to user | < Filename > | No |
| `FF` | Search all file bases for description | None | No |
| `FL` | List filespec in current file base only | Filespec (Overrides user input) | No |
| `FN` | Scan file sections for new files | <newtype> | No |
| `FP` | Change pointer date for new files | None | No |
| `FS` | Search all file bases for filespec | None | No |
| `FU` | Upload file from user to BBS | None | No |
| `FV` | List contents of an archived file | None | No |
| `FZ` | Set file bases to be scanned for new files | None | No |
| `F@` | Create temporary directory | None | No |
| `F#` | Display Line/Quick file base change | None | No |

### Hangup / Logoff (`H*`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `HC` | Careful logoff of user | [string] | No |
| `HI` | Immediate logoff of user | None | No |
| `HM` | Display string and logoff user | [string] | No |

### System (Single-Key)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `G` | Goodbye / Logoff | None | ✅ |

### Message System (`M*`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `MA` | Message base change | <base#> or {+/-} or <L> | No |
| `ME` | Send private mail to user | <User #> <;Reason> | No |
| `MK` | Edit/Delete outgoing private mail | None | No |
| `ML` | Send "mass mail" -  private mail sent to multiple users | None | No |
| `MM` | Read private mail | None | No |
| `MN` | Display new messages | <newtype> | No |
| `MP` | Post message in the current message base. | None | No |
| `MR` | Read messages in current base | None | No |
| `MS` | Scan messages in current base | <newtype> | No |
| `MU` | Lists users with access to the current message base | None | No |
| `MY` | Scan message bases for personal messages | None | No |
| `MZ` | Set message bases to be scanned for new messages | None | No |
| `M#` | Display Line/Quick message base change | None | No |

### Multinode (`N*`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `NA` | Toggle node page availability | None | No |
| `ND` | Hangup node | <Node #> | No |
| `NG` | Join Group Chat | None | No |
| `NO` | View users on all nodes | None | No |
| `NP` | Page another node for chat | <Node #> | No |
| `NS` | Send a message to another node | <node number> <;message to send> | No |
| `NT` | Stealth Mode On/Off | None | No |
| `NW` | Display String under Activity in Node Listing | [ String ] | No |

### System & User Operations (`O*`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `O1` | Logon to BBS (Shuttle) | None | No |
| `O2` | Apply to BBS as a new user (Shuttle) | None | No |
| `OA` | Allow auto-validation of users | [password]<;Level> | No |
| `OB` | User Statistics | <Letter> | No |
| `OC` | Page the SysOp | <user #> <;string> | No |
| `OE` | Pause Screen (centered) | <Override default pause text> | ✅ |
| `OF` | AR flag set/reset/toggle | [{function}{flag}] | No |
| `OG` | AC flag set/reset/toggle | [{function}{flag}] | No |
| `OL` | List today's callers | filename | No |
| `ON` | Clear Screen | None | No |
| `OP` | Modify user information | [info type] | No |
| `OR` | Change to another conference | <conference char> or <?> | No |
| `OS` | Go to bulletins menu | <main bulletin;sub-bulletin> | No |
| `OU` | User Listing | < ACS;filename > | No |
| `OV` | BBS Listing | <filename> | No |

### Automessage (`U*`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `UA` | Reply to author of current AutoMessage | None | No |
| `UR` | Display current AutoMessage | None | No |
| `UW` | Write AutoMessage | None | No |

### Voting (`V*`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `VA` | Add voting topic | None | No |
| `VL` | List voting topics | None | No |
| `VR` | View results of voting topic | <Question #> | No |
| `VT` | Track User's vote | <User #> | No |
| `VU` | View users who voted on Question | <Question #> | No |
| `VV` | Vote on all un-voted topics | None | No |
| `V#` | Vote on Question # | <Question #> | No |

### Credit System (`$+/ -`)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `$+` | Increase a user's credit | [ Amount ] | No |
| `$-` | Increase a user's debit | [ Amount ] | No |

### File Scanning (FILEP.MNU)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `L1` | Continue Listing | None | No |
| `L2` | Quit Listing | None | No |
| `L3` | Next file base | None | No |
| `L4` | Toggle NewScan of that base on/off | None | No |

### Message Scanning (READP.MNU)

| CmdKey | Function | Option(s) | Implemented |
|--------|----------|-----------|-------------|
| `RA` | Read Message Again | None | No |
| `RB` | Move Back in Thread | None | No |
| `RC` | Continuous Reading | None | No |
| `RD` | Delete Message | None | No |
| `RE` | Edit Message | None | No |
| `RF` | Forward in Thread | None | No |
| `RG` | Goto next Base | None | No |
| `RH` | Set Highread Pointer | None | No |
| `RI` | Ignore remaining messages, and set high pointer | None | No |
| `RL` | List Messages | None | No |
| `RM` | Move Message | None | No |
| `RN` | Next Message | None | No |
| `RQ` | Quit Reading | None | No |
| `RR` | Reply to Message | None | No |
| `RT` | Toggle NewScan of Message Base | None | No |
| `RU` | Edit User of Current Message | None | No |
| `RX` | Extract Message | None | No |
| `R#` | Allows User to Jump to message inputed. | None | No |
| `R-` | Read Previous Message | None | No |

---

**Next steps:** as Retrograde gains native support for these commands, update
this reference with implementation notes, return codes, and example Option
values so operators know which legacy behaviours still apply.
