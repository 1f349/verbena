package committer

import (
	"context"
	"github.com/1f349/verbena/internal/builder"
	"github.com/1f349/verbena/internal/database"
	"github.com/1f349/verbena/logger"
	"os/exec"
	"time"
)

type Committer struct {
	db      *database.Queries
	tick    time.Duration
	primary bool
	b       *builder.Builder
}

func New(db *database.Queries, tick time.Duration, primary bool, b *builder.Builder) *Committer {
	return &Committer{
		db:      db,
		tick:    tick,
		primary: primary,
		b:       b,
	}
}

func (c *Committer) Start() {
	if c.primary {
		go c.internalTick()
	}
}

func (c *Committer) internalTick() {
	t := time.NewTicker(c.tick)
	for {
		select {
		case <-t.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			zones, err := c.db.GetActiveZones(ctx)
			cancel()
			if err != nil {
				logger.Logger.Error("Failed to get list of active zones")
				return
			}
			for _, i := range zones {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				err = c.Commit(ctx, i)
				cancel()
				if err != nil {
					logger.Logger.Error("Failed to commit a zone", "zone id", i.ID, "zone name", i.Name, "err", err)
				}
			}
		}
	}
}

func (c *Committer) Commit(ctx context.Context, zone database.Zone) error {
	err := c.db.UseTx(ctx, func(tx *database.Queries) error {
		rowsUpdated, err := tx.CommitZoneRecords(ctx, zone.ID)
		if err != nil {
			return err
		}
		rowsDeleted, err := tx.CommitDeletedZoneRecords(ctx, zone.ID)
		if err != nil {
			return err
		}
		if rowsUpdated+rowsDeleted > 0 {
			err = tx.UpdateZoneSerial(ctx, zone.ID)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = c.b.Generate(ctx, zone)
	if err != nil {
		return err
	}

	err = c.bindReload(ctx, zone)
	if err != nil {
		return err
	}

	return c.bindNotify(ctx, zone)
}

func (c *Committer) bindReload(ctx context.Context, zone database.Zone) error {
	return exec.CommandContext(ctx, "/usr/sbin/rndc", "reload", zone.Name).Run()
}

func (c *Committer) bindNotify(ctx context.Context, zone database.Zone) error {
	return exec.CommandContext(ctx, "/usr/sbin/rndc", "notify", zone.Name).Run()
}
