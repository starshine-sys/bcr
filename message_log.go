package bcr

import (
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
)

// Log is an updateable log message
type Log struct {
	ctx    *Context
	logger *Logger

	msg *discord.Message

	title string
	color discord.Color

	logs []string
}

// NewLog creates a new Log
func (ctx *Context) NewLog(title string) *Log {
	return &Log{
		ctx:    ctx,
		logger: ctx.Router.Logger,

		title: title,
		color: ColourBlue,
	}
}

// Send ...
func (log *Log) Send() {
	e := &discord.Embed{
		Title:       log.title,
		Color:       log.color,
		Description: strings.Join(log.logs, "\n"),
	}

	if log.msg == nil {
		m, err := log.ctx.Send("", e)
		if err != nil {
			log.logger.Error("Error sending log message: %v", err)
		}
		log.msg = m
		return
	}

	_, err := log.ctx.Edit(log.msg, "", e)
	if err != nil {
		log.logger.Error("Error sending log message: %v", err)
	}
	return
}

// Log logs a normal message
func (log *Log) Log(msg string) {
	log.logs = append(log.logs, msg)
	log.logger.Info(msg)
	log.Send()
}

// Error logs an error message
func (log *Log) Error(msg string) {
	log.color = ColourRed
	log.logs = append(log.logs, msg)
	log.logger.Error(msg)
	log.Send()
}

// Replace replaces the latest message
func (log *Log) Replace(msg string) {
	log.logger.Info(msg)

	if len(log.logs) == 0 {
		log.logs = append(log.logs, msg)
	} else {
		log.logs[len(log.logs)-1] = msg
	}
	log.Send()
}

// SetTitle ...
func (log *Log) SetTitle(title string) {
	log.title = title
}

// SetColor ...
func (log *Log) SetColor(color discord.Color) {
	log.color = color
}
