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

## Architecture Overview
- **Go Language**: Backend implementation using Go 1.22.2
- **Telnet Protocol**: Direct telnet connections with custom input/output handling
- **Modular Design**: Separated concerns across multiple files

## Target Users  
- BBS enthusiasts accessing applications via telnet clients

## Dependencies


