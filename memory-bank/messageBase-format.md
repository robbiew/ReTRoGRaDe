# Retrograde - Message Base Formats

Retrograde Supports **three types of messages**:

## 1. **Local BBS Messages** (Non-FidoNet)
Simple messages between users on your BBS only:

```go
msg := jam.NewLocalMessage(
    "John",
    "Jane", 
    "Meeting Today",
    "Don't forget our 3pm meeting!",
)
mb.AddMessage(msg)
```

**Features:**
- No FidoNet addresses required
- No kludges, tearlines, or origin lines
- Uses `MSG_TYPELOCAL` attribute
- Perfect for private BBS conferences

## 2. **FidoNet Echomail** (Public Conference)
Public messages distributed across FidoNet:

```go
msg := jam.NewEchomailMessage(
    "SysOp",
    "All",
    "Welcome",
    "Message text here",
    "1:103/705", // Your FidoNet address
)
msg.MSGID = jam.GenerateMSGID("1:103/705", 12345)
msg.Origin = "My BBS"
msg.SeenBy = []string{"103/705", "103/1"}
msg.Path = []string{"103/705"}
mb.AddMessage(msg)
```

**Features:**
- Origin address required (FromAddr)
- Destination address optional (ToAddr not used)
- Includes Origin line, tearline, SEEN-BY, PATH
- Uses `MSG_TYPEECHO` attribute

## 3. **FidoNet Netmail** (Private Point-to-Point)
Private messages between specific FidoNet nodes:

```go
msg := jam.NewNetmailMessage(
    "John",
    "Jane",
    "Private Msg",
    "Secret stuff...",
    "1:103/705",    // From address
    "1:103/1.0",    // To address (with point)
)
msg.MSGID = jam.GenerateMSGID("1:103/705", 1)
msg.REPLY = "1:103/1 00001234" // Threading
mb.AddMessage(msg)
```

**Features:**
- Both FromAddr and ToAddr required
- Includes INTL, TOPT, FMPT kludges for routing
- No Origin line, SEEN-BY, or PATH (direct routing)
- Uses `MSG_TYPENET | MSG_PRIVATE` attributes
- Supports point addressing (node.point)

## Helper Methods

```go
// Check message type
if msg.IsLocal() {
    // Local BBS message
}
if msg.IsEchomail() {
    // FidoNet echomail
}
if msg.IsNetmail() {
    // FidoNet netmail
}
if msg.IsPrivate() {
    // Private message
}
```

The library automatically handles:
- Kludge formatting only for FidoNet messages
- INTL/TOPT/FMPT kludges for netmail routing
- Origin/tearline only for echomail
- SEEN-BY/PATH only for echomail
- Simple storage for local messages
