package cacheodon

import (
	"os"
	"testing"
	"time"
)

func TestInit(t *testing.T) {

	tempdir := t.TempDir()
	// Create an empty DB
	if db, err := NewFinderDB(tempdir + "/test.sqlite3"); err != nil {
		t.Fatal(err)
	} else {
		db.Close()
	}
}

// Returns a GeocacheLog and Geocache filled with the provided test data
func getTestData(finderName string, findTime time.Time, cacheCode string, logText string) (*GeocacheLog, *Geocache) {
	gc := Geocache{
		Code:          cacheCode,
		LastFoundTime: findTime,
	}
	l := GeocacheLog{
		UserName: finderName,
		LogText:  logText,
	}
	return &l, &gc
}

func TestAddLog(t *testing.T) {
	tempdir := t.TempDir()
	// Set the "current time" to midday so we don't run into issues with the midnight rollover.
	timeNow := time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC)
	timeMidnight := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	// Create an empty DB
	if db, err := NewFinderDB(tempdir + "/test.sqlite3"); err != nil {
		t.Fatal(err)
	} else {
		defer db.Close()
		// Add a find from right now
		db.AddLog(getTestData("testname", timeNow, "GC123", "testlog"))
		// Check one find in the last 24 hours.
		if want, got := 1, db.FindsSinceTime("testname", timeMidnight); want != got {
			t.Fatalf("FindsSinceMidnight returned wrong value: want %d, got %d", want, got)
		}
		// Add an irrelevant find.
		db.AddLog(getTestData("testname2", timeNow, "GC321", "testlog"))

		// Add another find ten minutes ago
		db.AddLog(getTestData("testname", timeNow.Add(-10*time.Minute), "GC456", "testlog2"))
		// Check we now have two finds since midnight
		if want, got := 2, db.FindsSinceTime("testname", timeMidnight); want != got {
			t.Fatalf("FindsSinceMidnight returned wrong value: want %d, got %d", want, got)
		}
		// Add a find 24 hours ago and ensure it doesn't count towards today's finds
		db.AddLog(getTestData("testname", timeNow.Add(-24*time.Hour), "GC789", "testlog3"))
		if want, got := 2, db.FindsSinceTime("testname", timeMidnight); want != got {
			t.Fatalf("FindsSinceMidnight returned wrong value: want %d, got %d", want, got)
		}
	}
}

func TestPersistence(t *testing.T) {
	tempdir := t.TempDir()
	// Set the "current time" to midday so we don't run into issues with the midnight rollover.
	timeNow := time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC)
	timeMidnight := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	gc := Geocache{
		Code:       "GC123",
		PlacedDate: "2023-03-02T16:44:59",
	}

	// Create an empty DB
	if db, err := NewFinderDB(tempdir + "/test.sqlite3"); err != nil {
		t.Fatal(err)
	} else {
		defer db.Close()
		db.AddLog(getTestData("testname", timeNow, "GC123", "testlog"))
		db.UpdateCache(&gc)
	}
	if db, err := NewFinderDB(tempdir + "/test.sqlite3"); err != nil {
		t.Fatal(err)
	} else {
		defer db.Close()
		if want, got := 1, db.FindsSinceTime("testname", timeMidnight); want != got {
			t.Fatalf("FindsSinceMidnight returned wrong value: want %d, got %d", want, got)
		}
		if new, _ := db.UpdateCache(&gc); new {
			t.Fatal("Cache should not be new")
		}
	}
}

func TestUpdateCache(t *testing.T) {
	tempdir := t.TempDir()
	// Create an empty DB
	if db, err := NewFinderDB(tempdir + "/test.sqlite3"); err != nil {
		t.Fatal(err)
	} else {
		defer db.Close()
		// Add a cache
		gc := Geocache{
			Code:       "GC123",
			PlacedDate: "2023-03-02T16:44:59",
		}
		if new, _ := db.UpdateCache(&gc); !new {
			t.Fatal("Cache should be new")
		}
		// Check the cache is there
		{
			new, updated := db.UpdateCache(&gc)
			if new {
				t.Fatal("Cache should not be new")
			}
			if updated {
				t.Fatal("Cache should not be updated")
			}
		}

		gc.LastFoundTime = gc.LastFoundTime.Add(time.Hour)

		{
			new, updated := db.UpdateCache(&gc)
			if new {
				t.Fatal("Cache should not be new")
			}
			if !updated {
				t.Fatal("Cache should be updated")
			}
		}
	}
}

func TestLastPostedFoundTime(t *testing.T) {
	tempdir := t.TempDir()
	// Set the "current time" to midday so we don't run into issues with the midnight rollover.
	timeNow := time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC)
	timeLater := time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create an empty DB
	if db, err := NewFinderDB(tempdir + "/test.sqlite3"); err != nil {
		t.Fatal(err)
	} else {
		defer db.Close()
		// These should be overridden
		db.SetLastPostedFoundTime(timeLater.Add(-1 * time.Hour))
		db.GetLastPostedFoundTime(timeLater)
		db.SetLastPostedFoundTime(timeLater.Add(1 * time.Hour))
		db.GetLastPostedFoundTime(timeLater)
		db.SetLastPostedFoundTime(timeNow)
		if want, got := timeNow, db.GetLastPostedFoundTime(timeLater); want != got {
			t.Fatalf("LastPostedFoundTime returned wrong value: want %s, got %s", want, got)
		}
	}
	// Check the value is persisted
	if db, err := NewFinderDB(tempdir + "/test.sqlite3"); err != nil {
		t.Fatal(err)
	} else {
		defer db.Close()
		// Read out the current state
		if want, got := timeNow, db.GetLastPostedFoundTime(timeLater); want != got {
			t.Fatalf("LastPostedFoundTime didn't remember the right value: want %s, got %s", want, got)
		}
		// Check that the State table only has one row
		var got int64
		db.db.Model(&State{}).Count(&got)
		if got != 1 {
			t.Fatalf("State table has wrong number of rows: want 1, got %d", got)
		}
	}
	// Copy the temp db to /tmp/ so we can inspect it
	if err := os.Rename(tempdir+"/test.sqlite3", "/tmp/test.sqlite3"); err != nil {
		t.Fatal(err)
	}
}

func TestLastPostedFoundTimeDefault(t *testing.T) {
	tempdir := t.TempDir()
	// Set the "current time" to midday so we don't run into issues with the midnight rollover.
	timeNow := time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC)
	timeLater := time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create an empty DB
	if db, err := NewFinderDB(tempdir + "/test.sqlite3"); err != nil {
		t.Fatal(err)
	} else {
		defer db.Close()
		if want, got := timeLater, db.GetLastPostedFoundTime(timeLater); want != got {
			t.Fatalf("LastPostedFoundTime returned wrong value: want %s, got %s", want, got)
		}
	}
	// Check the value is persisted, and that the default timeNow is ignored in favour
	// of returning the stored value.
	if db, err := NewFinderDB(tempdir + "/test.sqlite3"); err != nil {
		t.Fatal(err)
	} else {
		defer db.Close()
		if want, got := timeLater, db.GetLastPostedFoundTime(timeNow); want != got {
			t.Fatalf("LastPostedFoundTime didn't remember the right value: want %s, got %s", want, got)
		}
	}
}
