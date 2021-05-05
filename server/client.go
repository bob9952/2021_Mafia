package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
)

type roleType int

const (
	MAFIA roleType = iota
	TOWN
	DOCTOR
	INSPECTOR
	AVENGER
	WITCH
	JESTER
)

type client struct {
	conn        net.Conn
	nick        string
	room        *room
	commands    chan<- command
	role        roleType
	isAlive     bool
	hasVoted    bool
	isProtected bool
	numOfVotes  int
	isHealed    bool
}

func roleTypeToString(rt roleType) string {
	switch rt {
	case TOWN:
		return "TOWN"
	case MAFIA:
		return "MAFIA"
	case DOCTOR:
		return "DOCTOR"
	case INSPECTOR:
		return "INSPECTOR"
	case JESTER:
		return "JESTER"
	case AVENGER:
		return "AVENGER"
	case WITCH:
		return "WITCH"
	}

	return "error"
}

func (c *client) readInput() {
	for {
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			if err == io.EOF {
				c.commands <- command{
					id:     CMD_QUIT,
					client: c,
				}
			} else {
				fmt.Println(err)
			}
			return
		}

		msg = strings.Trim(msg, "\r\n")

		args := strings.Split(msg, " ")

		cmd := strings.TrimSpace(args[0])

		switch cmd {
		case "/join":
			if len(args) <= 1 {
				c.msg(fmt.Sprintf("invalid usage, missing name of the room"))
				break
			}
			c.commands <- command{
				id:     CMD_JOIN,
				client: c,
				args:   args,
			}
		case "/quit":
			c.commands <- command{
				id:     CMD_QUIT,
				client: c,
			}
		case "/start":
			c.commands <- command{
				id:     CMD_START,
				client: c,
			}
		case "/rooms":
			c.commands <- command{
				id:     CMD_ROOMS,
				client: c,
			}
		case "/leave":
			c.commands <- command{
				id:     CMD_LEAVE,
				client: c,
			}
		case "/list":
			c.commands <- command{
				id:     CMD_LIST,
				client: c,
			}
		case "/help":
			c.commands <- command{
				id:     CMD_HELP,
				client: c,
			}
		case "/kill":
			if len(args) != 2 {
				c.msg(fmt.Sprintf("invalid usage, missing name of the player"))
				break
			}
			c.commands <- command{
				id:     CMD_KILL,
				client: c,
				args:   args,
			}
		case "/protect":
			if len(args) != 2 {
				c.msg(fmt.Sprintf("invalid usage, missing name of the player"))
				break
			}
			c.commands <- command{
				id:     CMD_PROTECT,
				client: c,
				args:   args,
			}
		case "/pull":
			if len(args) != 2 {
				c.msg(fmt.Sprintf("invalid usage, missing name of the player"))
				break
			}
			c.commands <- command{
				id:     CMD_PULL,
				client: c,
				args:   args,
			}
		case "/inspect":
			if len(args) != 2 {
				c.msg(fmt.Sprintf("invalid usage, missing name of the player"))
				break
			}
			c.commands <- command{
				id:     CMD_INSPECT,
				client: c,
				args:   args,
			}
		case "/heal":
			if len(args) != 2 {
				c.msg(fmt.Sprintf("invalid usage, missing name of the player"))
				break
			}
			c.commands <- command{
				id:     CMD_HEAL,
				client: c,
				args:   args,
			}
		case "/poison":
			if len(args) != 2 {
				c.msg(fmt.Sprintf("invalid usage, missing name of the player"))
				break
			}
			c.commands <- command{
				id:     CMD_POISON,
				client: c,
				args:   args,
			}
		case "/vote":
			if len(args) != 2 {
				c.msg(fmt.Sprintf("invalid usage, missing name of the player"))
				break
			}
			c.commands <- command{
				id:     CMD_VOTE,
				client: c,
				args:   args,
			}
		default:
			if c.room == nil {
				c.msg(fmt.Sprintf("unknown command: %s", cmd))
			} else if c.room.state == WAITING {
				c.room.broadcast(c, "/text2 " + c.nick + ": " + msg)
				c.msg("/text2 " + c.nick + ": " + msg)
			} else if !c.isAlive {
				c.msg("/text1 You be dead")
				break
			} else if c.room.state == NIGHT && c.role == MAFIA {
				c.msg("/text2 " + c.nick + ": " + msg)
				c.room.mafiaBroadcast(c, "/text2 " + c.nick + ": " + msg)
			} else if c.room.state != NIGHT && c.room.state != VOTE {
				c.msg("/text2 " + c.nick + ": " + msg)
				c.room.broadcast(c, "/text2 " + c.nick + ": " + msg)
			}
		}
	}
}

func (c *client) msg(msg string) {
	arr := make([]byte, 4)
	binary.LittleEndian.PutUint32(arr[0:4], uint32(len(msg)))
	arr = append(arr, []byte(msg+"\n")...)
	_, _ = c.conn.Write(arr)
}
