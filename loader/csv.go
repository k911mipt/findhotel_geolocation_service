package loader

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"strconv"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/k911mipt/geolocation/store"
)

var errBadRecord = errors.New("bad csv record")

type Stats struct {
	DateStart       time.Time
	Elapsed         time.Duration
	Uploaded        uint64
	DuplicatedInCsv uint64
	DuplicatedInDB  uint64
	BadCsvRecords   uint64
}

func (s *Stats) Finish() {
	s.Elapsed = time.Since(s.DateStart)
}

func (s *Stats) Print(w io.Writer) error {
	_, err := fmt.Fprintf(w,
		`Import statistics:
	Added new records: %v
	Already present in DB: %v
	Duplicated records in CSV: %v
	Bad CSV records: %v
	Time spent: %v
`,
		s.Uploaded,
		s.DuplicatedInDB,
		s.DuplicatedInCsv,
		s.BadCsvRecords,
		s.Elapsed)
	return err
}

type Loader struct {
	db DBClient
}

func NewFromCSV(db DBClient) *Loader {
	return &Loader{
		db: db,
	}
}

type DBClient interface {
	InsertGeoInfos(ctx context.Context, rows []store.IpGeoInfo) (uint64, error)
}

func (l *Loader) Run(ctx context.Context, file io.Reader, filePath string, batchSize int) (*Stats, error) {
	stats := Stats{
		DateStart: time.Now(),
	}
	defer stats.Finish()

	batch := newRowBatch(batchSize)

	csvReader := csv.NewReader(file)
	csvReader.ReuseRecord = true
	csvReader.FieldsPerRecord = 7
	log.Printf("Starting import from %s", filePath)
	for {
		select {
		case <-ctx.Done():
			return &stats, fmt.Errorf("Import canceled")
		default:
			var ipGeoInfo store.IpGeoInfo
			err := readIpGeoInfo(csvReader, &ipGeoInfo)

			if err == io.EOF {
				err = l.uploadBatch(ctx, batch, &stats)
				if err != nil {
					log.Println("Import finished")
				}
				return &stats, err
			}
			if err == errBadRecord {
				stats.BadCsvRecords++
				continue
			}
			if err != nil {
				return &stats, err
			}

			if !batch.TryAdd(&ipGeoInfo) {
				stats.DuplicatedInCsv++
			}

			if batch.IsFull() {
				err = l.uploadBatch(ctx, batch, &stats)
				if err != nil {
					return &stats, err
				}
			}
		}
	}
}

func (l *Loader) uploadBatch(ctx context.Context, batch *rowBatch, stats *Stats) error {
	uploaded, err := l.db.InsertGeoInfos(ctx, batch.data)
	if err != nil {
		return err
	}
	stats.Uploaded += uploaded
	stats.DuplicatedInDB += uint64(len(batch.data)) - uploaded
	log.Printf("%v/%v rows added to DB", uploaded, len(batch.data))
	batch.Flush()
	return nil
}

type IPSet map[string]struct{}

type rowBatch struct {
	size  int
	index IPSet
	data  []store.IpGeoInfo
}

func newRowBatch(size int) *rowBatch {
	return &rowBatch{
		size:  size,
		index: make(IPSet),
		data:  make([]store.IpGeoInfo, 0, size),
	}
}

func (b *rowBatch) TryAdd(ipGeoInfo *store.IpGeoInfo) bool {
	if _, ok := b.index[ipGeoInfo.IpAdress]; ok {
		return false
	}
	b.index[ipGeoInfo.IpAdress] = struct{}{}
	b.data = append(b.data, *ipGeoInfo)
	return true
}

func (b *rowBatch) IsFull() bool {
	return len(b.data) >= b.size
}

func (b *rowBatch) Flush() {
	b.data = b.data[:0]
}

func readIpGeoInfo(csvReader *csv.Reader, ipgeoInfo *store.IpGeoInfo) error {
	record, err := csvReader.Read()
	if err != nil {
		return err
	}
	err = parseCsvRecord(record, ipgeoInfo)
	if err != nil {
		return errBadRecord // skip parse error processing
	}
	return nil
}

var ErrInvalidRecordSize = errors.New("Invalid row length")
var ErrInvalidIP = errors.New("Invalid IP address")
var ErrInvalidCountyCode = errors.New("Invalid country code")
var ErrInvalidCountry = errors.New("Invalid country")
var ErrInvalidCity = errors.New("Invalid city")
var ErrInvalidLatitude = errors.New("Invalid latitude")
var ErrInvalidLongitude = errors.New("Invalid longitude")

func parseCsvRecord(record []string, ipGeoInfo *store.IpGeoInfo) error {
	var errs *multierror.Error
	if len(record) != 7 {
		errs = multierror.Append(errs, ErrInvalidRecordSize)
	}

	ipGeoInfo.IpAdress = record[0]
	if net.ParseIP(ipGeoInfo.IpAdress) == nil {
		errs = multierror.Append(errs, ErrInvalidIP)
	}

	ipGeoInfo.CountryCode = record[1]
	if len(ipGeoInfo.CountryCode) != 2 {
		errs = multierror.Append(errs, ErrInvalidCountyCode)
	}

	ipGeoInfo.Country = record[2]
	if ipGeoInfo.Country == "" {
		errs = multierror.Append(errs, ErrInvalidCountry)
	}

	ipGeoInfo.City = record[3]
	if ipGeoInfo.City == "" {
		errs = multierror.Append(errs, ErrInvalidCity)
	}

	latitude, err := strconv.ParseFloat(record[4], 64)
	if err != nil || math.Abs(latitude) > 90 {
		errs = multierror.Append(errs, ErrInvalidLatitude)
	}
	ipGeoInfo.Latitude = latitude

	longitude, err := strconv.ParseFloat(record[5], 64)
	if err != nil || math.Abs(longitude) > 180 {
		errs = multierror.Append(errs, ErrInvalidLongitude)
	}
	ipGeoInfo.Longitude = longitude

	ipGeoInfo.MysteryValue = record[6]

	return errs.ErrorOrNil()
}
