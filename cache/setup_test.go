package cache

import (
	"github.com/dgraph-io/badger/v4"
	"log"
	"os"
	"testing"
)

var testBadgerCache BadgerCache

func TestMain(m *testing.M) {

	// clear and create badger database
	_ = os.RemoveAll("./testdata/tmp/badger")
	if _, err := os.Stat("./testdata/tmp"); os.IsNotExist(err) {
		err := os.Mkdir("./testdata/tmp", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	err := os.Mkdir("./testdata/tmp/badger", 0755)
	if err != nil {
		log.Fatal(err)
	}

	db, _ := badger.Open(badger.DefaultOptions("./testdata/tmp/badger"))
	testBadgerCache.Conn = db

	os.Exit(m.Run())
}
