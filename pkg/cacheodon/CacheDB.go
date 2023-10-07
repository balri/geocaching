package cacheodon

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// This stores the detail of a single find event.
type CacheFind struct {
	gorm.Model
	Name      string
	FindTime  time.Time
	FindType  string
	CacheCode string
	LogString string
}

type Cache struct {
	gorm.Model
	Code          string
	LastFoundTime time.Time
	Updated       bool `gorm:"-"` // This is not stored in the database, it is used to track if the cache was updated.
	New           bool `gorm:"-"` // This is not stored in the database, it is used to track if the cache was updated.
}

type State struct {
	gorm.Model
	LastPostedFoundTime time.Time
}

// This stores the finder database.
type FinderDB struct {
	db *gorm.DB
}

// Open the DB and migrate if required.
func (f *FinderDB) Init(filename string) error {
	var err error
	f.db, err = gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return err
	}

	// Migrate the schema
	f.db.AutoMigrate(&CacheFind{})
	f.db.AutoMigrate(&Cache{})
	f.db.AutoMigrate(&State{})

	return nil
}

// Close the DB.
func (f *FinderDB) Close() error {
	sqlDB, err := f.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// This returns the last posted found time. If no time was saved, the default time is returned.
func (f *FinderDB) GetLastPostedFoundTime(def time.Time) time.Time {
	var state State
	f.db.First(&state)
	// If state.LastPostedFoundTime is zero, then we haven't saved a time yet.
	// Save and return the default time.
	if state.LastPostedFoundTime.IsZero() {
		f.db.Create(&State{LastPostedFoundTime: def})
		return def
	}
	return state.LastPostedFoundTime
}

// This sets the last posted found time.
func (f *FinderDB) SetLastPostedFoundTime(t time.Time) {
	var state State
	if tx := f.db.First(&state); tx.RowsAffected == 0 {
		// No state record exists, create one.
		f.db.Create(&State{LastPostedFoundTime: t})
		return
	}
	state.LastPostedFoundTime = t
	f.db.Save(&state)
}

// TODO If a new cache shows up in the database publish a message about it.

// This adds a cache to the database. If the cache already exists, it is not added.

func (f *FinderDB) UpdateCache(gc *Geocache) (new bool, updated bool) {
	var count int64
	f.db.Model(&Cache{}).Where("code = ?", gc.Code).Count(&count)
	if count == 0 {
		// This is a new cache.
		f.db.Create(&Cache{
			Code:          gc.Code,
			LastFoundTime: gc.LastFoundTime,
			New:           true,
		})
		new = true
	} else {
		// Update the record
		var cache Cache
		f.db.First(&cache, "code = ?", gc.Code)
		if !cache.LastFoundTime.Equal(gc.LastFoundTime) {
			cache.Updated = true
			updated = true
			cache.LastFoundTime = gc.LastFoundTime
			f.db.Save(&cache)
		}
	}
	return new, updated
}

// This adds a find to the database
func (f *FinderDB) AddLog(cf *GeocacheLog, gc *Geocache) {
	f.db.Create(&CacheFind{
		Name:      cf.UserName,
		FindTime:  gc.LastFoundTime,
		CacheCode: gc.Code,
		LogString: cf.LogText,
		FindType:  cf.LogType,
	})
}

// This returns the number of finds since local midnight for a given name.
func (f *FinderDB) FindsSinceMidnight(name string) int {
	return f.FindsSinceTime(name, time.Now().Truncate(24*time.Hour))
}

// This returns the number of finds since local midnight for a given name.
func (f *FinderDB) FindsSinceTime(name string, t time.Time) int {
	var count int64
	f.db.Model(&CacheFind{}).Where("name = ? AND find_time >= ?", name, t).Count(&count)
	return int(count)
}

func NewFinderDB(filename string) (*FinderDB, error) {
	fdb := &FinderDB{}
	if err := fdb.Init(filename); err != nil {
		return fdb, err
	}

	return fdb, nil
}
