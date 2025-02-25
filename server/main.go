package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
  "strconv"
	"encoding/json"
  "github.com/google/uuid"
)

// userMap := make(map[string]string)

// var voteMap = make(map[string]string)


var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// var connections = make(map[*websocket.Conn]bool)

// TODO  reset votes button, move voting logic out of websocket function and into normal response

type Room struct {
  id string
  connections map[*websocket.Conn]Connection
  voteMap map[string]string
  showVotes bool
}

type Connection struct {
  isAdmin bool
}

var rooms = make(map[string]Room)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
  room := getRoom(r)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()


  connection := Connection{
    isAdmin: isAdmin(r) ,
  } 

	room.connections[conn] = connection

	defer delete(room.connections, conn)

	for {
		messageType, msg, err := conn.ReadMessage()
    log.Println("messagetype: ", messageType)
		if err != nil {
			log.Println("Read error:", err)
			break
		}
    var jsonData map[string]any
    err = json.Unmarshal(msg, &jsonData)
    if err != nil {
      fmt.Println("Error decoding JSON:", err)
      return
    }

    
    
		name, okName := jsonData["name"].(string) // Type assertion for "name"
		vote, okVote := jsonData["vote"].(string) // Type assertion for "vote"

		if !okName || !okVote {
			fmt.Println("Invalid data format")
			continue
		}

    room.voteMap[name] = vote
    // updatedVotesHTML := renderVoteDiv()


		// Echo the message back
    for conn := range room.connections {
      vote_div := renderVoteDiv(room) //TODO move this out
      err := conn.WriteMessage(messageType, []byte(vote_div))

      if err != nil {
          log.Println("Write error:", err)
          conn.Close()
          delete(room.connections, conn)
      }
    }
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./site/"))) // Serve HTML files
  http.HandleFunc("/join-room", joinRoom)
  http.HandleFunc("/create-room", createRoom)
  http.HandleFunc("/reveal-votes", revealVotesHandler)
  // http.HandleFunc("/reset-votes", resetVotesHandler)
  // http.HandleFunc("/refresh", refreshVotes)
	http.HandleFunc("/ws", handleWebSocket)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func revealVotesHandler(w http.ResponseWriter, r *http.Request){ 
  roomId := r.FormValue("room")

  room := rooms[roomId]
  room.showVotes = true
  vote_div := renderVoteDiv(room)
  sendResponse(room, vote_div)
}

func sendResponse(room Room, response string) {
  for conn := range room.connections {
    err := conn.WriteMessage(1, []byte(response))

    if err != nil {
        log.Println("Write error:", err)
        conn.Close()
        delete(room.connections, conn)
    }
  }
}



func createRoom(w http.ResponseWriter, r *http.Request) {
  newUUID := uuid.New().String()
  log.Println("Creating room: ", newUUID)
  room := Room{
    id: newUUID,
    connections: make(map[*websocket.Conn]Connection),
    voteMap: make(map[string]string),
    showVotes: false,
  }

  rooms[newUUID] = room
  name := r.FormValue("name")
  renderVotingRoom(w, name, room, true)
}

func getRoom(r *http.Request) Room{
    roomId  := r.URL.Query().Get("room")

    room := rooms[roomId]
    return room
}


func isAdmin(r *http.Request) bool{
    isAdminStr  := r.URL.Query().Get("admin")

    isAdmin, err := strconv.ParseBool(isAdminStr)

    if err != nil {
      log.Fatal(err)
    }
    
    return isAdmin
}



func joinRoom(w http.ResponseWriter, r *http.Request) {
    name := r.FormValue("name")
    roomId := r.FormValue("room")
    room := rooms[roomId] 
    room.voteMap[name] = "";   
    renderVotingRoom(w, name, room, false)
}

func renderVotingRoom(w http.ResponseWriter, name string, room Room, isAdmin bool) {

    votes := renderVoteDiv(room)
    data := struct {
        Room string
        Votes template.HTML
        Name  string
        Admin bool
    }{
        Room: room.id,
        Votes: template.HTML(votes),
        Name: name,
        Admin: isAdmin,
  }

    tmpl, err := template.ParseFiles("./site/voting.html")
    if err != nil {
        http.Error(w, "Unable to load template", http.StatusInternalServerError)
        return
    }

    err = tmpl.Execute(w, data)

    log.Println(err)
    if err != nil {
        http.Error(w, "Unable to render template", http.StatusInternalServerError)
        return
    }
}



func renderVoteDiv(room Room) string {
    data := struct {
        Votes map[string]string
        ShowVotes bool
    }{
        Votes: room.voteMap,
        ShowVotes: room.showVotes,
    }


    tmpl, err := template.ParseFiles("./site/votes_div.html")

    if err != nil {
        log.Println("Unable to load template", err)
        return ""
    }

    var buf bytes.Buffer
    // Execute the template with the data
    err = tmpl.Execute(&buf, data)
    if err != nil {
        log.Println("Unable to render template:", err)
        return ""
    }

    // Return the rendered template as a string
    // log.Println(buf.String())
    return buf.String()
}

