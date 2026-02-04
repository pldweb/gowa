package gateway

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pldweb/gowa/internal/config"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	waProto "go.mau.fi/whatsmeow/binary/proto"
)

// Gateway manages WhatsApp client connections
type Gateway struct {
	client      *whatsmeow.Client
	config      *config.Config
	container   *sqlstore.Container
	eventHandle chan interface{}
}

// New creates a new WhatsApp gateway instance
func New(cfg *config.Config) (*Gateway, error) {
	// Ensure session directory exists
	if err := os.MkdirAll(cfg.SessionPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	// Initialize database container for session storage
	dbPath := filepath.Join(cfg.SessionPath, "session.db")
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(context.Background(), "sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath), dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Get first device from store or create new one
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	gw := &Gateway{
		client:      client,
		config:      cfg,
		container:   container,
		eventHandle: make(chan interface{}, 100),
	}

	// Set up event handlers
	client.AddEventHandler(gw.eventHandler)

	return gw, nil
}

// Connect establishes connection to WhatsApp
func (g *Gateway) Connect(qrChan chan string) error {
	if g.client.Store.ID == nil {
		// No ID stored, new login required
		qrChanInternal, err := g.client.GetQRChannel(context.Background())
		if err != nil {
			// Check if already connected
			if g.client.IsConnected() {
				close(qrChan)
				return nil
			}
			return fmt.Errorf("failed to get QR channel: %w", err)
		}

		err = g.client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}

		// Wait for QR code or connection
		go func() {
			for evt := range qrChanInternal {
				if evt.Event == "code" {
					qrChan <- evt.Code
				} else {
					// Connected
					qrChan <- ""
					break
				}
			}
		}()
	} else {
		// Already have credentials, just connect
		err := g.client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		close(qrChan)
	}

	return nil
}

// IsConnected returns whether the client is connected
func (g *Gateway) IsConnected() bool {
	return g.client.IsConnected()
}

// GetJID returns the JID of the connected device
func (g *Gateway) GetJID() string {
	if g.client.Store.ID == nil {
		return ""
	}
	return g.client.Store.ID.String()
}

// SendMessage sends a text message to a phone number
func (g *Gateway) SendMessage(ctx context.Context, phone, message string) (string, error) {
	if !g.client.IsConnected() {
		return "", fmt.Errorf("client not connected")
	}

	// Parse phone number to JID
	phone = strings.TrimPrefix(phone, "+")
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	jid := types.NewJID(phone, types.DefaultUserServer)

	// Send message
	msg := &waProto.Message{
		Conversation: proto.String(message),
	}

	resp, err := g.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return resp.ID, nil
}

// Logout logs out from WhatsApp and removes session data
func (g *Gateway) Logout() error {
	if !g.client.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	err := g.client.Logout(context.Background())
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}

// Close closes the gateway and cleans up resources
func (g *Gateway) Close() {
	if g.client != nil {
		g.client.Disconnect()
	}
}

// eventHandler handles incoming WhatsApp events
func (g *Gateway) eventHandler(evt interface{}) {
	// Handle events here - can be extended for webhooks, logging, etc.
	select {
	case g.eventHandle <- evt:
	default:
		// Channel full, drop event
	}
}
