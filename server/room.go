package main

type gameState int
type outcome int

const (
	DAY gameState = iota
	NIGHT
	VOTE
	WAITING
)

const (
	NOTHING outcome = iota
	TOWN_WON
	MAFIA_WON
)

type room struct {
	name         string
	members      map[string]*client
	owner        string
	state        gameState
	numOfVotes   int
	numOfVoters  int
	numOfTown    int
	lastHealed   *client
	pulled       *client
	potionHeal   *client
	potionPoison *client
}

func (r *room) broadcast(sender *client, msg string) {
	for name, m := range r.members {
		if sender.nick != name {
			m.msg(msg)
		}
	}
}

func (r *room) mafiaBroadcast(sender *client, msg string) {
	for name, m := range r.members {
		if sender.nick != name && m.role == MAFIA {
			m.msg(msg)
		}
	}
}

func (r *room) whoToKill() *client {
	var max = 0
	var toKill *client
	for _, player := range r.members {
		if player.numOfVotes > max {
			max = player.numOfVotes
			toKill = player
		}
	}
	return toKill
}

func (r *room) whoToVoteKill() *client {
	var max = 0
	var toKill *client
	for _, player := range r.members {
		if player.numOfVotes > max {
			max = player.numOfVotes
			toKill = player
		} else if player.numOfVotes == max {
			toKill = nil
		}
	}
	return toKill
}

func (r *room) whoToHeal() *client {
	for _, player := range r.members {
		if player.isProtected {
			return player
		}
	}
	return nil
}

func (r *room) reset() {
	for _, player := range r.members {
		player.numOfVotes = 0
		player.hasVoted = false
		player.isProtected = false
		player.isHealed = false
	}
	r.numOfVotes = 0
}

func (r *room) gameOver() outcome {
	var town int
	var mafia int
	for _, player := range r.members {
		if player.isAlive {
			if player.role == MAFIA {
				mafia++
			} else {
				town++
			}
		}
	}
	if mafia == 0 {
		return TOWN_WON
	}
	if mafia >= town {
		return MAFIA_WON
	}

	return NOTHING
}
