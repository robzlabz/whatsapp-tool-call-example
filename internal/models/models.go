package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session represents a WhatsApp session
type Session struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	JID       string    `gorm:"uniqueIndex;not null" json:"jid"`
	Data      []byte    `gorm:"type:blob" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents a WhatsApp message
type Message struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	MessageID   string    `gorm:"uniqueIndex;not null" json:"message_id"`
	FromJID     string    `gorm:"not null" json:"from_jid"`
	ToJID       string    `gorm:"not null" json:"to_jid"`
	Content     string    `gorm:"type:text" json:"content"`
	MessageType string    `gorm:"not null" json:"message_type"`
	IsFromMe    bool      `gorm:"default:false" json:"is_from_me"`
	Timestamp   time.Time `gorm:"not null" json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Conversation represents a conversation thread
type Conversation struct {
	ID           uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	JID          string    `gorm:"uniqueIndex;not null" json:"jid"`
	LastMessage  string    `gorm:"type:text" json:"last_message"`
	MessageCount int       `gorm:"default:0" json:"message_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Messages     []Message `gorm:"foreignKey:FromJID;references:JID" json:"messages,omitempty"`
}

// ToolExecution represents a tool execution log
type ToolExecution struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	MessageID   string    `gorm:"not null" json:"message_id"`
	ToolName    string    `gorm:"not null" json:"tool_name"`
	Parameters  string    `gorm:"type:text" json:"parameters"`
	Result      string    `gorm:"type:text" json:"result"`
	Success     bool      `gorm:"default:false" json:"success"`
	ErrorMsg    string    `gorm:"type:text" json:"error_msg,omitempty"`
	ExecutionTime int64   `gorm:"not null" json:"execution_time"` // milliseconds
	CreatedAt   time.Time `json:"created_at"`
}

// BeforeCreate hooks for UUID generation
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (t *ToolExecution) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}