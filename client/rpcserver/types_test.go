// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

package rpcserver

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"decred.org/dcrdex/dex/encode"
)

func TestCheckNArgs(t *testing.T) {
	tests := []struct {
		name      string
		have      []string
		wantNArgs []int
		wantErr   bool
	}{{
		name:      "ok exact",
		have:      []string{"1", "2", "3"},
		wantNArgs: []int{3},
		wantErr:   false,
	}, {
		name:      "ok between",
		have:      []string{"1", "2", "3"},
		wantNArgs: []int{2, 4},
		wantErr:   false,
	}, {
		name:      "ok lower",
		have:      []string{"1", "2"},
		wantNArgs: []int{2, 4},
		wantErr:   false,
	}, {
		name:      "ok upper",
		have:      []string{"1", "2", "3", "4"},
		wantNArgs: []int{2, 4},
		wantErr:   false,
	}, {
		name:      "not exact",
		have:      []string{"1", "2", "3"},
		wantNArgs: []int{2},
		wantErr:   true,
	}, {
		name:      "too few",
		have:      []string{"1", "2"},
		wantNArgs: []int{3, 5},
		wantErr:   true,
	}, {
		name:      "too many",
		have:      []string{"1", "2", "3", "4", "5", "6"},
		wantNArgs: []int{2, 5},
		wantErr:   true,
	}}
	for _, test := range tests {
		pwArgs := make([]encode.PassBytes, len(test.have))
		for i, testValue := range test.have {
			pwArgs[i] = encode.PassBytes(testValue)
		}
		err := checkNArgs(&RawParams{PWArgs: pwArgs, Args: test.have}, test.wantNArgs, test.wantNArgs)
		if err != nil {
			if test.wantErr {
				continue
			}
			t.Fatalf("unexpected error for test %s: %v", test.name, err)
		}
		if test.wantErr {
			t.Fatalf("expected error for test %s", test.name)
		}
	}
}

func TestParseNewWalletArgs(t *testing.T) {
	paramsWithAssetID := func(id string) *RawParams {
		pw := encode.PassBytes("password123")
		pwArgs := []encode.PassBytes{pw, pw}
		args := []string{
			id,
			"default",
			"rpclisten=127.0.0.0",
		}
		return &RawParams{PWArgs: pwArgs, Args: args}
	}
	tests := []struct {
		name    string
		params  *RawParams
		wantErr error
	}{{
		name:   "ok",
		params: paramsWithAssetID("42"),
	}, {
		name:    "assetID is not int",
		params:  paramsWithAssetID("42.1"),
		wantErr: errArgs,
	}}
	for _, test := range tests {
		nwf, err := parseNewWalletArgs(test.params)
		if err != nil {
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("unexpected error %v for test %s",
					err, test.name)
			}
			continue
		}
		if !bytes.Equal(nwf.AppPass, test.params.PWArgs[0]) {
			t.Fatalf("appPass doesn't match")
		}
		if !bytes.Equal(nwf.WalletPass, test.params.PWArgs[1]) {
			t.Fatalf("walletPass doesn't match")
		}
		if fmt.Sprint(nwf.AssetID) != test.params.Args[0] {
			t.Fatalf("assetID doesn't match")
		}
		if nwf.Account != test.params.Args[1] {
			t.Fatalf("account doesn't match")
		}
		if nwf.ConfigText != test.params.Args[2] {
			t.Fatalf("config doesn't match")
		}
	}
}

func TestParseOpenWalletArgs(t *testing.T) {
	paramsWithAssetID := func(id string) *RawParams {
		pw := encode.PassBytes("password123")
		pwArgs := []encode.PassBytes{pw}
		args := []string{id}
		return &RawParams{PWArgs: pwArgs, Args: args}
	}
	tests := []struct {
		name    string
		params  *RawParams
		wantErr error
	}{{
		name:   "ok",
		params: paramsWithAssetID("42"),
	}, {
		name:    "assetID is not int",
		params:  paramsWithAssetID("42.1"),
		wantErr: errArgs,
	}}
	for _, test := range tests {
		owf, err := parseOpenWalletArgs(test.params)
		if err != nil {
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("unexpected error %v for test %s",
					err, test.name)
			}
			continue
		}
		if !bytes.Equal(owf.AppPass, test.params.PWArgs[0]) {
			t.Fatalf("appPass doesn't match")
		}
		if fmt.Sprint(owf.AssetID) != test.params.Args[0] {
			t.Fatalf("assetID doesn't match")
		}
	}
}

func TestCheckUIntArg(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    uint64
		bitSize int
		wantErr error
	}{{
		name:    "ok",
		arg:     "4294967295",
		want:    4294967295,
		bitSize: 32,
	}, {
		name:    "too big",
		arg:     "4294967296",
		bitSize: 32,
		wantErr: errArgs,
	}, {
		name:    "not int",
		arg:     "42.1",
		bitSize: 32,
		wantErr: errArgs,
	}, {
		name:    "negative",
		arg:     "-42",
		bitSize: 32,
		wantErr: errArgs,
	}}
	for _, test := range tests {
		res, err := checkUIntArg(test.arg, "name", test.bitSize)
		if err != nil {
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("unexpected error %v for test %s",
					err, test.name)
			}
			continue
		}
		if res != test.want {
			t.Fatalf("expected %d but got %d for test %q", test.want, res, test.name)
		}
	}
}

func TestCheckBoolArg(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    bool
		wantErr error
	}{{
		name: "ok string lower",
		arg:  "true",
		want: true,
	}, {
		name: "ok string upper",
		arg:  "False",
		want: false,
	}, {
		name: "ok int",
		arg:  "1",
		want: true,
	}, {
		name:    "string but not true or false",
		arg:     "blue",
		wantErr: errArgs,
	}, {
		name:    "int but not 0 or 1",
		arg:     "2",
		wantErr: errArgs,
	}}
	for _, test := range tests {
		res, err := checkBoolArg(test.arg, "name")
		if err != nil {
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("unexpected error %v for test %s",
					err, test.name)
			}
			continue
		}
		if res != test.want {
			t.Fatalf("wanted %v but got %v for test %v", test.want, res, test.name)
		}
	}
}

func TestParseRegisterArgs(t *testing.T) {
	paramsWithFee := func(fee string) *RawParams {
		pw := encode.PassBytes("password123")
		pwArgs := []encode.PassBytes{pw}
		args := []string{"dex", fee, "cert"}
		return &RawParams{PWArgs: pwArgs, Args: args}
	}
	tests := []struct {
		name    string
		params  *RawParams
		wantErr error
	}{{
		name:   "ok",
		params: paramsWithFee("1000"),
	}, {
		name:    "fee not int",
		params:  paramsWithFee("1000.0"),
		wantErr: errArgs,
	}}
	for _, test := range tests {
		reg, err := parseRegisterArgs(test.params)
		if err != nil {
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("unexpected error %v for test %s",
					err, test.name)
			}
			continue
		}
		if !bytes.Equal(reg.AppPass, test.params.PWArgs[0]) {
			t.Fatalf("appPass doesn't match")
		}
		if reg.Addr != test.params.Args[0] {
			t.Fatalf("url doesn't match")
		}
		if fmt.Sprint(reg.Fee) != test.params.Args[1] {
			t.Fatalf("fee doesn't match")
		}
		if fmt.Sprint(reg.Cert) != test.params.Args[2] {
			t.Fatalf("cert doesn't match")
		}
	}
}

func TestParseHelpArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *helpForm
		wantErr error
	}{{
		name: "ok no args",
		want: &helpForm{},
	}, {
		name: "ok help with",
		args: []string{"thing"},
		want: &helpForm{HelpWith: "thing"},
	}, {
		name: "ok help with include passwords",
		args: []string{"thing", "true"},
		want: &helpForm{HelpWith: "thing", IncludePasswords: true},
	}, {
		name:    "include passwords not boolean",
		args:    []string{"thing", "thing2"},
		wantErr: errArgs,
	}}
	for _, test := range tests {
		form, err := parseHelpArgs(&RawParams{Args: test.args})
		if err != nil {
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("unexpected error %v for test %s",
					err, test.name)
			}
			continue
		}
		if len(test.args) > 0 && form.HelpWith != test.args[0] {
			t.Fatalf("helpwith doesn't match")
		}
		if len(test.args) > 1 && fmt.Sprint(form.IncludePasswords) != test.args[1] {
			t.Fatalf("includepasswords doesn't match")
		}
	}
}

func TestTradeArgs(t *testing.T) {
	pw := encode.PassBytes("password123")
	goodParams := &RawParams{
		PWArgs: []encode.PassBytes{pw}, // 0. AppPass
		Args: []string{
			"1.2.3.4:3000", // 0. Host
			"true",         // 1. IsLimit
			"true",         // 2. Sell
			"0",            // 3. Base
			"42",           // 4. Quote
			"1",            // 5. Qty
			"1",            // 6. Rate
			"true",         // 7. TifNow
		}}
	paramsWith := func(idx int, thing string) *RawParams {
		newParams := &RawParams{
			PWArgs: make([]encode.PassBytes, 1),
			Args:   make([]string, 8),
		}
		copy(newParams.PWArgs, goodParams.PWArgs)
		copy(newParams.Args, goodParams.Args)
		newParams.Args[idx] = thing
		return newParams
	}
	tests := []struct {
		name    string
		params  *RawParams
		wantErr error
	}{{
		name:   "ok",
		params: goodParams,
	}, {
		name:    "isLimit not bool",
		params:  paramsWith(1, "blue"),
		wantErr: errArgs,
	}, {
		name:    "sell not bool",
		params:  paramsWith(2, "blue"),
		wantErr: errArgs,
	}, {
		name:    "base not uint32",
		params:  paramsWith(3, "-1"),
		wantErr: errArgs,
	}, {
		name:    "quote not uint32",
		params:  paramsWith(4, "-1"),
		wantErr: errArgs,
	}, {
		name:    "qty not uint64",
		params:  paramsWith(5, "-1"),
		wantErr: errArgs,
	}, {
		name:    "rate not uint64",
		params:  paramsWith(6, "-1"),
		wantErr: errArgs,
	}, {
		name:    "tifnow not bool",
		params:  paramsWith(7, "blue"),
		wantErr: errArgs,
	}}
	for _, test := range tests {
		reg, err := parseTradeArgs(test.params)
		if err != nil {
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("unexpected error %v for test %s",
					err, test.name)
			}
			continue
		}
		if !bytes.Equal(reg.AppPass, test.params.PWArgs[0]) {
			t.Fatalf("AppPass doesn't match")
		}
		if reg.SrvForm.Host != test.params.Args[0] {
			t.Fatalf("Host doesn't match")
		}
		if fmt.Sprint(reg.SrvForm.IsLimit) != test.params.Args[1] {
			t.Fatalf("IsLimit doesn't match")
		}
		if fmt.Sprint(reg.SrvForm.Sell) != test.params.Args[2] {
			t.Fatalf("Sell doesn't match")
		}
		if fmt.Sprint(reg.SrvForm.Base) != test.params.Args[3] {
			t.Fatalf("Base doesn't match")
		}
		if fmt.Sprint(reg.SrvForm.Quote) != test.params.Args[4] {
			t.Fatalf("Quote doesn't match")
		}
		if fmt.Sprint(reg.SrvForm.Qty) != test.params.Args[5] {
			t.Fatalf("Qty doesn't match")
		}
		if fmt.Sprint(reg.SrvForm.Rate) != test.params.Args[6] {
			t.Fatalf("Rate doesn't match")
		}
		if fmt.Sprint(reg.SrvForm.TifNow) != test.params.Args[7] {
			t.Fatalf("TifNow doesn't match")
		}
	}
}

func TestParseCancelArgs(t *testing.T) {
	paramsWithOrderID := func(orderID string) *RawParams {
		pw := encode.PassBytes("password123")
		pwArgs := []encode.PassBytes{pw}
		return &RawParams{PWArgs: pwArgs, Args: []string{orderID}}
	}
	tests := []struct {
		name    string
		params  *RawParams
		wantErr error
	}{{
		name:   "ok",
		params: paramsWithOrderID("fb94fe99e4e32200a341f0f1cb33f34a08ac23eedab636e8adb991fa76343e1e"),
	}, {
		name:    "order ID incorrect length",
		params:  paramsWithOrderID("94fe99e4e32200a341f0f1cb33f34a08ac23eedab636e8adb991fa76343e1e"),
		wantErr: errArgs,
	}, {
		name:    "order ID not hex",
		params:  paramsWithOrderID("zb94fe99e4e32200a341f0f1cb33f34a08ac23eedab636e8adb991fa76343e1e"),
		wantErr: errArgs,
	}}
	for _, test := range tests {
		reg, err := parseCancelArgs(test.params)
		if err != nil {
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("unexpected error %v for test %q",
					err, test.name)
			}
			continue
		}
		if !bytes.Equal(reg.AppPass, test.params.PWArgs[0]) {
			t.Fatalf("appPass doesn't match")
		}
		if fmt.Sprint(reg.OrderID) != test.params.Args[0] {
			t.Fatalf("order ID doesn't match")
		}
	}
}

func TestParseWithdrawArgs(t *testing.T) {
	paramsWithArgs := func(id, value string) *RawParams {
		pw := encode.PassBytes("password123")
		pwArgs := []encode.PassBytes{pw}
		args := []string{
			id,
			value,
			"abc",
		}
		return &RawParams{PWArgs: pwArgs, Args: args}
	}
	tests := []struct {
		name    string
		params  *RawParams
		wantErr error
	}{{
		name:   "ok",
		params: paramsWithArgs("42", "5000"),
	}, {
		name:    "assetID is not int",
		params:  paramsWithArgs("42.1", "5000"),
		wantErr: errArgs,
	}}
	for _, test := range tests {
		res, err := parseWithdrawArgs(test.params)
		if err != nil {
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("unexpected error %v for test %s",
					err, test.name)
			}
			continue
		}
		if !bytes.Equal(res.AppPass, test.params.PWArgs[0]) {
			t.Fatalf("appPass doesn't match")
		}
		if fmt.Sprint(res.AssetID) != test.params.Args[0] {
			t.Fatalf("assetID doesn't match")
		}
		if fmt.Sprint(res.Value) != test.params.Args[1] {
			t.Fatalf("value doesn't match")
		}
		if res.Address != test.params.Args[2] {
			t.Fatalf("address doesn't match")
		}
	}
}
