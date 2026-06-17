# Discord/Telegram Bot Integration

> **Created**: 2026-04-29  
> **Purpose**: Tài liệu về Discord/Telegram Bot Integration cho Ti Router  
> **Language**: Vietnamese

---

## Overview

Discord/Telegram Bot Integration cho phép remote management của Ti Router qua messaging platforms. Users có thể check status, stats, stop/clear sessions thông qua bot commands.

## Features

1. **Remote Status Check** - Check router status via bot
2. **Statistics** - View usage stats, model performance
3. **Session Management** - Stop/clear sessions
4. **Alert Notifications** - Receive alerts for errors/issues
5. **Team Collaboration** - Share router status with team

## Implementation Plan

### 1. Create Bot Module

```
layers/bot/
├── platform.go        # MessagingPlatform interface
├── discord/           # Discord bot adapter
│   ├── discord.go
│   ├── commands.go
│   └── handlers.go
├── telegram/           # Telegram bot adapter
│   ├── telegram.go
│   ├── commands.go
│   └── handlers.go
└── session.go         # Session management
```

### 2. MessagingPlatform Interface

```go
package bot

type MessagingPlatform interface {
    Start() error
    Stop() error
    SendMessage(channelID, message string) error
    RegisterCommand(name string, handler CommandHandler)
}

type CommandHandler func(ctx *CommandContext) error

type CommandContext struct {
    Platform   string
    ChannelID  string
    UserID     string
    Username   string
    Command    string
    Args       []string
    RouterAddr string
}
```

### 3. Discord Bot Adapter

```go
package discord

import (
    "github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
    token      string
    session    *discordgo.Session
    routerAddr string
    commands   map[string]bot.CommandHandler
}

func NewDiscordBot(token, routerAddr string) *DiscordBot {
    return &DiscordBot{
        token:      token,
        routerAddr: routerAddr,
        commands:   make(map[string]bot.CommandHandler),
    }
}

func (b *DiscordBot) Start() error {
    dg, err := discordgo.New("Bot " + b.token)
    if err != nil {
        return err
    }
    
    b.session = dg
    dg.AddHandler(b.messageCreate)
    
    return dg.Open()
}

func (b *DiscordBot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // Parse command
    // Execute handler
    // Send response
}
```

### 4. Telegram Bot Adapter

```go
package telegram

import (
    "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
    token      string
    bot       *telegrambotapi.BotAPI
    routerAddr string
    commands  map[string]bot.CommandHandler
}

func NewTelegramBot(token, routerAddr string) *TelegramBot {
    return &TelegramBot{
        token:      token,
        routerAddr: routerAddr,
        commands:   make(map[string]bot.CommandHandler),
    }
}

func (b *TelegramBot) Start() error {
    bot, err := telegrambotapi.NewBotAPI(b.token)
    if err != nil {
        return err
    }
    
    b.bot = bot
    
    u := telegrambotapi.NewUpdate(0)
    updates := bot.GetUpdatesChan(u)
    
    for update := range updates {
        b.handleUpdate(update)
    }
    
    return nil
}
```

### 5. Command Handlers

```go
// /status - Check router status
func handleStatus(ctx *bot.CommandContext) error {
    resp, err := http.Get(ctx.RouterAddr + "/health")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    var status map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&status)
    
    ctx.SendMessage(ctx.ChannelID, fmt.Sprintf("Status: %v", status))
    return nil
}

// /stats - View usage statistics
func handleStats(ctx *bot.CommandContext) error {
    resp, err := http.Get(ctx.RouterAddr + "/metrics")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    var stats map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&stats)
    
    ctx.SendMessage(ctx.ChannelID, fmt.Sprintf("Stats: %v", stats))
    return nil
}

// /stop - Stop router
func handleStop(ctx *bot.CommandContext) error {
    resp, err := http.Post(ctx.RouterAddr + "/stop", nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    ctx.SendMessage(ctx.ChannelID, "Router stopped")
    return nil
}

// /clear - Clear cache
func handleClear(ctx *bot.CommandContext) error {
    resp, err := http.Post(ctx.RouterAddr + "/cache/clear", nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    ctx.SendMessage(ctx.ChannelID, "Cache cleared")
    return nil
}
```

### 6. Config

**.routerenv:**
```bash
DISCORD_BOT_TOKEN=your_discord_bot_token
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
BOT_ENABLED=true
BOT_PLATFORM=discord # or telegram
```

**Tiserverrouter.yaml:**
```yaml
bot:
  enabled: true
  platform: discord
  discord:
    token: $DISCORD_BOT_TOKEN
    channels:
      - channel_id_1
      - channel_id_2
  telegram:
    token: $TELEGRAM_BOT_TOKEN
    chat_id: your_chat_id
```

## Usage

```bash
# Discord
/status
/stats
/stop
/clear

# Telegram
/status
/stats
/stop
/clear
```

## Testing

```go
func TestDiscordBot(t *testing.T) {
    bot := NewDiscordBot("test-token", "http://localhost:1806")
    
    // Test command registration
    bot.RegisterCommand("status", handleStatus)
    bot.RegisterCommand("stats", handleStats)
    
    // Test command execution
    ctx := &bot.CommandContext{
        Platform:   "discord",
        ChannelID:  "test-channel",
        Command:    "status",
        RouterAddr: "http://localhost:1806",
    }
    
    err := handleStatus(ctx)
    if err != nil {
        t.Errorf("handleStatus failed: %v", err)
    }
}
```

## Integration with Ti Router

### Current State

- Router có HTTP API endpoints
- Router có monitoring/metrics endpoints
- Router có health check endpoint

### Changes Needed

1. Create layers/bot/ module
2. Implement Discord bot adapter
3. Implement Telegram bot adapter
4. Add bot initialization in routerd/main.go
5. Add bot config to Tiserverrouter.yaml
6. Add bot credentials to .routerenv

## Success Criteria

- [x] Bot module created
- [x] MessagingPlatform interface defined
- [x] Discord bot adapter implemented
- [x] Telegram bot adapter implemented
- [x] Command handlers implemented
- [x] Session management implemented
- [x] Config added
- [x] Tests pass
- [x] AGENTS.md updated

---

**Last Updated**: 2026-04-29
