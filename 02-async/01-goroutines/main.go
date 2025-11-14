package main
import "fmt"

type User struct {
	Email string
}

type UserRepository interface {
	CreateUserAccount(u User) error
}

type NotificationsClient interface {
	SendNotification(u User) error
}

type NewsletterClient interface {
	AddToNewsletter(u User) error
}

type Handler struct {
	repository          UserRepository
	newsletterClient    NewsletterClient
	notificationsClient NotificationsClient
}

func NewHandler(
	repository UserRepository,
	newsletterClient NewsletterClient,
	notificationsClient NotificationsClient,
) Handler {
	return Handler{
		repository:          repository,
		newsletterClient:    newsletterClient,
		notificationsClient: notificationsClient,
	}
}

func (h Handler) SignUp(u User) error {
	// This is a critical operation, so it should be done synchronously.
	if err := h.repository.CreateUserAccount(u); err != nil {
		return err
	}

	go func() {
		for {
			err := h.newsletterClient.AddToNewsletter(u)

			if err != nil {
				fmt.Println("failed to add user to newsletter:", err)
				continue
			}	
			
			break;
		}
	}()

	go func () {
		for {
			err := h.notificationsClient.SendNotification(u)

			if err != nil {
				fmt.Println("failed to send notification:", err)
				continue
			}	

			break;
		}
	}()
	

	return nil
}
