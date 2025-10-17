# Command Key Reference

This document consolidates every menu command key described in the Renegade
v1.30 documentation (Chapter 11) and maps them into Retrograde's menu editor.
Use it as the canonical list when wiring menus or deciding which legacy
commands we still need to implement.

- **Function** comes directly from the original manual.
- **Option(s)** reflects the syntax shown in the manual. Angle brackets (`<>`)
  denote optional values, square brackets (`[]`) mean required literal input,
  and braces (`{}`) indicate mutually exclusive choices.
- All commands currently act as placeholders unless noted in the code base.

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

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `!D` | Download .QWK packet | None |
| `!P` | Set Message Pointers | None |
| `!U` | Upload .REP packet | None |

### Timebank (`$` commands)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `$D` | Deposit Time into Timebank | <Maxperday;Max Size of bank> |
| `$W` | Withdraw Time from Timebank | <Maxperday> |

### Sysop Functions (`*` commands)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `*B` | Enter the message base editor | None |
| `*C` | Change to a different user's account | None |
| `*D` | Enter the Mini-DOS environment | None |
| `*E` | Enter the event editor | None |
| `*F` | Enter the file base editor | None |
| `*L` | Show SysOp Log for certain day | None |
| `*N` | Edit a text file | None |
| `*P` | Enter the system configuration editor | None |
| `*R` | Enter Conference Editor | None |
| `*U` | Enter user editor | None |
| `*V` | Enter the voting editor | None |
| `*X` | Enter the protocol editor | None |
| `*Z` | Displays system activity log | None |
| `*1` | Edit file(s) in current file base | None |
| `*2` | Sort files in all file bases by name | None |
| `*3` | Read all users' private mail | None |
| `*4` | Download a file from anywhere on your computer | <filespec> |
| `*5` | Recheck files in current or all directories for size and online | None |
| `*6` | Upload file(s) not in file lists | None |
| `*7` | Validate files | None |
| `*8` | Add specs to all *.GIF files in current file base | None |
| `*9` | Pack the message bases | None |
| `*#` | Enter the menu editor | None |
| `*$` | Gives a long DOS directory of the current file base | None |
| `*%` | Gives a condensed DOS directory of the current file base | None |

### Navigation, Display, and Flow (`-`, `/`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `-C` | Display message on SysOp Window | <string> |
| `-F` | Display a text file | [filename] <.ext> |
| `/F` | Display a text file | [filename] <.ext> |
| `-L` | Display a line of text | [string] |
| `-N` | Shows question, displays quote if Y is pressed, and continues | [question;quote] |
| `-Q` | Read an Infoform questionnaire file (answers in .ASW) | <Infoform questionnaire filename> |
| `-R` | Read an Infoform questionnaire answer file | <Infoform questionnaire filename> |
| `-S` | Append line to SysOp log file | [string] |
| `-Y` | Shows question, displays quote if N is pressed, and continues | [question;quote] |
| `-;` | Execute macro | [macro] |
| `-$` | Prompt for password | [password] < <[;prompt]> [;bad-message] > |
| `-^` | Goto menu | [menu file] |
| `-/` | Gosub menu | [menu file] |
| `-\` | Return from menu | None |

### Archive Management (`A*`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `AA` | Add files to archive | None |
| `AC` | Convert between archive formats | None |
| `AE` | Extract files from archive | None |
| `AG` | Manipulate files extracted from archives | None |
| `AM` | Modify comment fields in archive | None |
| `AR` | Re-archive archived files using same format | None |
| `AT` | Run integrity test on archive file | None |

### Batch File Transfer (`B*`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `BC` | Clear batch queue | <U> |
| `BD` | Download batch queue | None |
| `BL` | List batch queue | <U> |
| `BR` | Remove single file from batch queue | <U> |
| `BU` | Upload batch queue | None |
| `B?` | Display number of files left in batch download queue | None |

### Dropfile / Door Launch (`D*`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `DC` | Create CHAIN.TXT (WWIV door) and execute Option | [command to execute] |
| `DD` | Create DORINFO1.DEF (RBBS door) and execute Option | [command to execute] |
| `DG` | Create DOOR.SYS (GAP door) and execute Option | [command to execute] |
| `DP` | Create PCBOARD.SYS (PCBoard door) and execute Option | [command to execute] |
| `DS` | Create SFDOORS.DAT (Spitfire door) and execute Option | [command to execute] |
| `DW` | Create CALLINFO.BBS (Wildcat! door) and execute Option | [command to execute] |
| `D-` | Execute Option without creating a door information file | [command to execute] |

### File System (`F*`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `FA` | Change file bases | <base#> or {+/-} or <L> |
| `FB` | Add file to Batch Download List | < Filename > |
| `FD` | Download file on BBS to user | < Filename > |
| `FF` | Search all file bases for description | None |
| `FL` | List filespec in current file base only | Filespec (Overrides user input) |
| `FN` | Scan file sections for new files | <newtype> |
| `FP` | Change pointer date for new files | None |
| `FS` | Search all file bases for filespec | None |
| `FU` | Upload file from user to BBS | None |
| `FV` | List contents of an archived file | None |
| `FZ` | Set file bases to be scanned for new files | None |
| `F@` | Create temporary directory | None |
| `F#` | Display Line/Quick file base change | None |

### Hangup / Logoff (`H*`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `HC` | Careful logoff of user | [string] |
| `HI` | Immediate logoff of user | None |
| `HM` | Display string and logoff user | [string] |

### Message System (`M*`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `MA` | Message base change | <base#> or {+/-} or <L> |
| `ME` | Send private mail to user | <User #> <;Reason> |
| `MK` | Edit/Delete outgoing private mail | None |
| `ML` | Send "mass mail" -  private mail sent to multiple users | None |
| `MM` | Read private mail | None |
| `MN` | Display new messages | <newtype> |
| `MP` | Post message in the current message base. | None |
| `MR` | Read messages in current base | None |
| `MS` | Scan messages in current base | <newtype> |
| `MU` | Lists users with access to the current message base | None |
| `MY` | Scan message bases for personal messages | None |
| `MZ` | Set message bases to be scanned for new messages | None |
| `M#` | Display Line/Quick message base change | None |

### Multinode (`N*`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `NA` | Toggle node page availability | None |
| `ND` | Hangup node | <Node #> |
| `NG` | Join Group Chat | None |
| `NO` | View users on all nodes | None |
| `NP` | Page another node for chat | <Node #> |
| `NS` | Send a message to another node | <node number> <;message to send> |
| `NT` | Stealth Mode On/Off | None |
| `NW` | Display String under Activity in Node Listing | [ String ] |

### System & User Operations (`O*`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `O1` | Logon to BBS (Shuttle) | None |
| `O2` | Apply to BBS as a new user (Shuttle) | None |
| `OA` | Allow auto-validation of users | [password]<;Level> |
| `OB` | User Statistics | <Letter> |
| `OC` | Page the SysOp | <user #> <;string> |
| `OE` | Pause Screen | None |
| `OF` | AR flag set/reset/toggle | [{function}{flag}] |
| `OG` | AC flag set/reset/toggle | [{function}{flag}] |
| `OL` | List today's callers | filename |
| `ON` | Clear Screen | None |
| `OP` | Modify user information | [info type] |
| `OR` | Change to another conference | <conference char> or <?> |
| `OS` | Go to bulletins menu | <main bulletin;sub-bulletin> |
| `OU` | User Listing | < ACS;filename > |
| `OV` | BBS Listing | <filename> |

### Automessage (`U*`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `UA` | Reply to author of current AutoMessage | None |
| `UR` | Display current AutoMessage | None |
| `UW` | Write AutoMessage | None |

### Voting (`V*`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `VA` | Add voting topic | None |
| `VL` | List voting topics | None |
| `VR` | View results of voting topic | <Question #> |
| `VT` | Track User's vote | <User #> |
| `VU` | View users who voted on Question | <Question #> |
| `VV` | Vote on all un-voted topics | None |
| `V#` | Vote on Question # | <Question #> |

### Credit System (`$+/ -`)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `$+` | Increase a user's credit | [ Amount ] |
| `$-` | Increase a user's debit | [ Amount ] |

### File Scanning (FILEP.MNU)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `L1` | Continue Listing | None |
| `L2` | Quit Listing | None |
| `L3` | Next file base | None |
| `L4` | Toggle NewScan of that base on/off | None |

### Message Scanning (READP.MNU)

| CmdKey | Function | Option(s) |
|--------|----------|-----------|
| `RA` | Read Message Again | None |
| `RB` | Move Back in Thread | None |
| `RC` | Continuous Reading | None |
| `RD` | Delete Message | None |
| `RE` | Edit Message | None |
| `RF` | Forward in Thread | None |
| `RG` | Goto next Base | None |
| `RH` | Set Highread Pointer | None |
| `RI` | Ignore remaining messages, and set high pointer | None |
| `RL` | List Messages | None |
| `RM` | Move Message | None |
| `RN` | Next Message | None |
| `RQ` | Quit Reading | None |
| `RR` | Reply to Message | None |
| `RT` | Toggle NewScan of Message Base | None |
| `RU` | Edit User of Current Message | None |
| `RX` | Extract Message | None |
| `R#` | Allows User to Jump to message inputed. | None |
| `R-` | Read Previous Message | None |

---

**Next steps:** as Retrograde gains native support for these commands, update
this reference with implementation notes, return codes, and example Option
values so operators know which legacy behaviours still apply.
