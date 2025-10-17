# Menu Command Key Execution

## Linking Menu Commands

When a user activates a menu command, Retrograde executes all of the commands which have the command letters which were entered by the user. If two different commands both have the same command letters in them, both commands are executed in the order they are in the menu.

When linking commands together, remember to put a description only on the first command in the link, and set the rest to null. This stops Retrograde from displaying the command several times on the generic menu.

## Menu System Commands (Command Keys)

The Command Key (CmdKey)'s are 2 letter codes which make Retrograde do a certain function.

### Command Metadata Fields

Each menu command now stores additional metadata to help with presentation and future help text:

- **Short Desc** – the concise label rendered in menus when the command is visible.
- **Long Desc** – an extended explanation shown in editors (and future help screens) so SysOps can document command behaviour.
- **Hidden** – when enabled the command remains active but is omitted from the rendered `[Key] Description` list, letting you build silent link chains or background automation.

## Special Menu Keys

Some special keys can be used in the 'Keys' field that aren't mapped to alpha-numeric keys. These special keys are:

- 'FIRSTCMD' as a key executes that command everytime the menu is first loaded.
- 'ENTER' as a key is the same as the user pressing 'ENTER'.
- 'ESC' as a key is the same as the user pressing 'ESCAPE'.
- 'ANYKEY' as a key executes that command no matter what key the user presses.
- 'NOKEY' as a key executes that command when the user does not press any key within the timeout period (defined in 'Options' in seconds)
- 'TAB' as a key is the same as the user pressing the 'TAB' key.
- 'F1' through 'F12' as keys execute that command when the user presses the corresponding function key.

### Example Menu Command with Special Key

    Command Number: 3
    Keys          : NOKEY
    Short Desc    : Auto Logoff
    Long Desc     : This command will logoff the user automatically if no key is pressed within 5 seconds.
    ACS Required  :
    CmdKeys       : G
    Options       : 5
    Active        : Yes
    Hidden        : Yes
