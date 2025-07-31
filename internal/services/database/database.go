package database

import (
	"fmt"
	"strings"

	"example-tool-call/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
}

func New(databaseURL string) (*DB, error) {
	var db *gorm.DB
	var err error

	// Parse database URL to determine driver
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else if strings.HasPrefix(databaseURL, "sqlite://") {
		dbPath := strings.TrimPrefix(databaseURL, "sqlite://")
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		return nil, fmt.Errorf("unsupported database URL format: %s", databaseURL)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(
		&models.Session{},
		&models.Message{},
		&models.Conversation{},
		&models.ToolExecution{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &DB{DB: db}, nil
}

func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Session operations
func (db *DB) SaveSession(session *models.Session) error {
	return db.Save(session).Error
}

func (db *DB) GetSession(jid string) (*models.Session, error) {
	var session models.Session
	err := db.Where("jid = ?", jid).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Message operations
func (db *DB) SaveMessage(message *models.Message) error {
	return db.Create(message).Error
}

func (db *DB) GetMessages(jid string, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := db.Where("from_jid = ? OR to_jid = ?", jid, jid).
		Order("timestamp DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// Conversation operations
func (db *DB) GetOrCreateConversation(jid string) (*models.Conversation, error) {
	var conversation models.Conversation
	err := db.Where("jid = ?", jid).First(&conversation).Error
	if err == gorm.ErrRecordNotFound {
		conversation = models.Conversation{
			JID:          jid,
			MessageCount: 0,
		}
		err = db.Create(&conversation).Error
	}
	return &conversation, err
}

func (db *DB) UpdateConversation(conversation *models.Conversation) error {
	return db.Save(conversation).Error
}

// Tool execution operations
func (db *DB) SaveToolExecution(execution *models.ToolExecution) error {
	return db.Create(execution).Error
}

func (db *DB) GetToolExecutions(messageID string) ([]models.ToolExecution, error) {
	var executions []models.ToolExecution
	err := db.Where("message_id = ?", messageID).Find(&executions).Error
	return executions, err
}