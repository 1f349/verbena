package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/1f349/verbena/internal/bind"
	"github.com/1f349/verbena/internal/database"
	"github.com/1f349/verbena/internal/zone"
	"github.com/1f349/verbena/logger"
	"github.com/charmbracelet/log"
	"github.com/gobuffalo/nulls"
)

type committerQueries interface {
	GetZoneActiveRecords(ctx context.Context, zoneID int64) ([]database.Record, error)
	GetActiveZones(ctx context.Context) ([]database.Zone, error)
}

type Builder struct {
	db          committerQueries
	genTick     time.Duration
	dir         string
	bindGenConf string
	nameservers []string
	genLock     sync.Mutex
}

func New(db committerQueries, genTick time.Duration, dir string, bindGenConf string, nameservers []string) (*Builder, error) {
	if len(nameservers) < 3 {
		return nil, fmt.Errorf("at least 3 nameservers are required")
	}
	return &Builder{
		db:          db,
		genTick:     genTick,
		dir:         dir,
		bindGenConf: bindGenConf,
		nameservers: nameservers,
	}, nil
}

func (b *Builder) Start() {
	go b.internalTicker()
}

func (b *Builder) internalTicker() {
	var loadedZones []string

	t := time.NewTicker(b.genTick)
	for {
		select {
		case <-t.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			zones, err := b.db.GetActiveZones(ctx)
			cancel()
			if err != nil {
				logger.Logger.Error("Failed to get list of active zones")
				return
			}

			var newLoadedZones []string
			for _, i := range zones {
				err = b.Generate(context.Background(), i)
				if err != nil {
					logger.Logger.Error("Failed to generate a zone", "zone id", i.ID, "zone name", i.Name, "err", err)
				}
				newLoadedZones = append(newLoadedZones, i.Name)
			}

			slices.Sort(newLoadedZones)

			// If the currently loaded zones and new loaded zones
			if !slices.Equal(newLoadedZones, loadedZones) {
				err = b.generateLocalGeneratedConfig(newLoadedZones)
				if err != nil {
					logger.Logger.Error("Failed to generate locally generated config")
				} else {
					loadedZones = newLoadedZones
				}
			}
		}
	}
}

func (b *Builder) Generate(ctx context.Context, zoneInfo database.Zone) error {
	b.genLock.Lock()
	defer b.genLock.Unlock()

	records, err := b.db.GetZoneActiveRecords(ctx, zoneInfo.ID)
	if err != nil {
		return err
	}

	zoneRecords := make([]zone.Record, 0, len(records)+len(b.nameservers))

	for _, i := range b.nameservers {
		zoneRecords = append(zoneRecords, zone.Record{
			Name:       "",
			TimeToLive: nulls.UInt32{},
			Type:       zone.NS,
			Value:      i,
		})
	}

	for _, i := range records {
		var ty zone.RecordType
		switch i.Type {
		case "NS":
			ty = zone.NS
		case "MX":
			ty = zone.MX
		case "A":
			ty = zone.A
		case "AAAA":
			ty = zone.AAAA
		case "CNAME":
			ty = zone.CNAME
		case "TXT":
			ty = zone.TXT
		case "SRV":
			ty = zone.SRV
		case "CAA":
			ty = zone.CAA
		default:
			return fmt.Errorf("unknown type: %s", i.Type)
		}
		zoneRecords = append(zoneRecords, zone.Record{
			Name: i.Name,
			TimeToLive: nulls.UInt32{
				UInt32: uint32(i.Ttl.Int32),
				Valid:  i.Ttl.Valid,
			},
			Type:  ty,
			Value: i.Value,
		})
	}

	zoneFileName := filepath.Join(b.dir, zoneInfo.Name+".zone")
	zoneFileTemp := filepath.Join(b.dir, zoneInfo.Name+".zone.temp")

	zoneFile, err := os.Create(zoneFileTemp)
	if err != nil {
		return err
	}
	defer zoneFile.Close()
	defer os.Remove(zoneFileTemp)

	err = zone.WriteZone(zoneFile, zoneInfo.Name, uint32(zoneInfo.Ttl), zone.SoaRecord{
		Nameserver: b.nameservers[0],
		Admin:      zoneInfo.Admin,
		Serial:     uint32(zoneInfo.Serial),
		Refresh:    uint32(zoneInfo.Refresh),
		Retry:      uint32(zoneInfo.Retry),
		Expire:     uint32(zoneInfo.Expire),
		TimeToLive: uint32(zoneInfo.Ttl),
	}, zoneRecords)
	if err != nil {
		return err
	}

	err = zoneFile.Close()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "/usr/bin/named-checkzone", zoneInfo.Name, zoneFileTemp)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if logger.Logger.GetLevel() >= log.DebugLevel {
			err = fmt.Errorf("named-checkzone failed with output: %w: %s", err, string(out))
		}
		return err
	}

	return os.Rename(zoneFileTemp, zoneFileName)
}

func (b *Builder) generateLocalGeneratedConfig(zones []string) error {
	bindLocalTempPath := b.bindGenConf + ".temp"
	bindLocalTemp, err := os.Create(bindLocalTempPath)
	if err != nil {
		return err
	}
	defer bindLocalTemp.Close()
	defer os.Remove(bindLocalTempPath)

	err = bind.WriteBindConfig(bindLocalTemp, b.dir, zones)
	if err != nil {
		return err
	}

	return os.Rename(bindLocalTempPath, b.bindGenConf)
}
