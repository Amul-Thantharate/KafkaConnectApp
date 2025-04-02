# 💬 KafkaConnect Chat Application 📚

[![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8?style=flat&logo=go)](https://golang.org)
[![MySQL](https://img.shields.io/badge/MySQL-8.0+-00758F?style=flat&logo=mysql)](https://www.mysql.com/)
[![Apache Kafka](https://img.shields.io/badge/Apache%20Kafka-3.0+-231F20?style=flat&logo=apache-kafka)](https://kafka.apache.org/)
[![Docker](https://img.shields.io/badge/Docker-20.10+-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

🚀 A modern, distributed chat system that combines the speed of Go, the reliability of MySQL, and the scalability of Kafka. Perfect for building real-time communication platforms that can handle thousands of concurrent users with ease. Whether you're creating a team chat, community platform, or learning distributed systems - this project has you covered!

A scalable, real-time chat application built with modern technologies:
- **Go (Golang)** for high-performance server-side operations
- **MySQL** for reliable user and group data persistence
- **Apache Kafka** for robust message queuing and real-time communication
- **Docker** for containerized deployment and easy scaling

## ✨ Key Features

### 👤 User Management
- Secure user authentication with bcrypt password encryption
- User registration and login system
- Real-time online/offline status tracking
- User session management

### 💬 Messaging System
- Real-time message delivery
- Private messaging between users
- Broadcast messages to all online users
- Persistent message history through Kafka

### 👥 Group Chat
- Create and manage chat groups
- Join/leave group functionality
- Group-specific messaging
- Member count tracking
- Real-time group updates

### 🔔 System Features
- Real-time notifications for user activities
- Join/leave event broadcasting
- Online user list updates
- Group membership tracking
- Error handling and user feedback

## 📋 Prerequisites

### Required Software
- Go 1.16 or higher
- Docker 20.10+ and Docker Compose
- NetCat (nc) for client connections

### System Requirements
- Minimum 4GB RAM
- 10GB free disk space
- Linux/Unix-based system (recommended)

## 🛠️ Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Amul-Thantharate/KafkaConnectApp.git
   cd KafkaConnectApp
   ```

2. **Start the infrastructure**
   ```bash
   docker-compose up -d
   ```
   This will start:
   - MySQL on port 3307
   - Zookeeper on port 2181
   - Kafka on port 9092

3. **Run the chat server**
   ```bash
   go run main.go
   ```
   The server will start on port 8080.

## 💬 Usage

1. **Connect to the chat server**
   ```bash
   nc localhost 8080
   ```

2. **Available Commands**
   ```
   /help                          - Show all available commands
   /register <username> <password> - Create a new account
   /login <username> <password>    - Login to your account
   /logout                        - Logout from current session
   /list_user                     - Show online users
   /pm <username> <message>       - Send private message
   /broadcast <message>           - Send message to all users
   /group_create <group_name>     - Create a new group
   /group_join <group_name>       - Join an existing group
   /group_leave <group_name>      - Leave a group
   /group_msg <group_name> <msg>  - Send message to group
   ```

## 🔍 Debugging and Monitoring

### Database Inspection

1. **Connect to MySQL**
   ```bash
   mysql -h localhost -P 3307 -u root -proot chat_db
   ```

2. **Useful MySQL queries**
   ```sql
   -- List all users
   SELECT * FROM users;

   -- Show online users
   SELECT * FROM users WHERE online = true;

   -- List all groups
   SELECT * FROM groups;

   -- Show group members
   SELECT g.name, u.username 
   FROM group_members gm 
   JOIN groups g ON gm.group_id = g.id 
   JOIN users u ON gm.user_id = u.id;
   ```

### Kafka Message Inspection

1. **View broadcast messages**
   ```bash
   # Install kafkacat if not available
   sudo apt-get install kafkacat

   # Listen to broadcast messages
   kafkacat -b localhost:9092 -t broadcast -C

   # Listen to system messages
   kafkacat -b localhost:9092 -t system -C
   ```

2. **Using Kafka Console Consumer**
   ```bash
   # If you have Kafka tools installed
   kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic broadcast --from-beginning
   ```

## 🏗️ Architecture

- **TCP Server**: Handles client connections and command processing
- **MySQL**: Stores user accounts, groups, and membership information
- **Kafka**: Manages message delivery and persistence
- **GORM**: ORM for database operations
- **bcrypt**: Secure password hashing

## 📝 Database Schema

```sql
-- Users table
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(255) UNIQUE,
    password VARCHAR(255),
    online BOOLEAN
);

-- Groups table
CREATE TABLE groups (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) UNIQUE
);

-- Group Members table
CREATE TABLE group_members (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT,
    group_id BIGINT,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (group_id) REFERENCES groups(id)
);
```

## 🛑 Stopping the Application

1. **Stop the Go server**
   ```bash
   # Press Ctrl+C in the terminal running the server
   # Or find and kill the process
   pkill -f "go run main.go"
   ```

2. **Stop the infrastructure**
   ```bash
   docker-compose down
   ```

## 🔒 Security Notes

- Passwords are hashed using bcrypt before storage
- All database credentials should be changed in production
- Consider adding TLS for secure communication in production

## 🐛 Troubleshooting

1. **MySQL Connection Issues**
   - Ensure MySQL is running: `docker-compose ps`
   - Check port availability: `netstat -an | grep 3307`
   - Verify credentials in database/database.go

2. **Kafka Connection Issues**
   - Ensure Kafka is running: `docker-compose ps`
   - Check Kafka logs: `docker-compose logs kafka`
   - Verify connection string in kafka/kafka.go

3. **Common Issues**
   - Port conflicts: Change ports in docker-compose.yml
   - Database migration: Check GORM logs for schema updates
   - Message delivery: Monitor Kafka consumer lag

## 🤝 Contributing

Feel free to submit issues, fork the repository, and create pull requests for any improvements.

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🌐 Contact

For questions or support, please contact Amul Thantharate at amulthantharate@gmail.com
Thanks for using this project! ❤️
