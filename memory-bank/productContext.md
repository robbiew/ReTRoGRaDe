# Retrograde Application Server - Product Context

## Project Overview

Retrograde Application Server is a telnet-based Bulletin Board System (BBS).

## Project Goals

- Provide network application processing for BBS enthusiasts

## Core Features

- **Telnet Server**: Custom telnet server on port 2323 with character-mode negotiation
- **Session Management**: User authentication and security level management
- **SQLite Data Storage**: BBS Configuration and User data stored in SQLite with helper functions
- **ANSI Art Support**: Display of ANSI artwork for enhanced user experience
- **TUI Configuration Editor**: Modern BubbleTea-based terminal interface for server configuration
- **Configurable Menus and Prompts**: Sysops can define menus and commands
- **JAM Message Base Format**: Messages are stored in JAM format
- **Fidonet Style Echomail and Netmail**: Echmail network support, including import and export
- **Configurable registration questions**: Sysop can decide which questiopns to ask new users

## Current Development Priorities

### User Security Levels & Management

**Status**: In Progress - High Priority

The next phase focuses on implementing a comprehensive user security and management system:

#### Security Levels

- **Regular User**: Level 10 - Standard user permissions
- **SysOp**: Level 100 - Full administrative access
- **Framework**: Extensible security level system (0-255 range)

#### User Management Tool

- **Location**: Configuration Editor → Editors → Users
- **Features**:
  - User listing with lightbar selection (Name, Level, UID)
  - User editing (handle, real name, location, email, call dates)
  - Security level management
  - Account status controls (active/inactive, locked/unlocked)

#### Implementation Plan

1. Define security level constants in configuration
2. Create user management interface in TUI editor
3. Implement user listing with database queries
4. Add user editing functionality
5. Integrate with existing authentication system

## Architecture Overview

- **Go Language**: Backend implementation using Go 1.22.2
- **Telnet Protocol**: Direct telnet connections with custom input/output handling
- **Modular Design**: Separated concerns across multiple files

## Target Users  

- BBS enthusiasts accessing applications via telnet clients

## Dependencies
