package domain_test

import (
	"drawwwingame/domain"
	"testing"
)

func TestEncryptAndDecrypt(t *testing.T) {
	patterns := []string{
		"test",
		"qwertyuiohgfdsasdfgh1234-5678SDFVFBGNHMTYJTERdsvbfgnrethrs",
		"SADVBFsdfb2345",
		";*234t5v ;",
		"sdafhgfj,",
		"sdsh3y5u65vf d",
		"a",
		"ty5476ikymnwrbe",
		"2rwetg",
	}
	for _, v := range patterns {
		str, _ := domain.Encrypt(v)
		val, _ := domain.Decrypt(str)
		check(t, v, val, v)
	}
}
