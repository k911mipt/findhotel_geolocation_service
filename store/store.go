package store

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrNotFound = errors.New("not found")

type IpGeoInfo struct {
	IpAdress     string  `json:"ip_address"`
	CountryCode  string  `json:"country_code,omitempty"`
	Country      string  `json:"country,omitempty"`
	City         string  `json:"city,omitempty"`
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
	MysteryValue string  `json:"mystery_value,omitempty"`
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(ctx context.Context, connStr string) (*Store, error) {
	pgxpoolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	pgxpoolConfig.ConnConfig.PreferSimpleProtocol = true
	pool, err := pgxpool.ConnectConfig(ctx, pgxpoolConfig)
	return &Store{pool: pool}, err
}

func (st *Store) FetchGeoInfo(ctx context.Context, ip string) (IpGeoInfo, error) {
	var result IpGeoInfo
	row := st.pool.QueryRow(ctx, `
		SELECT ip, country_code, country, city, latitude, longitude, mystery_value
		FROM ip_geoinfo
		WHERE ip = $1
	`, ip)
	err := row.Scan(&result.IpAdress, &result.CountryCode, &result.Country, &result.City, &result.Latitude, &result.Longitude, &result.MysteryValue)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, ErrNotFound
	}
	return result, err
}

func (st *Store) InsertGeoInfos(ctx context.Context, rows []IpGeoInfo) (uint64, error) {
	var rowsAffected uint64
	err := st.executeTransaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			CREATE TEMP TABLE tmp_ip_geoinfo
			ON COMMIT DROP
			AS (SELECT * FROM ip_geoinfo) WITH NO DATA`)
		if err != nil {
			return err
		}
		rowSrc := &rowIter{
			rows: rows,
		}
		_, err = tx.CopyFrom(ctx,
			[]string{"tmp_ip_geoinfo"},
			[]string{"ip", "country_code", "country", "city", "latitude", "longitude", "mystery_value"}, rowSrc)
		if err != nil {
			return err
		}
		log.Printf("Made bulk insert to temp table")

		cmd, err := tx.Exec(ctx, `
			INSERT INTO ip_geoinfo
			SELECT *
			FROM tmp_ip_geoinfo
			ON CONFLICT (ip) DO NOTHING
		`)
		log.Printf("Copied temp table to ip_geoinfo")
		rowsAffected = uint64(cmd.RowsAffected())

		return err
	})

	return rowsAffected, err
}

func (st *Store) executeTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	var tx pgx.Tx
	var err error
	tx, err = st.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if !(rollbackErr == nil || errors.Is(rollbackErr, pgx.ErrTxClosed)) {
			err = rollbackErr
		}
	}()

	fErr := fn(tx)
	if fErr != nil {
		_ = tx.Rollback(ctx)
		return fErr
	}

	return tx.Commit(ctx)
}

var errRowIterEnd = errors.New("no more rows")

type rowIter struct {
	pos  int
	rows []IpGeoInfo
}

func (i *rowIter) Next() bool {
	return i.pos < len(i.rows)
}

func (i *rowIter) Values() ([]interface{}, error) {
	if len(i.rows) <= i.pos {
		return nil, errRowIterEnd
	}
	r := i.rows[i.pos]
	i.pos++
	return []interface{}{r.IpAdress, r.CountryCode, r.Country, r.City, r.Latitude, r.Longitude, r.MysteryValue}, nil
}

func (i rowIter) Err() error {
	return nil
}
