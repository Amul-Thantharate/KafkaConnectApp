package server

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/Amul-Thantharate/KafkaConnect/kafka"

	"github.com/Amul-Thantharate/KafkaConnect/database"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	clients      = make(map[net.Conn]string)
	clientsMutex sync.Mutex
)

// Handle incoming client commands
func HandleCommand(conn net.Conn, message string) {
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	switch command {
	case "/help":
		showHelp(conn)
	case "/register":
		registerUser(conn, parts)
	case "/login":
		loginUser(conn, parts)
	case "/logout":
		logoutUser(conn)
	case "/list_user":
		listUsers(conn)
	case "/list_groups":
		listGroups(conn)
	case "/pm":
		privateMessage(conn, parts)
	case "/broadcast":
		broadcastMessage(conn, parts)
	case "/group_create":
		createGroup(conn, parts)
	case "/group_join":
		joinGroup(conn, parts)
	case "/group_leave":
		leaveGroup(conn, parts)
	case "/group_msg":
		sendGroupMessage(conn, parts)
	default:
		conn.Write([]byte("Unknown command. Type /help for available commands.\n"))
	}
}

// 🔑 User Authentication
func registerUser(conn net.Conn, parts []string) {
	if len(parts) < 3 {
		conn.Write([]byte("Usage: /register <username> <password>\n"))
		return
	}
	username, password := parts[1], parts[2]

	// Check if user exists
	var existingUser database.User
	result := database.DB.Where("username = ?", username).First(&existingUser)
	if result.Error == nil {
		conn.Write([]byte("Username already taken\n"))
		return
	}

	// Hash password and store user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	newUser := database.User{Username: username, Password: string(hashedPassword), Online: false}
	database.DB.Create(&newUser)

	conn.Write([]byte("Registration successful! Use /login <username> <password> to login.\n"))
}

func loginUser(conn net.Conn, parts []string) {
	if len(parts) != 3 {
		conn.Write([]byte("Usage: /login <username> <password>\n"))
		return
	}

	username := parts[1]
	password := parts[2]

	var user database.User
	result := database.DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		conn.Write([]byte("Invalid username or password\n"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		conn.Write([]byte("Invalid username or password\n"))
		return
	}

	clientsMutex.Lock()
	clients[conn] = username
	clientsMutex.Unlock()

	// Update user's online status
	database.DB.Model(&user).Update("online", true)

	conn.Write([]byte("Login successful! You can now chat.\n"))

	// Broadcast that a new user has joined
	broadcastSystemMessage(fmt.Sprintf("🟢 %s has joined the chat", username))
}

func logoutUser(conn net.Conn) {
	clientsMutex.Lock()
	username, exists := clients[conn]
	if exists {
		// Update user's online status first
		var user database.User
		database.DB.Where("username = ?", username).First(&user)
		database.DB.Model(&user).Update("online", false)

		// Remove from clients map
		delete(clients, conn)

		// Broadcast that user has left after updating status
		broadcastSystemMessage(fmt.Sprintf("🔴 %s has left the chat", username))
	}
	clientsMutex.Unlock()
	conn.Write([]byte("Logged out successfully\n"))
}

func showHelp(conn net.Conn) {
	helpText := `Available Commands:
/help                          - Show this help message
/register <username> <password> - Create a new account
/login <username> <password>    - Login to your account
/logout                        - Logout from current session
/list_user                     - Show online users
/list_groups                   - Show all available groups
/pm <username> <message>       - Send private message
/broadcast <message>           - Send message to all users
/group_create <group_name>     - Create a new group
/group_join <group_name>       - Join an existing group
/group_leave <group_name>      - Leave a group
/group_msg <group_name> <msg>  - Send message to group
`
	conn.Write([]byte(helpText))
}

// 📃 List Online Users
func listUsers(conn net.Conn) {
	var users []database.User
	database.DB.Where("online = ?", true).Find(&users)

	response := "Online Users:\n"
	for _, user := range users {
		response += "- " + user.Username + "\n"
	}
	conn.Write([]byte(response))
}

func listGroups(conn net.Conn) {
	var groups []database.Group
	database.DB.Find(&groups)

	if len(groups) == 0 {
		conn.Write([]byte("No groups available\n"))
		return
	}

	response := "Available Groups:\n"
	for _, group := range groups {
		// Get member count for each group
		var memberCount int64
		database.DB.Model(&database.GroupMember{}).Where("group_id = ?", group.ID).Count(&memberCount)
		response += fmt.Sprintf("- %s (Members: %d)\n", group.Name, memberCount)
	}
	conn.Write([]byte(response))
}

// ✉️ Private Messaging
func privateMessage(conn net.Conn, parts []string) {
	if len(parts) < 3 {
		conn.Write([]byte("Usage: /pm <username> <message>\n"))
		return
	}
	recipient, message := parts[1], strings.Join(parts[2:], " ")

	clientsMutex.Lock()
	for client, username := range clients {
		if username == recipient {
			client.Write([]byte("Private from " + clients[conn] + ": " + message + "\n"))
			clientsMutex.Unlock()
			return
		}
	}
	clientsMutex.Unlock()
	conn.Write([]byte("User not found or offline\n"))
}

// 📢 Broadcast Message
func broadcastMessage(conn net.Conn, parts []string) {
	if len(parts) < 2 {
		conn.Write([]byte("Usage: /broadcast <message>\n"))
		return
	}

	clientsMutex.Lock()
	sender := clients[conn]
	clientsMutex.Unlock()

	message := strings.Join(parts[1:], " ")
	broadcastText := fmt.Sprintf("📢 [%s]: %s\n", sender, message)

	// Send to Kafka
	kafka.SendMessage("broadcast", broadcastText)

	// Send to all connected clients
	clientsMutex.Lock()
	for client := range clients {
		client.Write([]byte(broadcastText))
	}
	clientsMutex.Unlock()
}

func broadcastSystemMessage(message string) {
	systemMessage := fmt.Sprintf("🔔 System: %s\n", message)

	// Send to Kafka
	kafka.SendMessage("system", systemMessage)

	// Send to all connected clients
	clientsMutex.Lock()
	for client := range clients {
		client.Write([]byte(systemMessage))
	}
	clientsMutex.Unlock()
}

// 👥 Group Chat Commands
func createGroup(conn net.Conn, parts []string) {
	if len(parts) < 2 {
		conn.Write([]byte("Usage: /group_create <group_name>\n"))
		return
	}
	groupName := parts[1]

	// Check if group exists
	var existingGroup database.Group
	result := database.DB.Where("name = ?", groupName).First(&existingGroup)
	if result.Error == nil {
		conn.Write([]byte("Group already exists\n"))
		return
	}

	// Create group
	newGroup := database.Group{Name: groupName}
	database.DB.Create(&newGroup)

	// Create Kafka topic for the group
	kafka.SendMessage(groupName, "Group created: "+groupName)

	conn.Write([]byte("Group created successfully\n"))
}

func joinGroup(conn net.Conn, parts []string) {
	if len(parts) < 2 {
		conn.Write([]byte("Usage: /group_join <group_name>\n"))
		return
	}
	groupName := parts[1]

	// Check if group exists
	var group database.Group
	result := database.DB.Where("name = ?", groupName).First(&group)
	if result.Error == gorm.ErrRecordNotFound {
		conn.Write([]byte("Group not found\n"))
		return
	}

	// Get user
	username, exists := clients[conn]
	if !exists {
		conn.Write([]byte("You must be logged in to join a group\n"))
		return
	}

	var user database.User
	database.DB.Where("username = ?", username).First(&user)

	// Add user to group
	database.DB.Create(&database.GroupMember{UserID: user.ID, GroupID: group.ID})
	conn.Write([]byte("Joined group successfully\n"))
}

func leaveGroup(conn net.Conn, parts []string) {
	if len(parts) < 2 {
		conn.Write([]byte("Usage: /group_leave <group_name>\n"))
		return
	}
	groupName := parts[1]

	username, exists := clients[conn]
	if !exists {
		conn.Write([]byte("You must be logged in to leave a group\n"))
		return
	}

	var group database.Group
	database.DB.Where("name = ?", groupName).First(&group)

	var user database.User
	database.DB.Where("username = ?", username).First(&user)

	// Remove user from group
	var groupMember database.GroupMember
	database.DB.Where("user_id = ? AND group_id = ?", user.ID, group.ID).First(&groupMember)
	database.DB.Delete(&groupMember)
	conn.Write([]byte("Left group successfully\n"))
}

func sendGroupMessage(conn net.Conn, parts []string) {
	if len(parts) < 3 {
		conn.Write([]byte("Usage: /group_msg <group_name> <message>\n"))
		return
	}
	groupName := parts[1]
	message := strings.Join(parts[2:], " ")

	// Check if group exists
	var group database.Group
	if err := database.DB.Where("name = ?", groupName).First(&group).Error; err != nil {
		conn.Write([]byte("Group not found\n"))
		return
	}

	// Check if user is member of the group
	var member database.GroupMember
	if err := database.DB.Where("group_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)",
		group.ID, clients[conn]).First(&member).Error; err != nil {
		conn.Write([]byte("You are not a member of this group\n"))
		return
	}

	// Format the message
	formattedMessage := fmt.Sprintf("👥 [%s][%s]: %s\n", groupName, clients[conn], message)

	// Send to Kafka
	kafka.SendMessage(groupName, formattedMessage)

	// Broadcast to all group members
	var groupMembers []database.GroupMember
	database.DB.Where("group_id = ?", group.ID).Find(&groupMembers)

	clientsMutex.Lock()
	for client, username := range clients {
		// Check if this client is a member of the group
		for _, member := range groupMembers {
			var user database.User
			if database.DB.First(&user, member.UserID).Error == nil && user.Username == username {
				client.Write([]byte(formattedMessage))
				break
			}
		}
	}
	clientsMutex.Unlock()

	conn.Write([]byte("Message sent to group\n"))
}
