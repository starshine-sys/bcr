# bcr

A command handler based on [arikawa](https://github.com/diamondburned/arikawa). Mostly for personal use, but feel free to use it elsewhere 🙂

Package `bot` contains a basic wrapper around `bcr`, for a categorized help command.

## Example

(replace `"token"` with your bot's token, and the user ID with your own ID)

```go
// create a router
router, err := bcr.NewWithState("token", []discord.UserID{0}, []string{"~"})

// make sure to add the message create handler
router.State.AddHandler(router.MessageCreate)

// add a command
router.AddCommand(&bcr.Command{
    Name:    "ping",
    Summary: "Ping pong!",

    Command: func(ctx *bcr.Context) (err error) {
        _, err = ctx.Send("Pong!", nil)
        return
    },
})

// connect to discord
if err := bot.Router.State.Open(); err != nil {
    log.Fatalln("Failed to connect:", err)
}
```