package metadata

import (
	"encoding/hex"
	"testing"
)

func TestStartWithSync(t *testing.T) {
	hexHeader := "49443322"
	data, _ := hex.DecodeString(hexHeader)

	isSync := StartsWithIdV3Sync(data)

	if !isSync {
		t.Error("This is a valid sync.")
	}
}

func TestStartWithSyncFailed(t *testing.T) {
	hexHeader := "45443322"
	data, _ := hex.DecodeString(hexHeader)

	isSync := StartsWithIdV3Sync(data)

	if isSync {
		t.Error("This is not a valid sync.")
	}
}

func TestStartWithSyncShort(t *testing.T) {
	hexHeader := "4944"
	data, _ := hex.DecodeString(hexHeader)

	isSync := StartsWithIdV3Sync(data)

	if isSync {
		t.Error("Data too short.")
	}
}

func TestHasFooter(t *testing.T) {
	hexHeader := "454433228713"
	data, _ := hex.DecodeString(hexHeader)

	hasFooter := HasIdV3Footer(data)

	if !hasFooter {
		t.Error("This tag has footer.")
	}
}

func TestHasFooterFail(t *testing.T) {
	hexHeader := "454433228723"
	data, _ := hex.DecodeString(hexHeader)

	hasFooter := HasIdV3Footer(data)

	if hasFooter {
		t.Error("This tag has no footer.")
	}
}

func TestHasFooterShort(t *testing.T) {
	hexHeader := "4544"
	data, _ := hex.DecodeString(hexHeader)

	hasFooter := HasIdV3Footer(data)

	if hasFooter {
		t.Error("This tag is too short.")
	}
}
