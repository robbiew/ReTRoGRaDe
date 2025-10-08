# Retrograde Application Server - Product Context

## Project Overview

Retrograde Application Server is a telnet-based Bulletin Board System (BBS).

## Project Goals

- Provide network application processing for BBS enthusiasts

## Core Features

- **Telnet Server**: Custom telnet server on port 2323 with character-mode negotiation
- **Session Management**: User authentication and security level management
- **SQLite Database**: All BBS data stored in SQLite with comprehensive schema including users, sessions, configuration, and audit logs
- **ANSI Art Support**: Display of ANSI artwork for enhanced user experience
- **TUI Configuration Editor**: Modern BubbleTea-based terminal interface for server configuration
- **Configurable Menus and Prompts**: Sysops can define menus and commands
- **JAM Message Base Format**: Messages are stored in JAM format
- **Fidonet Style Echomail and Netmail**: Echmail network support, including import and export
- **Configurable registration questions**: Sysop can decide which questiopns to ask new users

## Current Development Priorities

### Message Base Implementation

**Status**: In Progress - High Priority

The next phase focuses on completing the JAM message base implementation and user interface:

#### Message Areas

- **Local Messages**: User-to-user messaging within the BBS
- **Echomail Areas**: FidoNet-style public conferences with SEEN-BY/PATH routing
- **Netmail**: Private point-to-point messaging between FidoNet nodes

#### Message Base Features

- **JAM Format**: Full implementation of JAM message base specification
- **Message Reading**: Threaded message reading with quote support
- **Message Posting**: Rich text composition with automatic formatting
- **Area Management**: Configurable message areas and moderation
- **Network Support**: Echomail import/export with FidoNet compatibility

#### Implementation Plan

1. Complete JAM message base library implementation
2. Build message reading and posting UI
3. Implement echomail routing and networking
4. Add message base configuration in TUI editor
5. Integrate with existing user authentication system

## Architecture Overview

- **Go Language**: Backend implementation using Go 1.22.2
- **Telnet Protocol**: Direct telnet connections with custom input/output handling
- **Modular Design**: Separated concerns across multiple files

## Target Users

- BBS enthusiasts accessing applications via telnet clients

## Dependencies
