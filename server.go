package main

import "C"
import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

const MIN_PLAYERS = 6
const MAX_PLAYERS = 10
const CHAT_DURATION = 10

type typeRandom int

const (
	MAFIA1 typeRandom = iota
	MAFIA2
	TOWN1
	TOWN2
	DOCTOR1
	INSPECTOR1
	TOWN3
	JESTER1
	AVENGER1
	WITCH1
)

type server struct {
	rooms    map[string]*room
	commands chan command
	imena    []string
}

func newServer() *server {
	return &server{
		rooms:    make(map[string]*room),
		commands: make(chan command),
		imena:    make([]string, 0),
	}
}

func (s *server) run() {
	for cmd := range s.commands {
		s.commandHandler(&cmd)
	}
}

func (s *server) listUsers(c *client) {
	if c.room != nil {
		for player := range c.room.members {
			c.msg(player + ", ")
		}
	} else {
		c.msg("You're not in a room!")
	}
}

func (s *server) newClient(conn net.Conn) {

	c := &client{
		conn:     conn,
		nick:     "anonymous",
		commands: s.commands,
	}

	s.readUsername(c)

	c.readInput()
}

func (s *server) commandHandler(cmd *command) {
	switch cmd.id {
	case CMD_JOIN:
		s.join(cmd.client, cmd.args[1])
	case CMD_ROOMS:
		s.listRooms(cmd.client)
	case CMD_QUIT:
		s.quit(cmd.client)
	case CMD_LEAVE:
		s.quitCurrentRoom(cmd.client)
	case CMD_START:
		s.startGame(cmd.client)
	case CMD_LIST:
		s.listUsers(cmd.client)
	case CMD_KILL:
		s.Kill(cmd.client, cmd.args[1])
	case CMD_INSPECT:
		s.Inspect(cmd.client, cmd.args[1])
	case CMD_PROTECT:
		s.Protect(cmd.client, cmd.args[1])
	case CMD_VOTE:
		s.Vote(cmd.client, cmd.args[1])
	case CMD_PULL:
		s.Pull(cmd.client, cmd.args[1])
	case CMD_POISON:
		s.Poison(cmd.client, cmd.args[1])
	case CMD_HEAL:
		s.Heal(cmd.client, cmd.args[1])
	}
}

func (s *server) join(c *client, roomName string) {
	r, ok := s.rooms[roomName]
	if !ok {
		r = &room{
			name:    roomName,
			members: make(map[string]*client),
			owner:   c.nick,
			state:   WAITING,
		}
		s.rooms[roomName] = r
	}

	if len(r.members) == MAX_PLAYERS {
		c.msg("This room is full!")
		return
	}

	if r.state != WAITING {
		c.msg("The game is in progress!")
		return
	}

	r.members[c.nick] = c

	s.quitCurrentRoom(c)
	c.room = r

	r.broadcast(c, fmt.Sprintf("%s joined the room", c.nick))

	c.msg(fmt.Sprintf("welcome to %s", roomName))
}

func (s *server) listRooms(c *client) {
	var rooms []string
	for name := range s.rooms {
		rooms = append(rooms, name)
	}

	c.msg(fmt.Sprintf("available rooms: %s", strings.Join(rooms, ", ")))
}

func (s *server) quit(c *client) {
	log.Printf("client has left the chat: %s", c.conn.RemoteAddr().String())

	s.quitCurrentRoom(c)

	if c.conn != nil {
		_ = c.conn.Close()
	}

	for index, player := range s.imena {
		if c.nick == player {
			s.imena = append(s.imena[:index], s.imena[index+1:]...)
		}
	}
}

func (s *server) quitCurrentRoom(c *client) {
	if c.room != nil {
		oldRoom := s.rooms[c.room.name]
		delete(s.rooms[c.room.name].members, c.nick)
		if len(s.rooms[c.room.name].members) == 0 {
			delete(s.rooms, oldRoom.name)
			return
		}
		oldRoom.broadcast(c, fmt.Sprintf("%s has left the room", c.nick))

		if c.room.owner == c.nick {
			for k := range c.room.members {
				c.room.owner = k
				oldRoom.broadcast(c, fmt.Sprintf("%s is the new room owner", k))
				break
			}
		}

		if c.room.state != WAITING {
			c.room.numOfVoters--
			if c.role == TOWN || c.role == JESTER {
				c.room.numOfTown--
			}
		}

		c.room = nil
	}
}

func (s *server) readUsername(c *client) {
	c.msg("Enter username:")
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
		_, found := Find(s.imena, msg)
		if !found {
			c.nick = strings.Trim(msg, "\r\n")
			s.imena = append(s.imena, c.nick)
			break
		}
		c.msg("Username taken, try again!")
	}
}

func (s *server) startGame(c *client) {
	if c.room.owner != c.nick {
		c.msg("You aren't allowed to start the game")
		return
	} else if len(c.room.members) < MIN_PLAYERS {
		c.msg("Cannot start the game with this number of players!")
		return
	}

	roles := make([]typeRandom, 0)
	for i := 0; i < len(c.room.members); i++ {
		roles = append(roles, typeRandom(i))
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(roles), func(i, j int) { roles[i], roles[j] = roles[j], roles[i] })

	c.room.numOfVoters = len(c.room.members)
	c.room.numOfTown = 0

	var mafiaMembers []*client

	i := 0

	for _, player := range c.room.members {
		player.isAlive = true
		switch roles[i] {
		case TOWN1:
			player.role = TOWN
			player.msg("Your role is TOWN!")
			c.room.numOfTown++
		case TOWN2:
			player.role = TOWN
			player.msg("Your role is TOWN!")
			c.room.numOfTown++
		case TOWN3:
			player.role = TOWN
			player.msg("Your role is TOWN!")
			c.room.numOfTown++
		case MAFIA1:
			player.role = MAFIA
			player.msg("Your role is MAFIA!")
			mafiaMembers = append(mafiaMembers, player)
		case MAFIA2:
			player.role = MAFIA
			player.msg("Your role is MAFIA!")
			mafiaMembers = append(mafiaMembers, player)
		case DOCTOR1:
			player.role = DOCTOR
			player.msg("Your role is DOCTOR!")
		case INSPECTOR1:
			player.role = INSPECTOR
			player.msg("Your role is INSPECTOR!")
		case AVENGER1:
			player.role = AVENGER
			player.msg("Your role is AVENGER!")
		case WITCH1:
			player.role = WITCH
			player.msg("Your role is WITCH!")
		case JESTER1:
			player.role = JESTER
			player.msg("Your role is JESTER!")
			c.room.numOfTown++
		}
		i++
	}

	for _, member := range mafiaMembers {
		member.msg("Other mafia members: ")
		for _, member1 := range mafiaMembers {
			if member != member1 {
				member.msg(member1.nick + ", ")
			}
		}
	}

	s.Night(c.room)
}

func (s *server) Night(r *room) {

	r.state = NIGHT

	for _, player := range r.members {
		if player.isAlive {
			switch player.role {
			case TOWN:
				player.msg("You're town, you don't have anything to do")
			case MAFIA:
				player.msg("You're mafia. You can choose who to kill by typing /kill name")
			case DOCTOR:
				player.msg("You're a doctor. You can choose who to protect by typing /protect name")
			case INSPECTOR:
				player.msg("You're an inspector. You can choose who to inspect by typing /inspect name")
			case JESTER:
				player.msg("You're the jester! Don't laugh too loud, the mafia might kill you!")
			case AVENGER:
				player.msg("You're the avenger! You can choose who to pull to his death with /pull name" +
					"or you can skip with /pull skip")
			case WITCH:
				player.msg("You're the witch! You can give someone the posioned apple with /posion name or you" +
					"can give them the elixir of life with /heal name! You can also skip with /heal skip or /poison skip")
			}
		}
	}

	for cmd := range s.commands {
		s.commandHandler(&cmd)
		if r.numOfVotes == r.numOfVoters-r.numOfTown {
			break
		}
	}

	killed := r.whoToKill()

	if killed != r.whoToHeal() && !killed.isHealed {
		killed.isAlive = false
		killed.msg("You have died!")
		r.broadcast(killed, "Last night "+killed.nick+" died, RIP")
		if killed.role == TOWN || killed.role == JESTER {
			r.numOfTown--
		}
		r.numOfVoters--

		if killed.role == AVENGER && r.pulled != nil {
			r.pulled.isAlive = false
			r.pulled.msg("You have been pulled by the avenger!")
			r.broadcast(r.pulled, "Last night "+r.pulled.nick+" was pulled by the avenger, RIP")
			if r.pulled.role == TOWN || r.pulled.role == JESTER {
				r.numOfTown--
			}
			r.numOfVoters--
		}
	} else {
		r.members[r.owner].msg("Last night nobody died!")
		r.broadcast(r.members[r.owner], "Last night nobody died!")
	}

	if r.potionPoison != nil && r.potionPoison != killed {
		r.potionPoison.isAlive = false
		r.potionPoison.msg("You choked on an apple!")
		r.broadcast(r.potionPoison, "Last night "+r.potionPoison.nick+" choked on an apple, RIP")
		if r.potionPoison.role == TOWN || r.potionPoison.role == JESTER {
			r.numOfTown--
		}
		r.numOfVoters--
	}

	r.reset()

	isGameOver := r.gameOver()

	if isGameOver != NOTHING {
		if isGameOver == TOWN_WON {
			r.members[r.owner].msg("Town won!")
			r.broadcast(r.members[r.owner], "Town won!")
		} else {
			r.members[r.owner].msg("Mafia won!")
			r.broadcast(r.members[r.owner], "Mafia won!")
		}
		r.state = WAITING
		return
	}

	s.Day(r)

}

func (s *server) Day(r *room) {
	r.state = DAY
	r.members[r.owner].msg("Good morning!")
	r.broadcast(r.members[r.owner], "Good morning!")

	go func() {
		dayEnded := time.Now().Add(CHAT_DURATION * time.Second)
		for {
			if time.Now().After(dayEnded) {
				break
			}
			time.Sleep(time.Second)
		}
		r.members[r.owner].msg("It's voting time! You can vote by typing /vote name or /vote skip")
		r.broadcast(r.members[r.owner], "It's voting time! You can vote by typing /vote name or /vote skip")
		r.state = VOTE
	}()

	for cmd := range s.commands {
		s.commandHandler(&cmd)
		if r.numOfVotes == r.numOfVoters {
			break
		}
	}

	votekilled := r.whoToVoteKill()

	if votekilled != nil {
		votekilled.isAlive = false
		votekilled.msg("You have died!")
		r.broadcast(votekilled, votekilled.nick+" was hanged on the town square, RIP")
		if votekilled.role == JESTER {
			r.broadcast(votekilled, "You hanged the jester! Why so serious?\nThe jester has won!")
			votekilled.msg("You win, joker!")

			r.reset()
			r.state = WAITING
			return
		}
		if votekilled.role == TOWN {
			r.numOfTown--
		}
		r.numOfVoters--
	} else {
		r.members[r.owner].msg("Nobody was voted out!")
		r.broadcast(r.members[r.owner], "Nobody was voted out!")
	}

	r.reset()

	isGameOver := r.gameOver()

	if isGameOver != NOTHING {
		if isGameOver == TOWN_WON {
			r.members[r.owner].msg("Town won!")
			r.broadcast(r.members[r.owner], "Town won!")
		} else {
			r.members[r.owner].msg("Mafia won!")
			r.broadcast(r.members[r.owner], "Mafia won!")
		}
		r.state = WAITING
		return
	}

	s.Night(r)
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func (s *server) Kill(c *client, victim string) {
	if !c.isAlive {
		c.msg("You be dead")
		return
	}
	if c.room.state != NIGHT {
		c.msg("You can't kill people in broad daylight!")
		return
	}
	if c.role != MAFIA {
		c.msg("You can't kill people, you don't have a gun")
		return
	}
	if c.hasVoted {
		c.msg("You already acted")
		return
	}
	if v, ok := c.room.members[victim]; ok {
		if !v.isAlive {
			c.msg("You can't kill the dead!")
			return
		}
		v.numOfVotes++
		c.msg("You voted to kill " + v.nick)
	} else {
		c.msg("You can't kill the nonexistent!")
		return
	}
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Inspect(c *client, victim string) {
	if !c.isAlive {
		c.msg("You be dead")
		return
	}
	if c.room.state != NIGHT {
		c.msg("You can't inspect people during the day!")
		return
	}
	if c.role != INSPECTOR {
		c.msg("You can't inspect people, you're not an inspector'")
		return
	}
	if c.hasVoted {
		c.msg("You already acted")
		return
	}
	if v, ok := c.room.members[victim]; ok {
		if !v.isAlive {
			c.msg("You can't inspect the dead!")
			return
		}
		c.msg(v.nick + "'s role is " + roleTypeToString(v.role))
	} else {
		c.msg("You can't inspect the nonexistent!")
		return
	}
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Protect(c *client, victim string) {
	if !c.isAlive {
		c.msg("You be dead")
		return
	}

	if c.room.state != NIGHT {
		c.msg("You can't heal people during the day!")
		return
	}
	if c.role != DOCTOR {
		c.msg("You can't heal people, you're not a doctor")
		return
	}
	if c.hasVoted {
		c.msg("You already acted")
		return
	}
	if v, ok := c.room.members[victim]; ok {
		if !v.isAlive {
			c.msg("You can't heal the dead! You're not Jesus")
			return
		}
		if c.room.lastHealed == v {
			c.msg("You can't heal the same person two nights in a row!")
			return
		}
		v.isProtected = true
		c.room.lastHealed = v
		c.msg("You protected " + v.nick)
	} else {
		c.msg("You can't protect the nonexistent!")
		return
	}
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Vote(c *client, victim string) {
	if !c.isAlive {
		c.msg("You be dead")
		return
	}

	if c.room.state != VOTE {
		c.msg("You can't vote right now!")
		return
	}
	if c.hasVoted {
		c.msg("You already voted")
		return
	}
	if victim != "skip" {
		if v, ok := c.room.members[victim]; ok {
			if !v.isAlive {
				c.msg("You can't vote to hang the dead!")
				return
			}
			c.msg("You voted for " + v.nick)
			v.numOfVotes++
		} else {
			c.msg("You can't vote for the nonexistent!")
			return
		}
	}
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Pull(c *client, victim string) {
	if !c.isAlive {
		c.msg("You be dead")
		return
	}

	if c.room.state != NIGHT {
		c.msg("You can't pull people during the day! It's rude!")
		return
	}
	if c.role != AVENGER {
		c.msg("You can't pull people, you're not a marvel superhero")
		return
	}
	if c.hasVoted {
		c.msg("You already acted")
		return
	}
	if victim != "skip" {
		if v, ok := c.room.members[victim]; ok {
			if !v.isAlive {
				c.msg("You can't pull the dead!")
				return
			}
			c.room.pulled = v
			c.msg("You pulled " + v.nick)
		} else {
			c.msg("You can't pull the nonexistent!")
			return
		}
	}
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Poison(c *client, victim string) {
	if !c.isAlive {
		c.msg("You be dead")
		return
	}
	if c.room.state != NIGHT {
		c.msg("You can't posion people during the day! It's rude!")
		return
	}
	if c.role != WITCH {
		c.msg("You can't poison people, you're not Bill Gates")
		return
	}
	if c.hasVoted {
		c.msg("You already acted")
		return
	}
	if c.room.potionPoison != nil {
		c.msg("You're all out of apples!")
		return
	}
	if victim != "skip" {
		if v, ok := c.room.members[victim]; ok {
			if !v.isAlive {
				c.msg("You can't poison the dead!")
				return
			}
			c.room.potionPoison = v
			c.msg("You posioned " + v.nick)
		} else {
			c.msg("You can't posion the nonexistent!")
			return
		}
	}
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Heal(c *client, victim string) {
	if !c.isAlive {
		c.msg("You be dead")
		return
	}
	if c.room.state != NIGHT {
		c.msg("You can't heal people during the day!")
		return
	}
	if c.role != WITCH {
		c.msg("You can't heal people, you're not Bill Gates")
		return
	}
	if c.hasVoted {
		c.msg("You already acted")
		return
	}
	if c.room.potionHeal != nil {
		c.msg("You're all out of elixirs!")
		return
	}
	if victim != "skip" {
		if v, ok := c.room.members[victim]; ok {
			if !v.isAlive {
				c.msg("You can't heal the dead! You're not Jesus")
				return
			}
			c.room.potionHeal = v
			v.isHealed = true
			c.msg("You healed " + v.nick)
		} else {
			c.msg("You can't heal the nonexistent!")
			return
		}
	}
	c.hasVoted = true
	c.room.numOfVotes++
}
