package objmanager

import (
	"time"

	"github.com/giongto35/gowog/server/game/mappkg"
	"github.com/giongto35/gowog/server/game/playerpkg"
	"github.com/giongto35/gowog/server/game/shootpkg"
)

type ObjectManager interface {
	RegisterPlayer(clientID int32, name string) playerpkg.Player
	RegisterShoot(player playerpkg.Player, x float32, y float32, dx float32, dy float32, startTime time.Time) shootpkg.Shoot
	RemovePlayerByID(id int32)
	RemovePlayerByClientID(id int32)

	//RangePlayers(f func(player playerpkg.Player))
	GetPlayers() []playerpkg.Player
	GetMap() mappkg.Map
	GetPlayerByID(id int32) (playerpkg.Player, bool)
	MovePlayer(player playerpkg.Player, dx float32, dy float32, speed float32, timeElapsed float32)

	Update()
}
