package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

//go:embed index.html wordlist.txt
var content embed.FS

var (
	wordlist    = []string{}
	clients     = make(map[*websocket.Conn]bool)
	clientsMux  sync.Mutex
	broadcaster = make(chan string)
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	port := flag.String("port", "9090", "port to run the server on")
	wordlistFile := flag.String("wordlist", "", "path to the wordlist file")

	flag.Parse()

	// Load the wordlist from the specified file
	if *wordlistFile != "" {
		var err error
		wordlist, err = loadWordlist(*wordlistFile)

		if err != nil {
			log.Fatalf("Error loading wordlist: %v", err)
		}
	} else {
		// Use the embedded wordlist.txt file
		embeddedFile, err := content.ReadFile("wordlist.txt")
		if err != nil {
			log.Fatalf("Error reading embedded wordlist: %v", err)
		}

		scanner := bufio.NewScanner(strings.NewReader(string(embeddedFile)))
		var words []string
		for scanner.Scan() {
			word := strings.TrimSpace(scanner.Text())
			if word != "" {
				words = append(words, word)
			}
		}

		if len(words) > 0 {
			wordlist = words
		}
	}

	// Serve the static HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		fmt.Printf("%s has connected to http://localhost:%s\n", clientIP, *port)

		tmpl := template.Must(template.ParseFS(content, "index.html"))
		tmpl.Execute(w, nil)

	})

	// Handle WebSocket connections
	http.HandleFunc("/ws", handleConnections)

	// Start the broadcaster in a goroutine
	go handleMessages()

	// Start the server
	fmt.Printf("Server started at http://localhost:%s\n", *port)
	fmt.Println("CefEnum. Enter HELP for a list of commands.")
	fmt.Println("Or enter JavaScript to be sent to the client.")
	fmt.Println("https://github.com/darkforge-labs/cefenum")

	go userInput()

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	updateClientWordlist(ws, wordlist)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer ws.Close()

	// Register new client
	clientsMux.Lock()
	clients[ws] = true
	clientsMux.Unlock()

	// Simple message handler to keep connection alive
	for {
		//print the message received
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
		fmt.Printf("< %s\n", msg)
		// Send the message to the broadcaster
		// broadcaster <- string(msg)

	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		js := <-broadcaster

		// Send it to every client connected
		clientsMux.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(js))
			if err != nil {
				log.Printf("Error sending message: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		clientsMux.Unlock()
	}
}

func userInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("# ") // Optional prompt
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Trim the newline and any spaces
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		// explode the input by spaces
		parts := strings.Fields(input)

		if len(parts) > 0 {
			// Check if the first part is a command
			switch strings.ToLower(parts[0]) {
			case "exit", "quit":
				fmt.Println("Exiting...")
				os.Exit(0)
			case "clear":
				fmt.Print("\033[H\033[2J")
				fmt.Print("\033[3J")
			case "fuzz":
				broadcaster <- "bindCommon();"
			case "detect":
				broadcaster <- "detectCefSharp();"
			case "brute":
				broadcaster <- "bruteForce();"
			case "bind":
				if len(parts) == 1 {
					fmt.Println("Usage: bind <objectName>")
				} else {
					objectName := strings.Join(parts[1:], " ")
					broadcaster <- fmt.Sprintf("bind('%s');", objectName)
				}
			case "help", "?":
				fmt.Println("Available commands:")
				fmt.Println("  exit/quit - Exit the program")
				fmt.Println("  clear - Clear the console")
				fmt.Println("  fuzz - Bind common objects")
				fmt.Println("  detect - Detect if CefSharp is present")
				fmt.Println("  brute - Brute force object discovery")
				fmt.Println("  bind <objectName> - Inspect an object")
				fmt.Println("  help - Show this help message")
			default:
				// Send the input to the broadcaster
				broadcaster <- input
			}
		}
	}
}

func loadWordlist(filePath string) ([]string, error) {

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening wordlist file: %v", err)
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	var words []string

	// Read each line as a word
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, word)
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading wordlist file: %v", err)
	}

	return words, nil
}

func updateClientWordlist(client *websocket.Conn, wordlist []string) {
	// Send the wordlist to the client
	fmt.Println("Sending wordlist to client")
	clientsMux.Lock()
	defer clientsMux.Unlock()

	// print the percentage of the wordlist
	percentage := float64(len(wordlist)) / float64(len(wordlist)) * 100
	for i, word := range wordlist {
		fmt.Printf("\rSending word %d of %d (%.2f%%)", i+1, len(wordlist), percentage)
		percentage = float64(i+1) / float64(len(wordlist)) * 100
		if percentage == 100 {
			fmt.Printf("Wordlist sent to client\n")
		}
		// Send the word to the client
		js := fmt.Sprintf("commonObjectNames.push(\"%s\");", word)
		err := client.WriteMessage(websocket.TextMessage, []byte(js))
		if err != nil {
			log.Printf("Error sending word: %v", err)
			client.Close()
			delete(clients, client)
			break
		}
	}
}
