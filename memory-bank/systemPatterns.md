# System Patterns

## Architectural Patterns

- **Modular Design**: Separated concerns across multiple packages (auth, config, database, etc.)
- **Dependency Injection**: Use of interfaces and factory functions (e.g., getStorage()) for database abstraction
- **Global State Management**: Limited use of global variables, preferring function-based access

## Code Patterns

- **Error Handling**: Consistent use of error wrapping and checking
- **Validation Loops**: Repeated input validation with retry mechanisms in user prompts
- **Security Measures**: Password hashing, IP blocklisting, attempt limits
- **Centralized UI Helpers**: Shared terminal interaction helpers (prompts, passwords, errors, ANSI sequences) in ui.InteractiveTerminal
- **Consistent Screen Positioning**: Unified escape sequence helpers for cursor positioning across UI components
- **Centralized Sanitization**: Single filename sanitization utility replacing duplicate regex implementations
- **Tightened ANSI Art Resolution**: Config-driven art directory management with metadata trimming

## Communication Patterns

- **Telnet Protocol**: Direct telnet connections with custom I/O handling
- **ANSI Art Display**: Support for terminal graphics and colors
