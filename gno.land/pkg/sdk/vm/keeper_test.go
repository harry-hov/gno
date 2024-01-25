package vm

// TODO: move most of the logic in ROOT/gno.land/...

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jaekwon/testify/assert"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestVMKeeperAddPackage(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

	pkgPath := "gno.land/r/test"

	// Create test package.
	files := []*std.MemFile{
		{
			Name: "test.gno",
			Body: `package test

import "std"

func Echo() string {
	return "hello world"
}`,
		},
	}
	modfile := &std.MemMod{
		ImportPath: pkgPath,
		Version:    "v0.0.0",
	}

	msg1 := NewMsgAddPackage(addr, modfile, files)
	assert.Nil(t, env.vmk.gnoStore.GetPackage(pkgPath, modfile.Version, false))

	err := env.vmk.AddPackage(ctx, msg1)

	assert.NoError(t, err)
	assert.NotNil(t, env.vmk.gnoStore.GetPackage(pkgPath, modfile.Version, false))

	err = env.vmk.AddPackage(ctx, msg1)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, InvalidPkgPathError{}))
}

// Sending total send amount succeeds.
func TestVMKeeperOrigSend1(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

	pkgPath := "gno.land/r/test"

	// Create test package.
	files := []*std.MemFile{
		{"init.gno", `
package test

import "std"

func init() {
}

func Echo(msg string) string {
	addr := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()
	send := std.GetOrigSend()
	banker := std.GetBanker(std.BankerTypeOrigSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}
	modfile := &std.MemMod{
		ImportPath: pkgPath,
		Version:    "v0.0.0",
	}

	msg1 := NewMsgAddPackage(addr, modfile, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins("10000000ugnot")
	msg2 := NewMsgCall(addr, coins, pkgPath, modfile.Version, "Echo", []string{"hello world"})
	res, err := env.vmk.Call(ctx, msg2)
	assert.NoError(t, err)
	assert.Equal(t, res, `("echo:hello world" string)`)
	// t.Log("result:", res)
}

// Sending too much fails
func TestVMKeeperOrigSend2(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

	pkgPath := "gno.land/r/test"

	// Create test package.
	files := []*std.MemFile{
		{"init.gno", `
package test

import "std"

var admin std.Address

func init() {
     admin =	std.GetOrigCaller()
}

func Echo(msg string) string {
	addr := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()
	send := std.GetOrigSend()
	banker := std.GetBanker(std.BankerTypeOrigSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}

func GetAdmin() string {
	return admin.String()
}
`},
	}
	modfile := &std.MemMod{
		ImportPath: pkgPath,
		Version:    "v0.0.0",
	}

	msg1 := NewMsgAddPackage(addr, modfile, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins("11000000ugnot")
	msg2 := NewMsgCall(addr, coins, pkgPath, modfile.Version, "Echo", []string{"hello world"})
	res, err := env.vmk.Call(ctx, msg2)
	assert.Error(t, err)
	assert.Equal(t, res, "")
	fmt.Println(err.Error())
	assert.True(t, strings.Contains(err.Error(), "insufficient coins error"))
}

// Sending more than tx send fails.
func TestVMKeeperOrigSend3(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

	pkgPath := "gno.land/r/test"

	// Create test package.
	files := []*std.MemFile{
		{"init.gno", `
package test

import "std"

func init() {
}

func Echo(msg string) string {
	addr := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()
	send := std.Coins{{"ugnot", 10000000}}
	banker := std.GetBanker(std.BankerTypeOrigSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}
	modfile := &std.MemMod{
		ImportPath: pkgPath,
		Version:    "v0.0.0",
	}

	msg1 := NewMsgAddPackage(addr, modfile, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins("9000000ugnot")
	msg2 := NewMsgCall(addr, coins, pkgPath, modfile.Version, "Echo", []string{"hello world"})
	// XXX change this into an error and make sure error message is descriptive.
	_, err = env.vmk.Call(ctx, msg2)
	assert.Error(t, err)
}

// Sending realm package coins succeeds.
func TestVMKeeperRealmSend1(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

	pkgPath := "gno.land/r/test"

	// Create test package.
	files := []*std.MemFile{
		{"init.gno", `
package test

import "std"

func init() {
}

func Echo(msg string) string {
	addr := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()
	send := std.Coins{{"ugnot", 10000000}}
	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}
	modfile := &std.MemMod{
		ImportPath: pkgPath,
		Version:    "v0.0.0",
	}

	msg1 := NewMsgAddPackage(addr, modfile, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins("10000000ugnot")
	msg2 := NewMsgCall(addr, coins, pkgPath, modfile.Version, "Echo", []string{"hello world"})
	res, err := env.vmk.Call(ctx, msg2)
	assert.NoError(t, err)
	assert.Equal(t, res, `("echo:hello world" string)`)
}

// Sending too much realm package coins fails.
func TestVMKeeperRealmSend2(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

	pkgPath := "gno.land/r/test"

	// Create test package.
	files := []*std.MemFile{
		{"init.gno", `
package test

import "std"

func init() {
}

func Echo(msg string) string {
	addr := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()
	send := std.Coins{{"ugnot", 10000000}}
	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}
	modfile := &std.MemMod{
		ImportPath: pkgPath,
		Version:    "v0.0.0",
	}

	msg1 := NewMsgAddPackage(addr, modfile, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins("9000000ugnot")
	msg2 := NewMsgCall(addr, coins, pkgPath, modfile.Version, "Echo", []string{"hello world"})
	// XXX change this into an error and make sure error message is descriptive.
	_, err = env.vmk.Call(ctx, msg2)
	assert.Error(t, err)
}

// Assign admin as OrigCaller on deploying the package.
func TestVMKeeperOrigCallerInit(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

	pkgPath := "gno.land/r/test"

	// Create test package.
	files := []*std.MemFile{
		{"init.gno", `
package test

import "std"

var admin std.Address

func init() {
     admin =	std.GetOrigCaller()
}

func Echo(msg string) string {
	addr := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()
	send := std.GetOrigSend()
	banker := std.GetBanker(std.BankerTypeOrigSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}

func GetAdmin() string {
	return admin.String()
}

`},
	}
	modfile := &std.MemMod{
		ImportPath: pkgPath,
		Version:    "v0.0.0",
	}

	msg1 := NewMsgAddPackage(addr, modfile, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run GetAdmin()
	coins := std.MustParseCoins("")
	msg2 := NewMsgCall(addr, coins, pkgPath, modfile.Version, "GetAdmin", []string{})
	res, err := env.vmk.Call(ctx, msg2)
	addrString := fmt.Sprintf("(\"%s\" string)", addr.String())
	assert.NoError(t, err)
	assert.Equal(t, res, addrString)
}

// Call Run without imports, without variables.
func TestVMKeeperRunSimple(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)

	pkgPath := "gno.land/r/test"
	files := []*std.MemFile{
		{"script.gno", `
package main

func main() {
	println("hello world!")
}
`},
	}
	modfile := &std.MemMod{
		ImportPath: pkgPath,
		Version:    "v0.0.0",
	}

	coins := std.MustParseCoins("")
	msg2 := NewMsgRun(addr, coins, modfile, files)
	res, err := env.vmk.Run(ctx, msg2)
	assert.NoError(t, err)
	assert.Equal(t, res, "hello world!\n")
}

// Call Run with stdlibs.
func TestVMKeeperRunImportStdlibs(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)

	pkgPath := "gno.land/r/test"
	files := []*std.MemFile{
		{"script.gno", `
package main

import "std"

func main() {
	addr := std.GetOrigCaller()
	println("hello world!", addr)
}
`},
	}
	modfile := &std.MemMod{
		ImportPath: pkgPath,
		Version:    "v0.0.0",
	}

	coins := std.MustParseCoins("")
	msg2 := NewMsgRun(addr, coins, modfile, files)
	res, err := env.vmk.Run(ctx, msg2)
	assert.NoError(t, err)
	expectedString := fmt.Sprintf("hello world! %s\n", addr.String())
	assert.Equal(t, res, expectedString)
}

func TestNumberOfArgsError(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

	// Create test package.
	files := []*std.MemFile{
		{
			Name: "test.gno",
			Body: `package test

import "std"

func Echo(msg string) string {
	return "echo:"+msg
}`,
		},
	}
	pkgPath := "gno.land/r/test"
	modfile := &std.MemMod{
		ImportPath: pkgPath,
		Version:    "v0.0.0",
	}
	msg1 := NewMsgAddPackage(addr, modfile, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Call Echo function with wrong number of arguments
	coins := std.MustParseCoins("1ugnot")
	msg2 := NewMsgCall(addr, coins, pkgPath, modfile.Version, "Echo", []string{"hello world", "extra arg"})
	assert.PanicsWithValue(
		t,
		func() {
			env.vmk.Call(ctx, msg2)
		},
		"wrong number of arguments in call to Echo: want 1 got 2",
	)
}
