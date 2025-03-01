package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"
)

// TODO
// - join room link
// - create observer option

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Room struct {
  id string
  userMap map[string]*User
  showVotes bool
  cardValues []string
}

type User struct {
  conn *websocket.Conn
  vote string
  name string
  // isAdmin bool
}

var rooms = make(map[string]*Room)

func getBasePath() string {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}
	dir := filepath.Dir(execPath) 
	return dir

}

func getHtmlPath(file_name string)string{
  basePath := getBasePath()
  
  filePath := filepath.Join(basePath, "site", file_name)
  return filePath
}

func main() {

  basePath := getBasePath()
  staticDir := filepath.Join(basePath, "site", "static")

  fs := http.FileServer(http.Dir(staticDir))
  http.Handle("/static/", http.StripPrefix("/static/", fs)) // Serve static files
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
      log.Printf("Received request for: %s", r.URL.Path)
      http.ServeFile(w, r, getHtmlPath("index.html"))
  })
  http.HandleFunc("/join-room", joinRoom)
  http.HandleFunc("/create-room", createRoom)
  http.HandleFunc("/reveal-votes", revealVotesHandler)
  http.HandleFunc("/update-vote", updateVoteHandler)
  http.HandleFunc("/reset-votes", resetVotesHandler)
	http.HandleFunc("/ws", handleWebSocket)

	log.Println("Server started on :8080")
  err := http.ListenAndServe(":8080", nil)
  if err != nil {
      log.Fatalf("Server failed: %v", err)
  }
}

func getRoom(r *http.Request) *Room{
    roomId  := r.URL.Query().Get("room")

    room := rooms[roomId]
    return room
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {

  room := getRoom(r)

  userId := r.URL.Query().Get("userid")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()
  defer userCleanup(room, userId)

  room.userMap[userId].conn = conn

  sendWebsocket(room, renderVoteDiv(room))
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
	}
}

func userCleanup(room *Room, userId string){
  delete(room.userMap, userId)
  voteDiv := renderVoteDiv(room)
  sendWebsocket(room, voteDiv)
}


func updateVoteHandler(w http.ResponseWriter, r *http.Request){ 
  roomId := r.FormValue("room")
  userId := r.FormValue("userid")
  vote := r.FormValue("vote")
  log.Println("vote:", vote)

  room := rooms[roomId]
  user := room.userMap[userId]
  user.vote = vote

  room.userMap[userId] = user

  vote_div := renderVoteDiv(room)

  renderVoteButtonDiv(w, room, vote)
  sendWebsocket(room, vote_div)
}

func renderVoteButtonDiv(w http.ResponseWriter, room *Room, vote string){
    data := struct {
        CardValues []string
        CurrentVote string
    }{
        CardValues: room.cardValues,
        CurrentVote: vote,
    }

    tmpl, err := template.ParseFiles(getHtmlPath("vote_button_div.html"))
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

func resetVotesHandler(w http.ResponseWriter, r *http.Request) {
  roomId := r.FormValue("room")
  room := rooms[roomId]
  room.showVotes = false

  var keys []string
  for userId := range room.userMap {
      keys = append(keys, userId)
  }

  for _, userId := range keys {
      user := room.userMap[userId]
      user.vote = "" 
      room.userMap[userId] = user 
  }

  vote_div := renderVoteDiv(room)
  sendWebsocket(room, vote_div)
}


func revealVotesHandler(w http.ResponseWriter, r *http.Request){ 
  roomId := r.FormValue("room")

  room := rooms[roomId]
  room.showVotes = true
  vote_div := renderVoteDiv(room)
  sendWebsocket(room, vote_div)
}

func sendWebsocket(room *Room, response string) {
  for userId, user := range room.userMap {
    err := user.conn.WriteMessage(1, []byte(response))

    if err != nil {
        log.Println("Write error:", err)
        user.conn.Close()
        delete(room.userMap, userId)
    }
  }
}



func createRoom(w http.ResponseWriter, r *http.Request) {
  uid := ulid.Make().String()

  userId := ulid.Make().String()
  name := r.FormValue("name")
  cardValuesStr := r.FormValue("card-values")
  cardValues := strings.Split(cardValuesStr, ",")

  user := &User{
    name: name,
  }

  log.Println("Creating room: ", uid)
  room := &Room{
    id: uid,
    userMap: map[string]*User{userId: user},
    showVotes: false,
    cardValues: cardValues,
  }

  rooms[uid] = room

  renderVotingRoom(w, userId, room, true)
}



func joinRoom(w http.ResponseWriter, r *http.Request) {
    uid := ulid.Make().String()
    name := r.FormValue("name")
    roomId := r.FormValue("room")

    room, exists := rooms[roomId]
    if !exists {
        http.Error(w, "Room not found", http.StatusNotFound)
        return
    }
    user := &User{
      name: name,
    }

    room.userMap[uid] = user;   
    renderVotingRoom(w, uid, room, false)
}

func renderVotingRoom(w http.ResponseWriter, userId string, room *Room, isAdmin bool) {

    votes := renderVoteDiv(room)
    data := struct {
        Room string
        Votes template.HTML
        Admin bool
        UserId string
        CardValues []string
    }{
        Room: room.id,
        Votes: template.HTML(votes),
        Admin: isAdmin,
        UserId: userId, 
        CardValues: room.cardValues,
    }

    tmpl, err := template.ParseFiles(getHtmlPath("voting.html"))
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



func renderVoteDiv(room *Room) string {

    votes := make(map[string]string)
      
    for _, user := range room.userMap {
      if room.showVotes {
        votes[user.name] = user.vote
      } else if user.vote == "" {
        votes[user.name] = ""
      } else {
        votes[user.name] = "*"
      }
    }

    data := struct {
        Votes map[string]string
    }{
        Votes: votes,
    }


    tmpl, err := template.ParseFiles(getHtmlPath("votes_div.html"))

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

    return buf.String()
}

