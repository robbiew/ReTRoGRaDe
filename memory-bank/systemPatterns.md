# System Patterns

## Architectural Patterns

- **Modular Design**: Separated concerns across multiple packages (auth, config, database, etc.)
- **Dependency Injection**: Use of interfaces and factory functions (e.g., getStorage()) for database abstraction
- **Global State Management**: Limited use of global variables, preferring function-based access

## Code Patterns

- **Error Handling**: Consistent use of error wrapping and checking
- **Validation Loops**: Repeated input validation with retry mechanisms in user prompts
- **Security Measures**: Password hashing, IP blacklisting, attempt limits

## Communication Patterns

- **Telnet Protocol**: Direct telnet connections with custom I/O handling
- **ANSI Art Display**: Support for terminal graphics and colors
