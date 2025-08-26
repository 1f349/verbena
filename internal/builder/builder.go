package builder

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/1f349/verbena/conf"
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
	cmd         conf.CmdConf
}

func New(db committerQueries, genTick time.Duration, dir string, bindGenConf string, nameservers []string, cmd conf.CmdConf) (*Builder, error) {
	if len(nameservers) < 3 {
		return nil, fmt.Errorf("at least 3 nameservers are required")
	}
	return &Builder{
		db:          db,
		genTick:     genTick,
		dir:         dir,
		bindGenConf: bindGenConf,
		nameservers: nameservers,
		cmd:         cmd,
	}, nil
}

func (b *Builder) Start() {
	go b.internalTicker()
}

func (b *Builder) internalTicker() {
	var loadedZones []string
	var v4loadedReverseZones []netip.Prefix
	var v6loadedReverseZones []netip.Prefix
	// TODO: figure out how to handle reverse zones here

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
			var newV4loadedReverseZones []netip.Prefix
			var newV6loadedReverseZones []netip.Prefix

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
				err = b.generateLocalGeneratedConfig(ctx, newLoadedZones)
				if err != nil {
					logger.Logger.Error("Failed to generate locally generated config")
				} else {
					loadedZones = newLoadedZones
				}

				// TODO: add function to generate stubs for used prefixes
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
		ty := zone.RecordTypeFromString(i.Type)
		if !ty.IsValid() {
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

	return b.generateZoneFile(ctx, zoneInfo.Name, zone.SoaRecord{
		Nameserver: b.nameservers[0],
		Admin:      zoneInfo.Admin,
		Serial:     uint32(zoneInfo.Serial),
		Refresh:    uint32(zoneInfo.Refresh),
		Retry:      uint32(zoneInfo.Retry),
		Expire:     uint32(zoneInfo.Expire),
		TimeToLive: uint32(zoneInfo.Ttl),
	}, zoneRecords)
}

func (b *Builder) generatePrefixStubs(ctx context.Context, zoneInfo database.Zone, prefix netip.Prefix) error {
	zoneRecords := zone.Rfc2317CNAMEs(prefix, nulls.NewUInt32(uint32(zoneInfo.Ttl)))

	// Prefix stubs are not required
	if zoneRecords == nil {
		return nil
	}

	nsRecords := make([]zone.Record, 0, len(b.nameservers))
	for i, ns := range b.nameservers {
		nsRecords[i] = zone.Record{
			Name:       "",
			TimeToLive: nulls.UInt32{},
			Type:       zone.NS,
			Value:      ns,
		}
	}

	zoneRecords = append(nsRecords, zoneRecords...)

	b.genLock.Lock()
	defer b.genLock.Unlock()

	return b.generateZoneFile(ctx, prefix.Addr().String()+"_"+strconv.Itoa(prefix.Bits()), zone.SoaRecord{
		Nameserver: b.nameservers[0],
		Admin:      zoneInfo.Admin,
		Serial:     uint32(zoneInfo.Serial),
		Refresh:    uint32(zoneInfo.Refresh),
		Retry:      uint32(zoneInfo.Retry),
		Expire:     uint32(zoneInfo.Expire),
		TimeToLive: uint32(zoneInfo.Ttl),
	}, zoneRecords)
}

func (b *Builder) generateZoneFile(ctx context.Context, zoneName string, soaRecord zone.SoaRecord, zoneRecords []zone.Record) error {
	zoneFileName := filepath.Join(b.dir, zoneName+".zone")
	zoneFileTemp := filepath.Join(b.dir, zoneName+".zone.temp")

	zoneFile, err := os.Create(zoneFileTemp)
	if err != nil {
		return err
	}
	defer zoneFile.Close()
	defer os.Remove(zoneFileTemp)

	err = zone.WriteZone(zoneFile, zoneName, soaRecord.TimeToLive, soaRecord, zoneRecords)
	if err != nil {
		return err
	}

	err = zoneFile.Close()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, b.cmd.CheckZone, zoneName, zoneFileTemp)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if logger.Logger.GetLevel() >= log.DebugLevel {
			err = fmt.Errorf("named-checkzone failed with output: %w: %s", err, string(out))
		}
		return err
	}

	err = os.Rename(zoneFileTemp, zoneFileName)
	if err != nil {
		return err
	}

	return b.bindReloadZone(ctx, zoneName)
}

func (b *Builder) generateLocalGeneratedConfig(ctx context.Context, zones []string) error {
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

	err = os.Rename(bindLocalTempPath, b.bindGenConf)
	if err != nil {
		return err
	}

	return b.bindReload(ctx)
}

func (b *Builder) bindReload(ctx context.Context) error {
	return exec.CommandContext(ctx, b.cmd.Rndc, "reload").Run()
}

func (b *Builder) bindReloadZone(ctx context.Context, zoneName string) error {
	return exec.CommandContext(ctx, b.cmd.Rndc, "reload", zoneName).Run()
}
