package main

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

var rolesPool = []roleType{
	MAFIA,
	MAFIA,
	TOWN,
	TOWN,
	DOCTOR,
	INSPECTOR,
	TOWN,
	JESTER,
	AVENGER,
	WITCH,
}

type server struct {
	rooms    map[string]*room
	commands chan command
	names    []string
}

func newServer() *server {
	return &server{
		rooms:    make(map[string]*room),
		commands: make(chan command),
		names:    make([]string, 0),
	}
}

func (s *server) run() {
	for cmd := range s.commands {
		activeRoom := s.commandHandler(&cmd)
		activeRoom.dayNightHandler()
	}
}

func (s *server) listUsers(c *client) {
	if c.room != nil {
		players := ""
		for player := range c.room.members {
			players += player + "\n"
		}
		c.msg("/text2 " + players)
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

func (s *server) commandHandler(cmd *command) *room {
	activeRoom := cmd.client.room
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
	case CMD_HELP:
		s.Help(cmd.client)
	}
	return activeRoom
}

func (r *room) dayNightHandler() {
	if r == nil {
		return
	}
	if r.state == NIGHT {
		r.night()
	} else if r.state == VOTE {
		r.day()
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

	if r == c.room {
		c.msg("You're already in this room!")
		return
	}

	if len(r.members) == MAX_PLAYERS {
		c.msg("/text1 This room is full!")
		return
	}

	if r.state != WAITING {
		c.msg("/text1 The game is in progress!")
		return
	}

	r.members[c.nick] = c

	s.quitCurrentRoom(c)
	c.room = r

	players := ""
	for player := range c.room.members {
		players += player + " "
	}
	players = strings.TrimSuffix(players, " ")
	r.broadcast(c, fmt.Sprintf("/join "+c.room.name+" "+players))

	c.msg(fmt.Sprintf("/join " + c.room.name + " " + players))
}

func (s *server) listRooms(c *client) {
	var rooms []string
	for name := range s.rooms {
		rooms = append(rooms, name)
	}

	c.msg(fmt.Sprintf("/text2 available rooms: %s", strings.Join(rooms, ", ")))
}

func (s *server) quit(c *client) {
	log.Printf("client has left the chat: %s", c.conn.RemoteAddr().String())

	s.quitCurrentRoom(c)

	if c.conn != nil {
		_ = c.conn.Close()
	}

	for index, player := range s.names {
		if c.nick == player {
			s.names = append(s.names[:index], s.names[index+1:]...)
		}
	}
}

func (s *server) quitCurrentRoom(c *client) {
	if c.room != nil {

		c.msg("/open " + c.nick + " Welcome to Mafia Server.\nUse /help to see all available commands")

		oldRoom := s.rooms[c.room.name]
		delete(s.rooms[c.room.name].members, c.nick)
		if len(s.rooms[c.room.name].members) == 0 {
			delete(s.rooms, oldRoom.name)
			c.room = nil
			return
		}

		if c.room.owner == c.nick {
			for k := range c.room.members {
				c.room.owner = k
				oldRoom.broadcast(c, fmt.Sprintf("/owner %s is the new room owner", k))
				break
			}
		}

		players := ""
		for player := range c.room.members {
			players += player + " "
		}
		players = strings.TrimSuffix(players, " ")
		c.room.broadcast(c, fmt.Sprintf("/leave "+c.nick+" "+players))

		if c.room.state != WAITING && c.isAlive == true {
			c.room.numOfVoters--
			if c.hasVoted == true {
				c.room.numOfVotes--
			}
			if c.role == TOWN || c.role == JESTER {
				c.room.numOfTown--
			}
		}

		c.room = nil
	}
}

func (s *server) readUsername(c *client) {
	for {
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.quit(c)
			} else {
				fmt.Println(err)
			}
			return
		}
		name := strings.Trim(msg, "\r\n")
		spaces := strings.Contains(name, " ")
		if spaces {
			c.msg("/firstws No whitespaces in name allowed, try again!")
			continue
		}
		_, found := Find(s.names, name)
		if found {
			c.msg("/firstname Username taken, try again!")
			continue
		}
		c.nick = strings.Trim(msg, "\r\n")
		s.names = append(s.names, c.nick)
		c.msg("/open " + c.nick + " Welcome to Mafia Server.\nUse /help to see all available commands")
		break
	}
}

func (r *room) sendRoleDescriptions() {
	for _, player := range r.members {
		if player.isAlive {
			switch player.role {
			case TOWN:
				player.msg("/night You're town, you don't have anything to do")
			case MAFIA:
				player.msg("/night You're mafia. You can choose who to kill by clicking the button below their picture")
			case DOCTOR:
				player.msg("/night You're a doctor. You can choose who to protect by clicking the button below their picture")
			case INSPECTOR:
				player.msg("/night You're an inspector. You can choose who to inspect by by clicking the button below their picture")
			case JESTER:
				player.msg("/night You're the jester! Don't laugh too loud, the mafia might kill you!")
			case AVENGER:
				player.msg("/night You're the avenger! You can choose who to pull to his by clicking the button below their picture or you can skip")
			case WITCH:
				player.msg("/night You're the witch! You can give someone the poisoned apple by clicking the poison button below their picture" +
					"or you can give them the elixir of life by clicking the heal button! You can also skip if you want to")
			}
		} else {
			player.msg("/night You be dead")
		}
	}
}

func (s *server) startGame(c *client) {
	if c.room != nil {
		if c.room.owner != c.nick {
			c.msg("/text1 You aren't allowed to start the game")
			return
		} else if len(c.room.members) < MIN_PLAYERS {
			c.msg("/text1 Cannot start the game with this number of players!")
			return
		}

		roles := make([]roleType, len(c.room.members))
		copy(roles, rolesPool)

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(roles), func(i, j int) { roles[i], roles[j] = roles[j], roles[i] })

		c.room.numOfVoters = len(c.room.members)
		c.room.numOfTown = 0

		var mafiaMembers []*client

		i := 0

		for _, player := range c.room.members {
			player.isAlive = true
			player.isHealed = false;
			player.isProtected = false;
			player.hasVoted = false;
			player.role = roles[i]
			switch roles[i] {
			case TOWN:
				player.msg("/role town Your role is TOWN!")
				c.room.numOfTown++
			case MAFIA:
				player.msg("/role mafia Your role is MAFIA!")
				mafiaMembers = append(mafiaMembers, player)
			case DOCTOR:
				player.msg("/role doctor Your role is DOCTOR!")
			case INSPECTOR:
				player.msg("/role inspector Your role is INSPECTOR!")
			case AVENGER:
				player.msg("/role avenger Your role is AVENGER!")
			case WITCH:
				player.msg("/role witch Your role is WITCH!")
			case JESTER:
				player.msg("/role jester Your role is JESTER!")
				c.room.numOfTown++
			}
			i++
		}

		for _, member := range mafiaMembers {
			member.msg("/text2 Other mafia members: ")
			for _, member1 := range mafiaMembers {
				if member != member1 {
					member.msg("/text2 " + member1.nick + ", ")
				}
			}
		}

		c.room.lastHealed = nil
		c.room.pulled = nil
		c.room.potionHeal = nil
		c.room.potionPoison = nil

		c.room.state = NIGHT
		c.room.sendRoleDescriptions()
	} else {
		c.msg("/text1 You're not in a room!")
	}
}

func (r *room) night() {

	if r.numOfVotes != r.numOfVoters-r.numOfTown {
		return
	}

	killed := r.whoToKill()

	if killed != r.whoToHeal() && !killed.isHealed {
		killed.isAlive = false
		killed.msg("/dead " + killed.nick + " You have died!")
		r.broadcast(killed, "/dead " + killed.nick + " Last night " + killed.nick + " died, RIP")
		if killed.role == TOWN || killed.role == JESTER {
			r.numOfTown--
		}
		r.numOfVoters--

		if killed.role == AVENGER && r.pulled != nil {
			r.pulled.isAlive = false
			r.pulled.msg("/dead " + r.pulled.nick + " You have been pulled by the avenger!")
			r.broadcast(r.pulled, "/dead " + r.pulled.nick + " Last night " + r.pulled.nick + " was pulled by the avenger, RIP")
			if r.pulled.role == TOWN || r.pulled.role == JESTER {
				r.numOfTown--
			}
			r.numOfVoters--
		}
	} else {
		r.members[r.owner].msg("/text2 Last night nobody died!")
		r.broadcast(r.members[r.owner], "/text2 Last night nobody died!")
	}

	if r.potionPoison != nil && r.potionPoison != killed && r.potionPoison.isAlive {
		r.potionPoison.isAlive = false
		r.potionPoison.msg("/dead " + r.potionPoison.nick + " You choked on an apple!")
		r.broadcast(r.potionPoison, "/dead " + r.potionPoison.nick + " Last night " + r.potionPoison.nick + " choked on an apple, RIP")
		if r.potionPoison.role == TOWN || r.potionPoison.role == JESTER {
			r.numOfTown--
		}
		r.numOfVoters--
	}

	r.reset()

	isGameOver := r.gameOver()

	players := ""
	for player := range r.members {
		players += player + " "
	}
	players = strings.TrimSuffix(players, " ")

	if isGameOver != NOTHING {
		if isGameOver == TOWN_WON {
			r.members[r.owner].msg("/join " + r.name + " " + players)
			r.members[r.owner].msg("/won Town won!")
			r.broadcast(r.members[r.owner], "/join " + r.name + " " + players)
			r.broadcast(r.members[r.owner], "/won Town won!")
		} else {
			r.members[r.owner].msg("/join " + r.name + " " + players)
			r.members[r.owner].msg("/won Mafia won!")
			r.broadcast(r.members[r.owner], "/join " + r.name + " " + players)
			r.broadcast(r.members[r.owner], "/won Mafia won!")
		}
		r.state = WAITING
		return
	}

	r.state = DAY
	r.members[r.owner].msg("/day Good morning!")
	r.broadcast(r.members[r.owner], "/day Good morning!")

	go func() {
		dayEnded := time.Now().Add(CHAT_DURATION * time.Second)
		for {
			if time.Now().After(dayEnded) {
				break
			}
			time.Sleep(time.Second)
		}
		r.members[r.owner].msg("/vote It's voting time! You can vote by clicking the appropriate button")
		r.broadcast(r.members[r.owner], "/vote It's voting time! You can vote by clicking the appropriate button")
		r.state = VOTE
	}()
}

func (r *room) day() {

	if r.numOfVotes != r.numOfVoters {
		return
	}

	votekilled := r.whoToVoteKill()

	players := ""
	for player := range r.members {
		players += player + " "
	}
	players = strings.TrimSuffix(players, " ")

	if votekilled != nil {
		votekilled.isAlive = false
		votekilled.msg("/dead " + votekilled.nick + " You have died!")
		r.broadcast(votekilled, "/dead " + votekilled.nick + " " + votekilled.nick + " was hanged on the town square, RIP")
		if votekilled.role == JESTER {
			r.broadcast(votekilled, "/join " + r.name + " " + players)
			r.broadcast(votekilled, "/won You hanged the jester! Why so serious?\nThe jester has won!")
			votekilled.msg("/join " + r.name + " " + players)
			votekilled.msg("/won You win, joker!")

			r.reset()
			r.state = WAITING
			return
		}
		if votekilled.role == TOWN {
			r.numOfTown--
		}
		r.numOfVoters--
	} else {
		r.members[r.owner].msg("/text2 Nobody was voted out!")
		r.broadcast(r.members[r.owner], "/text2 Nobody was voted out!")
	}

	r.reset()

	isGameOver := r.gameOver()

	if isGameOver != NOTHING {
		if isGameOver == TOWN_WON {
			r.members[r.owner].msg("/join " + r.name + " " + players)
			r.members[r.owner].msg("/won Town won!")
			r.broadcast(r.members[r.owner], "/join "+r.name+" "+players)
			r.broadcast(r.members[r.owner], "/won Town won!")
		} else {
			r.members[r.owner].msg("/join " + r.name + " " + players)
			r.members[r.owner].msg("/won Mafia won!")
			r.broadcast(r.members[r.owner], "/join "+r.name+" "+players)
			r.broadcast(r.members[r.owner], "/won Mafia won!")
		}
		r.state = WAITING
		return
	}

	r.state = NIGHT
	r.sendRoleDescriptions()
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
		c.msg("/text1 You be dead")
		return
	}
	if c.room.state != NIGHT {
		c.msg("/text1 You can't kill people in broad daylight!")
		return
	}
	if c.role != MAFIA {
		c.msg("/text1 You can't kill people, you don't have a gun")
		return
	}
	if c.hasVoted {
		c.msg("/text1 You already acted")
		return
	}
	if v, ok := c.room.members[victim]; ok {
		if !v.isAlive {
			c.msg("/text1 You can't kill the dead!")
			return
		}
		v.numOfVotes++
		c.msg("/text1 You voted to kill " + v.nick)
	} else {
		c.msg("/text1 You can't kill the nonexistent!")
		return
	}
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Inspect(c *client, victim string) {
	if !c.isAlive {
		c.msg("/text1 You be dead")
		return
	}
	if c.room.state != NIGHT {
		c.msg("/text1 You can't inspect people during the day!")
		return
	}
	if c.role != INSPECTOR {
		c.msg("/text1 You can't inspect people, you're not an inspector'")
		return
	}
	if c.hasVoted {
		c.msg("/text1 You already acted")
		return
	}
	if v, ok := c.room.members[victim]; ok {
		if !v.isAlive {
			c.msg("/text1 You can't inspect the dead!")
			return
		}
		c.msg("/text2 " + v.nick + "'s role is " + roleTypeToString(v.role))
	} else {
		c.msg("/text1 You can't inspect the nonexistent!")
		return
	}
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Protect(c *client, victim string) {
	if !c.isAlive {
		c.msg("/text1 You be dead")
		return
	}

	if c.room.state != NIGHT {
		c.msg("/text1 You can't heal people during the day!")
		return
	}
	if c.role != DOCTOR {
		c.msg("/text1 You can't heal people, you're not a doctor")
		return
	}
	if c.hasVoted {
		c.msg("/text1 You already acted")
		return
	}
	if v, ok := c.room.members[victim]; ok {
		if !v.isAlive {
			c.msg("/text1 You can't heal the dead! You're not Jesus")
			return
		}
		if c.room.lastHealed == v {
			c.msg("/text1 You can't heal the same person two nights in a row!")
			return
		}
		v.isProtected = true
		c.room.lastHealed = v
		c.msg("/text1 You protected " + v.nick)
	} else {
		c.msg("/text1 You can't protect the nonexistent!")
		return
	}
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Vote(c *client, victim string) {
	if !c.isAlive {
		c.msg("/text1 You be dead")
		return
	}

	if c.room.state != VOTE {
		c.msg("/text1 You can't vote right now!")
		return
	}
	if c.hasVoted {
		c.msg("/text1 You already voted")
		return
	}
	if victim != "skip" {
		if v, ok := c.room.members[victim]; ok {
			if !v.isAlive {
				c.msg("/text1 You can't vote to hang the dead!")
				return
			}
			c.msg("/text1 You voted for " + v.nick)
			v.numOfVotes++
		} else {
			c.msg("/text1 You can't vote for the nonexistent!")
			return
		}
	}
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Pull(c *client, victim string) {
	if !c.isAlive {
		c.msg("/text1 You be dead")
		return
	}

	if c.room.state != NIGHT {
		c.msg("/text1 You can't pull people during the day! It's rude!")
		return
	}
	if c.role != AVENGER {
		c.msg("/text1 You can't pull people, you're not a marvel superhero")
		return
	}
	if c.hasVoted {
		c.msg("/text1 You already acted")
		return
	}
	if victim != "skip" {
		if v, ok := c.room.members[victim]; ok {
			if !v.isAlive {
				c.msg("/text1 You can't pull the dead!")
				return
			}
			c.room.pulled = v
			c.msg("/text1 You pulled " + v.nick)
		} else {
			c.msg("/text1 You can't pull the nonexistent!")
			return
		}
	}
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Poison(c *client, victim string) {
	if !c.isAlive {
		c.msg("/text1 You be dead")
		return
	}
	if c.room.state != NIGHT {
		c.msg("/text1 You can't poison people during the day! It's rude!")
		return
	}
	if c.role != WITCH {
		c.msg("/text1 You can't poison people, you're not Bill Gates")
		return
	}
	if c.hasVoted {
		c.msg("/text1 You already acted")
		return
	}
	if victim == "skip" {
		c.hasVoted = true
		c.room.numOfVotes++
		return
	}
	if c.room.potionPoison != nil {
		c.msg("/text1 You're all out of apples!")
		return
	}
	if v, ok := c.room.members[victim]; ok {
		if !v.isAlive {
			c.msg("/text1 You can't poison the dead!")
			return
		}
		c.room.potionPoison = v
		c.msg("/text1 You poisoned " + v.nick)
	} else {
		c.msg("/text1 You can't poison the nonexistent!")
		return
	}
		
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Heal(c *client, victim string) {
	if !c.isAlive {
		c.msg("/text1 You be dead")
		return
	}
	if c.room.state != NIGHT {
		c.msg("/text1 You can't heal people during the day!")
		return
	}
	if c.role != WITCH {
		c.msg("/text1 You can't heal people, you're not Bill Gates")
		return
	}
	if c.hasVoted {
		c.msg("/text1 You already acted")
		return
	}
	if victim == "skip" {
		c.hasVoted = true
		c.room.numOfVotes++
		return
	}
	if c.room.potionHeal != nil {
		c.msg("/text1 You're all out of elixirs!")
		return
	}
	if v, ok := c.room.members[victim]; ok {
		if !v.isAlive {
			c.msg("/text1 You can't heal the dead! You're not Jesus")
			return
		}
		c.room.potionHeal = v
		v.isHealed = true
		c.msg("/text1 You healed " + v.nick)
	} else {
		c.msg("/text1 You can't heal the nonexistent!")
		return
	}
	
	c.hasVoted = true
	c.room.numOfVotes++
}

func (s *server) Help(c *client) {
	c.msg("/text2 \nAvailable commands:\n/rooms - list all rooms" + "\n" +
		"/join name - join existing room or create it" + "\n" +
		"/leave - leave current room" + "\n" +
		"/quit - quit game" + "\n" +
		"/list - list all users in current room" + "\n" +
		"/start - starts the game\n")
}
