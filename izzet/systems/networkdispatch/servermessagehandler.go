package networkdispatch

import (
	"fmt"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/knetwork"
	"github.com/kkevinchou/izzet/izzet/managers/player"
	"github.com/kkevinchou/izzet/izzet/utils/entityutils"
	"github.com/kkevinchou/kitolib/network"
)

func serverMessageHandler(world World, message *network.Message) {
	player := world.GetPlayerByID(message.SenderID)
	singleton := world.GetSingleton()
	if player == nil {
		fmt.Println(fmt.Errorf("failed to find player with id %d", message.SenderID))
		return
	}

	if message.MessageType == knetwork.MessageTypeCreatePlayer {
		handleCreatePlayer(player, message, world)
	} else if message.MessageType == knetwork.MessageTypeInput {
		inputMessage := knetwork.InputMessage{}
		err := network.DeserializeBody(message, &inputMessage)
		if err != nil {
			panic(err)
		}

		singleton.InputBuffer.PushInput(world.CommandFrame(), message.CommandFrame, message.SenderID, time.Now(), &inputMessage)
	} else if message.MessageType == knetwork.MessageTypePing {
		var pingMessage knetwork.PingMessage
		err := network.DeserializeBody(message, &pingMessage)
		if err != nil {
			fmt.Printf("error deserializing ping body %s\n", err)
		}
		msg := knetwork.AckPingMessage{PingSendTime: pingMessage.SendTime}
		err = player.Client.SendMessage(knetwork.MessageTypeAckPing, msg)
		if err != nil {
			fmt.Printf("error sending ackping message %s\n", err)
		}
	} else if message.MessageType == knetwork.MessageTypeRPC {
		var rpcMessage knetwork.RPCMessage
		err := network.DeserializeBody(message, &rpcMessage)
		if err != nil {
			fmt.Printf("error deserializing ping body %s\n", err)
		}
		world.GetEventBroker().Broadcast(&events.RPCEvent{PlayerID: message.SenderID, Command: rpcMessage.Command})
	} else {
		fmt.Println("unknown message type:", message.MessageType, string(message.Body))
	}
}

// TODO: in the future this should be handled by some other system via an event
func handleCreatePlayer(player *player.Player, message *network.Message, world World) {
	playerID := message.SenderID

	bob := entities.NewBob()
	player.EntityID = bob.ID

	cc := bob.ComponentContainer

	camera := entities.NewThirdPersonCamera(mgl64.Vec3{}, mgl64.Vec2{0, 0}, player.ID, player.EntityID)
	cameraComponentContainer := camera.GetComponentContainer()
	fmt.Println("Server camera initialized at position", cameraComponentContainer.TransformComponent.Position)

	cc.ThirdPersonControllerComponent.CameraID = camera.GetID()

	world.RegisterEntities([]entities.Entity{bob, camera})
	fmt.Println("Created and registered a new bob with id", bob.ID)

	snapshots := map[int]knetwork.EntitySnapshot{}
	for _, entity := range world.QueryEntity(components.ComponentFlagNetwork) {
		if entity.GetID() == bob.ID {
			continue
		}
		snapshots[entity.GetID()] = entityutils.ConstructEntitySnapshot(entity)
	}

	ack := &knetwork.AckCreatePlayerMessage{
		PlayerID:    playerID,
		EntityID:    bob.ID,
		CameraID:    camera.ID,
		Position:    cc.TransformComponent.Position,
		Orientation: cc.TransformComponent.Orientation,
		Entities:    snapshots,
	}

	player.Client.SendMessage(network.MessageTypeAckCreatePlayer, ack)
	fmt.Println("Sent entity ack creation message")
}
