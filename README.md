# bcr

A command handler based on [arikawa](https://github.com/diamondburned/arikawa).  
Made with sharding in mind so it kinda sucks for bots made to run on a single server, but oh well

Mostly for personal use, but feel free to use it elsewhere 🙂

Package `bot` contains a basic wrapper around `bcr`, for a categorized help command.

## Example

(replace `"token"` with your bot's token, and the user ID with your own ID)

```go
// create a router
router, err := bcr.NewWithState("token", []discord.UserID{0}, []string{"~"})

// make sure to add the message and interaction create handlers
router.AddHandler(router.MessageCreate)
router.AddHandler(router.InteractionCreate)

// add a command
router.AddCommand(&bcr.Command{
    Name:    "ping",
    Summary: "Ping pong!",

    SlashCommand: func(ctx bcr.Contexter) error {
        return ctx.SendX("Pong!")
    },
    Options: &[]discord.CommandOption{},
})

// populate router.Bot before running this
if err := router.SyncCommands(); err != nil {
    log.Fatalln("Failed to sync slash commands:", err)
}

// connect to discord
if err := router.ShardManager.Open(context.Background()); err != nil {
    log.Fatalln("Failed to connect:", err)
}

// block forever
select {}
```
