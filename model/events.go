package model

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

var didInit bool = false
var db *sql.DB

type EventRow struct {
	Time  int64
	Count int64
}

type TimeStat struct {
	Time  int64 `json:"time"`
	Count int64 `json:"count"`
}

type EventDef struct {
	Id       string
	Event    string
	LastSeen string
}

func InitEvents() {
	query := `
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event TEXT,
			lastSeen TEXT
		);`

	_, err := db.Exec(query)
	if err != nil {
		log.Println("failed to create table: %w", err)
	}
}

func InitDailyEvent(event string) {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS daily_%s (
			time INTEGER PRIMARY KEY,
			count INTEGER
		);`, event)

	_, err := db.Exec(query)
	if err != nil {
		log.Println("failed to create table: %w", err)
	}
}

func InitHourlyEvent(event string) {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS hourly_%s (
			time INTEGER PRIMARY KEY,
			count INTEGER
		);`, event)

	_, err := db.Exec(query)
	if err != nil {
		log.Println("failed to create table: %w", err)
	}
}

func InitMinutelyEvent(event string) {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS minutely_%s (
			time INTEGER PRIMARY KEY,
			count INTEGER
		);`, event)

	_, err := db.Exec(query)
	if err != nil {
		log.Println("failed to create table: %w", err)
	}
}

func InitEvent(event string) {
	row := db.QueryRow("select * from event where event = ?", event)

	var eventDef EventDef
	err := row.Scan(&eventDef.Id, &eventDef.Event, &eventDef.LastSeen)
	if err != nil {
		// New Event
		db.Exec("insert into events (event) values (?)", event)
		InitDailyEvent(event)
		InitHourlyEvent(event)
		InitMinutelyEvent(event)
	} else {
		// Skip

	}
}

func SubmitDailyEvent(event string) {
	currentTime := time.Now()

	dayStart := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	time := dayStart.Unix()

	query := fmt.Sprintf("select * from daily_%s where time = ?", event)

	row := db.QueryRow(query, time)
	var eventRow EventRow
	err := row.Scan(&eventRow.Time, &eventRow.Count)
	if err != nil {
		query = fmt.Sprintf("insert into daily_%s (time, count) values (?, ?)", event)
		db.Exec(query, time, 1)

	} else {
		nextCount := eventRow.Count + 1
		query = fmt.Sprintf("update daily_%s set count = ? where time = ?", event)
		db.Exec(query, nextCount, time)
	}
}

func SubmitHourlyEvent(event string) {
	currentTime := time.Now()

	hourStart := currentTime.Truncate(time.Hour)
	time := hourStart.Unix()

	query := fmt.Sprintf("select * from hourly_%s where time = ?", event)

	row := db.QueryRow(query, time)
	var eventRow EventRow
	err := row.Scan(&eventRow.Time, &eventRow.Count)
	if err != nil {
		query = fmt.Sprintf("insert into hourly_%s (time, count) values (?, ?)", event)
		db.Exec(query, time, 1)

	} else {
		nextCount := eventRow.Count + 1
		query = fmt.Sprintf("update hourly_%s set count = ? where time = ?", event)
		db.Exec(query, nextCount, time)

	}
}

func SubmitMinuteEvent(event string) {
	currentTime := time.Now()

	minuteStart := currentTime.Truncate(time.Minute)
	time := minuteStart.Unix()

	query := fmt.Sprintf("select * from minutely_%s where time = ?", event)

	row := db.QueryRow(query, time)
	var eventRow EventRow
	err := row.Scan(&eventRow.Time, &eventRow.Count)
	if err != nil {
		query = fmt.Sprintf("insert into minutely_%s (time, count) values (?, ?)", event)
		db.Exec(query, time, 1)

	} else {
		nextCount := eventRow.Count + 1
		query = fmt.Sprintf("update minutely_%s set count = ? where time = ?", event)
		db.Exec(query, nextCount, time)

	}
}

func GetDailyStat(event string) *[60]TimeStat {
	currentTime := time.Now()

	dayStart := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	toTimestamp := dayStart.Unix()

	fromTime := dayStart.AddDate(0, 0, -60)
	fromTimestamp := time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(), 0, 0, 0, 0, fromTime.Location()).Unix()

	query := fmt.Sprintf("select * from daily_%s where time between ? and ?", event)

	rows, err := db.Query(query, fromTimestamp, toTimestamp)
	if err != nil {
		panic(err)
	}

	countMap := make(map[int64]int64)
	for rows.Next() {
		var eventRow EventRow
		err = rows.Scan(&eventRow.Time, &eventRow.Count)
		if err != nil {
			panic(err)
		}

		countMap[eventRow.Time] = eventRow.Count
	}

	var statsArray [60]TimeStat
	for i := 0; i < 60; i++ {
		iTime := dayStart.AddDate(0, 0, -1*i)
		iTimestamp := time.Date(iTime.Year(), iTime.Month(), iTime.Day(), 0, 0, 0, 0, iTime.Location()).Unix()

		iCount := int64(0)

		foundCount, ok := countMap[iTimestamp]
		if ok {
			iCount = foundCount
		}

		iStatItem := TimeStat{
			Time:  iTimestamp,
			Count: iCount,
		}

		statsArray[i] = iStatItem
	}

	return &statsArray
}

func GetHourlyStat(event string) *[48]TimeStat {
	currentTime := time.Now()

	hourStart := currentTime.Truncate(time.Hour)
	toTimestamp := hourStart.Unix()

	fromTime := hourStart.Add(-48 * time.Hour)
	fromTimestamp := fromTime.Unix()

	query := fmt.Sprintf("select * from hourly_%s where time between ? and ?", event)

	rows, err := db.Query(query, fromTimestamp, toTimestamp)
	if err != nil {
		panic(err)
	}

	countMap := make(map[int64]int64)
	for rows.Next() {
		var eventRow EventRow
		err = rows.Scan(&eventRow.Time, &eventRow.Count)
		if err != nil {
			panic(err)
		}

		countMap[eventRow.Time] = eventRow.Count
	}

	var statsArray [48]TimeStat
	for i := 0; i < 48; i++ {
		iTime := hourStart.Add(time.Duration(-i) * time.Hour)
		iTimestamp := iTime.Unix()
		iCount := int64(0)

		foundCount, ok := countMap[iTimestamp]
		if ok {
			iCount = foundCount
		}

		iStatItem := TimeStat{
			Time:  iTimestamp,
			Count: iCount,
		}

		statsArray[i] = iStatItem
	}

	return &statsArray
}

func GetMinuteStat(event string) *[60]TimeStat {
	currentTime := time.Now()

	minuteStart := currentTime.Truncate(time.Minute)
	toTimestamp := minuteStart.Unix()

	fromTime := minuteStart.Add(-60 * time.Minute)
	fromTimestamp := fromTime.Unix()

	query := fmt.Sprintf("select * from minutely_%s where time between ? and ?", event)

	rows, err := db.Query(query, fromTimestamp, toTimestamp)
	if err != nil {
		panic(err)
	}

	countMap := make(map[int64]int64)
	for rows.Next() {
		var eventRow EventRow
		err = rows.Scan(&eventRow.Time, &eventRow.Count)
		if err != nil {
			panic(err)
		}

		countMap[eventRow.Time] = eventRow.Count
	}

	var statsArray [60]TimeStat
	for i := 0; i < 60; i++ {
		iTime := minuteStart.Add(time.Duration(-i) * time.Minute)
		iTimestamp := iTime.Unix()
		iCount := int64(0)

		foundCount, ok := countMap[iTimestamp]
		if ok {
			iCount = foundCount
		}

		iStatItem := TimeStat{
			Time:  iTimestamp,
			Count: iCount,
		}

		statsArray[i] = iStatItem
	}

	return &statsArray

}
