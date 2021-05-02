package main

type CommandID int

const (
	CMD_JOIN CommandID = iota
	CMD_ROOMS
	CMD_QUIT
	CMD_LEAVE
	CMD_START
	CMD_LIST
	CMD_KILL
	CMD_INSPECT
	CMD_PROTECT
	CMD_VOTE
	CMD_PULL
	CMD_HEAL
	CMD_POISON
	CMD_HELP
)

type command struct {
	id     CommandID
	client *client
	args   []string
}
