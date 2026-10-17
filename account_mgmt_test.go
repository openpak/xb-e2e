// SPDX-License-Identifier: AGPL-3.0-only

package e2e

// The dashboard's Xbox Live > ACCOUNT MANAGEMENT requests (native service 5)
// against the real xb-account handler after a real enrolment and signup: the
// account screens must report the real account, refuse billing, and the key a
// pass code change returns must be the key the KDC signs the user in with.

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/openpak/xb-common/krb"
	"github.com/openpak/xb-common/xbfeature"
)

type bufConn struct {
	in  [][]byte
	out bytes.Buffer
}

func (c *bufConn) Recv() ([]byte, error) {
	if len(c.in) == 0 {
		return nil, io.EOF
	}
	b := c.in[0]
	c.in = c.in[1:]
	return b, nil
}
func (c *bufConn) Send(b []byte) error { c.out.Write(b); return nil }

func svc5Call(t *testing.T, h xbfeature.Handler, info xbfeature.Info, path string, body []byte) (xerr string, out []byte) {
	t.Helper()
	req := "POST " + path + " HTTP/1.0\r\nUser-Agent: 5/1.00.5849\r\nContent-type: xon/5\r\nContent-Length: " +
		strconv.Itoa(len(body)) + "\r\n\r\n" + string(body)
	c := &bufConn{in: [][]byte{[]byte(req)}}
	if err := h.Serve(context.Background(), info, c); err != nil {
		t.Fatal(err)
	}
	head, b, _ := strings.Cut(c.out.String(), "\r\n\r\n")
	if !strings.HasPrefix(head, "HTTP/1.0 200") {
		return "http:" + strings.SplitN(head, "\r\n", 2)[0], nil
	}
	_, rest, _ := strings.Cut(head, "X-Err: ")
	xerr, _, _ = strings.Cut(rest, "\r\n")
	return xerr, []byte(b)
}

func TestAccountManagementAfterEnrolmentAndSignup(t *testing.T) {
	e := newEnrolWorld(t)
	_, r := e.register(serial)
	if st, body := e.web("approve", coreAcct, r.PairingCode); st != 200 {
		t.Fatalf("approve %d %s", st, body)
	}
	st, bb := e.post(e.prov, "/internal/v1/vouchers", map[string]string{"core_account_id": coreAcct, "machine_id": machineID})
	var v struct {
		Code string `json:"code"`
	}
	json.Unmarshal(bb, &v)
	if st != 200 {
		t.Fatalf("voucher %d %s", st, bb)
	}
	su, err := e.acct.Enrol.CompleteSignup(context.Background(), machineID, strings.ToLower(v.Code), gamertag)
	if err != nil {
		t.Fatal(err)
	}
	signed := e.fullSignIn(console{machineKey: consoleKey(), xuid: su.XUID})

	h := e.acct.Svc5(func(x uint64) ([16]byte, error) {
		_, secret := e.userKeys.Active()
		return krb.DeriveUserKey(secret, x), nil
	})
	xuid, _ := strconv.ParseUint(su.XUID, 16, 64)
	machine, _ := strconv.ParseUint(machineID, 16, 64)
	info := xbfeature.Info{PrincipalID: machineID, CoreAccountID: coreAcct, ServiceID: 5}
	le := binary.LittleEndian

	// Entering the screen: getsubscriptionstatus (observed 28-byte request), then troubleshootaccount.
	req := le.AppendUint64(nil, xuid)
	req = le.AppendUint64(req, machine)
	req = append(req, 0, 0, 0xfe, 0xff, 1, 0, 0, 0, 1, 0, 0, 0)
	if x, body := svc5Call(t, h, info, "/xuacs/getsubscriptionstatus.ashx", req); x != "0" || len(body) != 65 || body[8] != 0 {
		t.Fatalf("subscription status %q %x", x, body)
	}
	if x, body := svc5Call(t, h, info, "/xuacs/troubleshootaccount.ashx", req[:16]); x != "0" || len(body) != 32 {
		t.Fatalf("troubleshoot %q %x", x, body)
	}
	// Credit Card: the last-four gate never passes, billing data is never accepted.
	pin := append(le.AppendUint64(nil, xuid), '1', '2', '3', '4')
	if x, _ := svc5Call(t, h, info, "/xuacs/verifybillingpin.ashx", pin); x != "80154006" {
		t.Fatalf("billing pin %q", x)
	}
	if x, _ := svc5Call(t, h, info, "/xuacs/setpaymentinfo.ashx", make([]byte, 400)); x != "80154003" {
		t.Fatalf("setpaymentinfo %q", x)
	}
	// Account settings: stored for this user.
	if x, _ := svc5Call(t, h, info, "/xuacs/setusersettings.ashx", le.AppendUint16(le.AppendUint64(nil, xuid), 0x0005)); x != "0" {
		t.Fatalf("setusersettings %q", x)
	}
	if w, ok, err := e.acct.UserSettings(su.XUID); err != nil || !ok || w != 5 {
		t.Fatalf("stored settings %x %v %v", w, ok, err)
	}
	// Pass code: the key returned is the one the KDC signs this user in with.
	x, key := svc5Call(t, h, info, "/xuacs/setuserpin.ashx", le.AppendUint64(nil, xuid))
	if x != "0" || !bytes.Equal(key, signed.userKey[:]) {
		t.Fatalf("setuserpin %q key %x, KDC user key %x", x, key, signed.userKey[:])
	}
	// A different console (other account) cannot touch this user.
	other := info
	other.CoreAccountID = "core-account-2"
	if x, _ := svc5Call(t, h, other, "/xuacs/setuserpin.ashx", le.AppendUint64(nil, xuid)); x != "8015400a" {
		t.Fatalf("foreign setuserpin %q", x)
	}
}
