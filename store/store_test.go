package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyIterator(t *testing.T) {
	iter := rowIter{
		pos:  0,
		rows: make([]IpGeoInfo, 0),
	}
	require.False(t, iter.Next())
	_, err := iter.Values()
	require.ErrorIs(t, err, errRowIterEnd)
}

func TestOneItemIterator(t *testing.T) {
	iter := rowIter{
		pos:  0,
		rows: make([]IpGeoInfo, 1),
	}
	require.True(t, iter.Next())
	_, err := iter.Values()
	require.ErrorIs(t, err, nil)
	require.False(t, iter.Next())
	_, err = iter.Values()
	require.ErrorIs(t, err, errRowIterEnd)
}

func TestOneItemValuesIterator(t *testing.T) {
	want := []interface{}{"8.8.8.8", "CC", "Country", "City", 0.1, 0.1, "mystery"}

	g := IpGeoInfo{
		IpAdress:     "8.8.8.8",
		CountryCode:  "CC",
		Country:      "Country",
		City:         "City",
		Latitude:     0.1,
		Longitude:    0.1,
		MysteryValue: "mystery",
	}

	iter := rowIter{
		pos: 0,
		rows: []IpGeoInfo{
			g,
		},
	}

	require.True(t, iter.Next())
	got, err := iter.Values()
	require.ErrorIs(t, err, nil)
	require.ElementsMatch(t, got, want)
}
