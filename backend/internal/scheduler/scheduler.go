// Package scheduler provides scheduled job execution for FutureSignals.
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/leeaandrob/futuresignals/internal/content"
	"github.com/leeaandrob/futuresignals/internal/models"
	syncer "github.com/leeaandrob/futuresignals/internal/sync"
	"github.com/rs/zerolog/log"
)

// Job represents a scheduled job.
type Job struct {
	Name     string
	Schedule Schedule
	Handler  func(ctx context.Context) error
	LastRun  time.Time
	NextRun  time.Time
}

// Schedule defines when a job should run.
type Schedule struct {
	// For fixed-interval jobs
	Interval time.Duration

	// For time-of-day jobs (in UTC)
	Hour   int
	Minute int

	// Days (0=Sunday, 1=Monday, etc.)
	Days []int

	// Type of schedule
	Type ScheduleType
}

// ScheduleType defines the type of schedule.
type ScheduleType string

const (
	ScheduleInterval   ScheduleType = "interval"
	ScheduleDaily      ScheduleType = "daily"
	ScheduleWeekly     ScheduleType = "weekly"
)

// Scheduler manages scheduled jobs and event-driven content generation.
type Scheduler struct {
	generator *content.Generator
	syncer    *syncer.Syncer

	jobs    []*Job
	jobsMux sync.RWMutex

	// Event processing
	eventChan <-chan syncer.Event

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewScheduler creates a new scheduler.
func NewScheduler(generator *content.Generator, sync *syncer.Syncer) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Scheduler{
		generator: generator,
		syncer:    sync,
		jobs:      make([]*Job, 0),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Subscribe to syncer events
	if sync != nil {
		s.eventChan = sync.Subscribe()
	}

	// Register default jobs
	s.registerDefaultJobs()

	return s
}

// registerDefaultJobs sets up the default content generation schedule.
func (s *Scheduler) registerDefaultJobs() {
	// Morning briefing at 8:00 UTC
	s.AddJob(&Job{
		Name: "morning-briefing",
		Schedule: Schedule{
			Type:   ScheduleDaily,
			Hour:   8,
			Minute: 0,
		},
		Handler: func(ctx context.Context) error {
			_, err := s.generator.GenerateBriefing(ctx, models.BriefingMorning)
			return err
		},
	})

	// Midday pulse at 12:00 UTC
	s.AddJob(&Job{
		Name: "midday-pulse",
		Schedule: Schedule{
			Type:   ScheduleDaily,
			Hour:   12,
			Minute: 0,
		},
		Handler: func(ctx context.Context) error {
			_, err := s.generator.GenerateBriefing(ctx, models.BriefingMidday)
			return err
		},
	})

	// Evening wrap at 18:00 UTC
	s.AddJob(&Job{
		Name: "evening-wrap",
		Schedule: Schedule{
			Type:   ScheduleDaily,
			Hour:   18,
			Minute: 0,
		},
		Handler: func(ctx context.Context) error {
			_, err := s.generator.GenerateBriefing(ctx, models.BriefingEvening)
			return err
		},
	})

	// Weekly digest on Monday at 10:00 UTC
	s.AddJob(&Job{
		Name: "weekly-digest",
		Schedule: Schedule{
			Type:   ScheduleWeekly,
			Hour:   10,
			Minute: 0,
			Days:   []int{1}, // Monday
		},
		Handler: func(ctx context.Context) error {
			_, err := s.generator.GenerateBriefing(ctx, models.BriefingWeekly)
			return err
		},
	})

	// Trending update every 2 hours
	s.AddJob(&Job{
		Name: "trending-update",
		Schedule: Schedule{
			Type:     ScheduleInterval,
			Interval: 2 * time.Hour,
		},
		Handler: func(ctx context.Context) error {
			_, err := s.generator.GenerateTrending(ctx, 10)
			return err
		},
	})

	// Category digests - one per category per day, staggered
	categories := []string{"crypto", "politics", "tech", "sports", "finance"}
	for i, cat := range categories {
		category := cat // capture for closure
		hour := 9 + i   // Stagger: 9:00, 10:00, 11:00, etc.

		s.AddJob(&Job{
			Name: category + "-digest",
			Schedule: Schedule{
				Type:   ScheduleDaily,
				Hour:   hour,
				Minute: 30,
			},
			Handler: func(ctx context.Context) error {
				_, err := s.generator.GenerateCategoryDigest(ctx, category, 10)
				return err
			},
		})
	}
}

// AddJob adds a job to the scheduler.
func (s *Scheduler) AddJob(job *Job) {
	s.jobsMux.Lock()
	defer s.jobsMux.Unlock()

	job.NextRun = s.calculateNextRun(job.Schedule)
	s.jobs = append(s.jobs, job)

	log.Info().
		Str("job", job.Name).
		Time("next_run", job.NextRun).
		Msg("Job registered")
}

// Start begins the scheduler.
func (s *Scheduler) Start() {
	log.Info().Int("jobs", len(s.jobs)).Msg("Starting scheduler")

	// Start the job executor
	s.wg.Add(1)
	go s.jobLoop()

	// Start the event processor
	if s.eventChan != nil {
		s.wg.Add(1)
		go s.eventLoop()
	}
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() {
	log.Info().Msg("Stopping scheduler")
	s.cancel()
	s.wg.Wait()
}

// jobLoop checks and runs scheduled jobs.
func (s *Scheduler) jobLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkAndRunJobs()
		}
	}
}

// checkAndRunJobs runs any jobs that are due.
func (s *Scheduler) checkAndRunJobs() {
	now := time.Now().UTC()

	s.jobsMux.Lock()
	defer s.jobsMux.Unlock()

	for _, job := range s.jobs {
		if now.After(job.NextRun) || now.Equal(job.NextRun) {
			go s.runJob(job)
			job.LastRun = now
			job.NextRun = s.calculateNextRun(job.Schedule)

			log.Debug().
				Str("job", job.Name).
				Time("next_run", job.NextRun).
				Msg("Job scheduled for next run")
		}
	}
}

// runJob executes a job.
func (s *Scheduler) runJob(job *Job) {
	log.Info().Str("job", job.Name).Msg("Running job")

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	if err := job.Handler(ctx); err != nil {
		log.Error().Err(err).Str("job", job.Name).Msg("Job failed")
	} else {
		log.Info().Str("job", job.Name).Msg("Job completed")
	}
}

// calculateNextRun calculates the next run time for a schedule.
func (s *Scheduler) calculateNextRun(schedule Schedule) time.Time {
	now := time.Now().UTC()

	switch schedule.Type {
	case ScheduleInterval:
		return now.Add(schedule.Interval)

	case ScheduleDaily:
		next := time.Date(now.Year(), now.Month(), now.Day(),
			schedule.Hour, schedule.Minute, 0, 0, time.UTC)
		if next.Before(now) || next.Equal(now) {
			next = next.Add(24 * time.Hour)
		}
		return next

	case ScheduleWeekly:
		next := time.Date(now.Year(), now.Month(), now.Day(),
			schedule.Hour, schedule.Minute, 0, 0, time.UTC)

		// Find next matching day
		for i := 0; i < 7; i++ {
			dayOfWeek := int(next.Weekday())
			for _, d := range schedule.Days {
				if d == dayOfWeek && next.After(now) {
					return next
				}
			}
			next = next.Add(24 * time.Hour)
		}
		return next

	default:
		return now.Add(time.Hour)
	}
}

// eventLoop processes events from the syncer.
func (s *Scheduler) eventLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return

		case event, ok := <-s.eventChan:
			if !ok {
				return
			}
			s.processEvent(event)
		}
	}
}

// processEvent handles a market event and generates content if appropriate.
func (s *Scheduler) processEvent(event syncer.Event) {
	log.Debug().
		Str("type", string(event.Type)).
		Str("market", event.Market.Question).
		Msg("Processing event")

	ctx, cancel := context.WithTimeout(s.ctx, 2*time.Minute)
	defer cancel()

	switch event.Type {
	case syncer.EventBreakingMove:
		// Generate breaking news for significant movements
		if _, err := s.generator.GenerateBreaking(ctx, event); err != nil {
			log.Error().Err(err).Msg("Failed to generate breaking article")
		}

	case syncer.EventNewMarket:
		// Generate article for new high-volume markets
		if event.Market.Volume24h >= 50000 {
			if _, err := s.generator.GenerateNewMarket(ctx, event.Market); err != nil {
				log.Error().Err(err).Msg("Failed to generate new market article")
			}
		}

	case syncer.EventThresholdCross:
		// Generate article when market crosses key thresholds
		threshold := event.Metadata["threshold"].(float64)
		if threshold >= 0.75 || threshold <= 0.25 {
			// Only for extreme thresholds
			if _, err := s.generator.GenerateBreaking(ctx, event); err != nil {
				log.Error().Err(err).Msg("Failed to generate threshold article")
			}
		}

	case syncer.EventVolumeSpike:
		// Could generate article for volume spikes
		log.Info().
			Str("market", event.Market.Question).
			Float64("multiplier", event.Metadata["multiplier"].(float64)).
			Msg("Volume spike detected")
	}
}

// RunJobNow runs a specific job immediately by name.
func (s *Scheduler) RunJobNow(name string) error {
	s.jobsMux.RLock()
	defer s.jobsMux.RUnlock()

	for _, job := range s.jobs {
		if job.Name == name {
			go s.runJob(job)
			return nil
		}
	}

	return nil
}

// GetJobStatus returns the status of all jobs.
func (s *Scheduler) GetJobStatus() []map[string]interface{} {
	s.jobsMux.RLock()
	defer s.jobsMux.RUnlock()

	status := make([]map[string]interface{}, len(s.jobs))
	for i, job := range s.jobs {
		status[i] = map[string]interface{}{
			"name":     job.Name,
			"last_run": job.LastRun,
			"next_run": job.NextRun,
		}
	}
	return status
}
