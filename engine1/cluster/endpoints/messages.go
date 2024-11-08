package endpoints

import "github.com/njtc406/chaosengine/engine1/actor"

type EndpointTerminatedEvent struct {
	Address string
}

type EndpointConnectedEvent struct {
	Address string
}

type remoteWatch struct {
	Watcher *actor.PID
	Watchee *actor.PID
}

type remoteUnwatch struct {
	Watcher *actor.PID
	Watchee *actor.PID
}

type remoteDeliver struct {
	header       actor.ReadonlyMessageHeader
	message      interface{}
	target       *actor.PID
	sender       *actor.PID
	serializerID int32
}

type remoteTerminate struct {
	Watcher *actor.PID
	Watchee *actor.PID
}

type JsonMessage struct {
	TypeName string
	Json     string
}

type (
	// Ping is message sent by the actor system to probe an actor is started.
	Ping struct{}

	// Pong is response for ping.
	Pong struct{}
)
