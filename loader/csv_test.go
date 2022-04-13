package loader

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/k911mipt/geolocation/store"
)

const csvSource = `
ip_address,country_code,country,city,latitude,longitude,mystery_value
200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
160.103.7.140,CZ,Nicaragua,New Neva,-68.31023296602508,-37.62435199624531,7301823115
70.95.73.73,TL,Saudi Arabia,Gradymouth,-49.16675918861615,-86.05920084416894,2559997162
70.95.73.73,IL,Turkey,Lake Osvaldo,-27.03246415903554,118.74210995605495,8971989993
,PY,Falkland Islands (Malvinas),,75.41685191518815,-144.6943217219469,0`

type storeMock struct {
}

func (s *storeMock) InsertGeoInfos(ctx context.Context, rows []store.IpGeoInfo) (uint64, error) {
	return 2, nil
}

func TestImportData(t *testing.T) {
	ioReader := strings.NewReader(csvSource)
	storeMock := storeMock{}
	loader := NewFromCSV(&storeMock)
	stats, err := loader.Run(context.Background(), ioReader, "", 100)
	if err != nil {
		fmt.Print(err)
		t.Error(err)
	}
	assert.Equal(t, uint64(2), stats.Uploaded)
	assert.Equal(t, uint64(1), stats.DuplicatedInDB)
	assert.Equal(t, uint64(1), stats.DuplicatedInCsv)
	assert.Equal(t, uint64(2), stats.BadCsvRecords)
}
