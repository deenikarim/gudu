package mails

import (
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

// Scheduler schedules emails to be sent at a later time
type Scheduler struct {
	C         *cron.Cron
	Queue     chan *Message
	Transport MailTransport
}

// NewScheduler creates a new Scheduler
func NewScheduler(t MailTransport) *Scheduler {
	return &Scheduler{
		C:         cron.New(cron.WithSeconds()), // Ensure we support second-level granularity
		Queue:     make(chan *Message, 100),
		Transport: t,
	}
}

// ScheduleEmail schedules an email to be sent at a specific time
func (s *Scheduler) ScheduleEmail(message *Message, sendTime time.Time) (cron.EntryID, error) {
	// Convert sendTime to cron expression with second-level granularity
	cronExpr := sendTime.Format("05 04 15 02 Jan Mon")

	id, err := s.C.AddFunc(cronExpr, func() {
		s.Queue <- message
		log.Printf("Scheduled email sent to %v", message.To)
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	go func() {
		for msg := range s.Queue {
			if err := s.Transport.Send(msg); err != nil {
				log.Printf("Failed to send scheduled email to %v: %v", msg.To, err)
			} else {
				log.Printf("Scheduled email sent successfully to %v", msg.To)
			}
		}
	}()
	s.C.Start()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.C.Stop()
	close(s.Queue)
}
