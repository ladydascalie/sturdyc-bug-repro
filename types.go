package main

import "errors"

// Direction defines the direction of the leaderboard entries will be sorted in,
// either ascending or descending.
// Ascending means that the lowest score will be at the top of the leaderboard.
// Descending means that the highest score will be at the top of the leaderboard.
type Direction string

func (d Direction) String() string { return string(d) }

const (
	Ascending  Direction = "ascending"
	Descending Direction = "descending"
)

func (d Direction) MarshalText() ([]byte, error) {
	return []byte(d), nil
}

func (d *Direction) UnmarshalText(text []byte) error {
	switch string(text) {
	case "ascending":
		*d = Ascending
	case "descending":
		*d = Descending
	default:
		return errors.New("invalid direction")
	}
	return nil
}

const (
	Player  Type = "player"
	Generic Type = "generic"
)

// Type defines what kind of leaderboard you are using, player or generic.
// The difference lies in what value is used as the member id.
//
// In the case of a player leaderboard, the member id is the player id.
// In the case of a generic leaderboard, the member id is some string provided by the user.
type Type string

func (t Type) String() string { return string(t) }

func (t Type) MarshalText() ([]byte, error) {
	return []byte(t), nil
}

func (t *Type) UnmarshalText(text []byte) error {
	switch string(text) {
	case "player":
		*t = Player
	case "generic":
		*t = Generic
	default:
		return errors.New("invalid leaderboard type")
	}
	return nil
}
