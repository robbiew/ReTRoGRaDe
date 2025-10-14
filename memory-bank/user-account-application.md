# User Account Application Guide

This document provides step-by-step instructions for new users to connect to Retrograde BBS and apply for an account via telnet.

## Prerequisites

- Telnet client installed on your system
- Network connectivity to the BBS server
- Valid hostname/IP address and port information

## Connection Instructions

### Telnet Connection

Connect using the telnet command:

```
telnet {hostname} {port}
```

**Default Port:** 23 (standard telnet port)

**Example:**
```
telnet bbs.example.com 23
```

## Account Application Process

1. **Connect to the BBS** using the telnet command above
2. **Enter "NEW"** when prompted for username (or enter your existing username to login)
3. **Enter Username** when prompted (3-20 characters, letters/numbers/spaces only)
4. **Enter Password** when prompted (4-20 characters, will be masked with *)
5. **Confirm Password** by re-entering it
6. **Enter Email** (required, up to 30 characters)
7. **Enter Terminal Preferences** (width/height, defaults provided)
8. **Review Account Summary** and confirm creation by pressing Y

Upon successful registration, you will be automatically logged in and can begin using the BBS.

## References

For detailed technical information about the authentication flow, security measures, and system implementation, see:

- [Complete Connection Flow Diagram](applicationFlow.md#complete-connection-flow-diagram)
- [Registration Flow](applicationFlow.md#phase-4-authentication---registration-flow)
- [Security Considerations](applicationFlow.md#security-considerations)
- [Error Handling](applicationFlow.md#error-handling--edge-cases)