// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserIntegration(t *testing.T) {
	input := "https://playbyplay.sport5.co.il/?GameID=125423&FLNum=16"
	expected := "https://rgevod.akamaized.net/vodedge/_definst_/mp4:rge/bynet/sport5/sport5/PRV5/5s5NQgT0ng/App/NM_VTR_TAK_KASH_131122_1800.mp4/chunklist_b1800000.m3u8"
	p := NewParser()
	actual, err := p.FindDownloadLink(input)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}
