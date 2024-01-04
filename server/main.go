package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid/v4"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
)

var (
	host      = "http://localhost:7880"
	apiKey    = "devkey"
	apiSecret = "secret"
)

type CreateRoomResponse struct {
	RoomName string `json:"roomName"`
	Token    string `json:"token"`
}

type CounterIncrementRequest struct {
	RoomName string `json:"roomName"`
}

type RoomMetadata struct {
	Counter int `json:"counter"`
}

func createRoomMetadata(counter int) []byte {
	metadata := RoomMetadata{
		Counter: counter,
	}
	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		panic(err)
	}
	return metadataJson
}

func main() {
	if os.Getenv("LIVEKIT_HOST") != "" {
		host = os.Getenv("LIVEKIT_HOST")
	}

	log.Printf("Using host: %s", host)

	roomClient := lksdk.NewRoomServiceClient(host, apiKey, apiSecret)

	globalCounter := 0

	r := gin.Default()

	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, you've reached the homepage!")
	})

	r.POST("/counter-increment", func(c *gin.Context) {
		var req CounterIncrementRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.String(http.StatusBadRequest, "Error: %v", err)
			return
		}

		rooms, err := roomClient.ListRooms(context.Background(), &livekit.ListRoomsRequest{
			Names: []string{req.RoomName},
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "Error listing rooms: %v", err)
			return
		}

		if len(rooms.Rooms) != 1 {
			c.String(http.StatusBadRequest, "Error: room not found")
			return
		}

		room := rooms.Rooms[0]
		metadata := room.Metadata
		var roomMetadata RoomMetadata
		err = json.Unmarshal([]byte(metadata), &roomMetadata)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error unmarshalling metadata: %v", err)
			return
		}

		log.Printf("%s counter: %d -> %d", room.Name, roomMetadata.Counter, roomMetadata.Counter+1)

		roomMetadata.Counter++
		room.Metadata = string(createRoomMetadata(roomMetadata.Counter))

		_, err = roomClient.UpdateRoomMetadata(context.Background(), &livekit.UpdateRoomMetadataRequest{
			Room:     room.Name,
			Metadata: room.Metadata,
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "Error updating room metadata: %v", err)
			return
		}

		// Just 200 OK
		c.String(http.StatusOK, "")
	})

	r.POST("/create-room", func(c *gin.Context) {
		roomName := "Room " + shortuuid.New()
		userIdentity := "User " + shortuuid.New()
		userName := "Name " + shortuuid.New()
		room, err := roomClient.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
			Name:     roomName,
			Metadata: string(createRoomMetadata(globalCounter)),
		})
		globalCounter++
		if err != nil {
			c.String(http.StatusInternalServerError, "Error creating room: %v", err)
			return
		}

		// generate token
		at := auth.NewAccessToken(apiKey, apiSecret)
		canUpdateOwnMetadata := true
		grant := &auth.VideoGrant{
			RoomJoin:             true,
			Room:                 room.Name,
			CanUpdateOwnMetadata: &canUpdateOwnMetadata,
		}
		at.AddGrant(grant).
			SetIdentity(userIdentity).
			SetName(userName).
			SetValidFor(60 * time.Second) // 60 seconds

		token, err := at.ToJWT()
		if err != nil {
			c.String(http.StatusInternalServerError, "Error generating token: %v", err)
			return
		}

		c.JSON(http.StatusOK, CreateRoomResponse{
			RoomName: roomName,
			Token:    token,
		})
	})

	r.Run(":8101")
}
