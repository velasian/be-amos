package firebase

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// Client wraps the Firebase Cloud Messaging client.
type Client struct {
	msgClient *messaging.Client
	available bool
}

// NewClient initializes a Firebase Admin SDK client using a service account key file.
// Returns a no-op client if the key file is not found (non-fatal for development).
func NewClient() *Client {
	keyPath := os.Getenv("FIREBASE_CREDENTIALS")
	if keyPath == "" {
		keyPath = "firebase-service-account.json"
	}

	// Check if credentials file exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		log.Printf("⚠️  Firebase credentials not found at '%s' (push notifications disabled)", keyPath)
		return &Client{available: false}
	}

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(keyPath))
	if err != nil {
		log.Printf("⚠️  Firebase init failed: %v (push notifications disabled)", err)
		return &Client{available: false}
	}

	msgClient, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("⚠️  Firebase Messaging init failed: %v (push notifications disabled)", err)
		return &Client{available: false}
	}

	log.Println("✅ Firebase Cloud Messaging connected")
	return &Client{
		msgClient: msgClient,
		available: true,
	}
}

// IsAvailable returns whether Firebase is properly configured.
func (c *Client) IsAvailable() bool {
	return c.available
}

// SendToToken sends a push notification to a single device token.
func (c *Client) SendToToken(ctx context.Context, token, title, body string, data map[string]string) error {
	if !c.available {
		log.Println("[FCM] Skipped: Firebase not configured")
		return nil
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound:       "default",
				ChannelID:   "amos_default",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
			},
		},
	}

	response, err := c.msgClient.Send(ctx, message)
	if err != nil {
		return err
	}

	log.Printf("[FCM] Notification sent: %s", response)
	return nil
}

// SendToMultiple sends a push notification to multiple device tokens.
func (c *Client) SendToMultiple(ctx context.Context, tokens []string, title, body string, data map[string]string) (int, int, error) {
	if !c.available {
		log.Println("[FCM] Skipped: Firebase not configured")
		return 0, 0, nil
	}

	if len(tokens) == 0 {
		return 0, 0, nil
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound:       "default",
				ChannelID:   "amos_default",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
			},
		},
	}

	response, err := c.msgClient.SendEachForMulticast(ctx, message)
	if err != nil {
		return 0, 0, err
	}

	log.Printf("[FCM] Multicast sent: %d success, %d failure", response.SuccessCount, response.FailureCount)
	return response.SuccessCount, response.FailureCount, nil
}
