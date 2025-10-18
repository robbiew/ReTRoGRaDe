# Active Context

## Current Focus

- Completing new user registration and login
- Implementing Menu Editor to allow constructing of menus in TUI
- Create essential menu commands/command keys for Main Menu
- build the UI experience in the BBS for menus

## Next Focus

- Wiring TUI conference/area editors into JAM message workflow (listings, scanning, posting)
- Implementing JAM message base functionality and user interface
- Building message reading, posting, and area management features
- Adding echomail and netmail support for FidoNet networking

## Recent Changes

- Optimized file/folder organization
- Improved ANSI art printing functions
- Improved login UI
- Completed user management interface in TUI configuration editor
- Added TUI editors for message conferences and areas, including default Local Areas → General Chatter seed data
- Implemented full JAM message base library with local/echomail/netmail support
- Added guided first-time setup with database initialization
- Updated all documentation to reflect current application state

## Open Questions/Issues

- Message base UI integration with main BBS interface
- Echomail routing and network connectivity implementation
- Integration of IceEdit, legacy DOSemu-based BBS Full Screen Editor
- Creation of "prelogin" menu that can run scripts, dosemu programs/doors and other commands after a user logs in, and before loading first menu (e.g. MainMenu)
