package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pldweb/gowa/internal/config"
	"github.com/pldweb/gowa/internal/gateway"
	"github.com/skip2/go-qrcode"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize WhatsApp gateway
	gw, err := gateway.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize gateway: %v", err)
	}
	defer gw.Close()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "GoWA - WhatsApp Gateway Multidevice",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"service": "gowa",
		})
	})

	// Get QR code for pairing
	app.Get("/qr", func(c *fiber.Ctx) error {
		if gw.IsConnected() {
			return c.Status(400).JSON(fiber.Map{
				"error": "Already connected",
			})
		}

		qrChan := make(chan string, 1)
		err := gw.Connect(qrChan)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Wait for QR code
		qrCode := <-qrChan
		if qrCode == "" {
			return c.JSON(fiber.Map{
				"status": "connected",
			})
		}

		// Generate QR code image
		png, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to generate QR code",
			})
		}

		return c.JSON(fiber.Map{
			"qr_code": qrCode,
			"qr_image": base64.StdEncoding.EncodeToString(png),
		})
	})

	// Get connection status
	app.Get("/status", func(c *fiber.Ctx) error {
		connected := gw.IsConnected()
		var jid string
		if connected {
			jid = gw.GetJID()
		}
		return c.JSON(fiber.Map{
			"connected": connected,
			"jid": jid,
		})
	})

	// Send message
	app.Post("/send", func(c *fiber.Ctx) error {
		if !gw.IsConnected() {
			return c.Status(400).JSON(fiber.Map{
				"error": "Not connected",
			})
		}

		type SendRequest struct {
			Phone   string `json:"phone"`
			Message string `json:"message"`
		}

		var req SendRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Phone == "" || req.Message == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Phone and message are required",
			})
		}

		messageID, err := gw.SendMessage(c.Context(), req.Phone, req.Message)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status": "sent",
			"message_id": messageID,
		})
	})

	// Logout
	app.Post("/logout", func(c *fiber.Ctx) error {
		if !gw.IsConnected() {
			return c.Status(400).JSON(fiber.Map{
				"error": "Not connected",
			})
		}

		err := gw.Logout()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status": "logged out",
		})
	})

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("Starting server on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down gracefully...")
	app.Shutdown()
}
