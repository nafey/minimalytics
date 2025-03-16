package model

import (
	"database/sql"
	"errors"
	"log"
	"time"
	// "minimalytics/model"
)

type Graph struct {
	Id          int64  `json:"id"`
	DashboardId int64  `json:"dashboardId"`
	Name        string `json:"name"`
	Event       string `json:"event"`
	Period      string `json:"period"`
	Length      int64  `json:"length"`
	CreatedOn   string `json:"createdOn"`
}

type GraphUpdate struct {
	Name   string `json:"name"`
	Event  string `json:"event"`
	Period string `json:"period"`
	Length int64  `json:"length"`
}

type GraphCreate struct {
	DashboardId int64  `json:"dashboardId"`
	Name        string `json:"name"`
	Event       string `json:"event"`
	Period      string `json:"period"`
	Length      int64  `json:"length"`
}

func InitGraphs() {
	query := `
		CREATE TABLE IF NOT EXISTS graphs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dashboardId INTEGER NOT NULL,
			name TEXT,
			event TEXT,
			period TEXT,
			length INTEGER,
			createdOn TEXT
		);`
	_, err := db.Exec(query)
	if err != nil {
		log.Println("failed to create table: %w", err)
		return
	}
	return

}

func IsValidGraphId(graphId int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (
		SELECT 1
		FROM graphs
		WHERE id = ?
	);`

	err := db.QueryRow(query, graphId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GetDashboardGraphs(dashboardId int64) ([]Graph, error) {
	var graphs []Graph

	exists, _ := IsValidDashboardId(dashboardId)
	if !exists {
		return graphs, errors.New("Invalid DashboardId")
	}

	rows, err := db.Query("select * from graphs where dashboardId = ?", dashboardId)
	if err != nil {
		return graphs, err
	}
	defer rows.Close()

	for rows.Next() {
		var graph Graph
		err := rows.Scan(&graph.Id, &graph.DashboardId, &graph.Name, &graph.Event, &graph.Period, &graph.Length, &graph.CreatedOn)
		if err != nil {
			return graphs, err
		}
		graphs = append(graphs, graph)
	}

	return graphs, err
}

func GetGraph(graphId int64) (Graph, error) {
	row := db.QueryRow("select * from graphs where id = ?", graphId)

	var graph Graph
	err := row.Scan(&graph.Id, &graph.DashboardId, &graph.Name, &graph.Event, &graph.Period, &graph.Length, &graph.CreatedOn)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return graph, errors.New("Invalid graphId")
		}
	}

	return graph, err
}

func UpdateGraph(graphId int64, updateGraph GraphUpdate) error {
	_, err := GetGraph(graphId)
	if err != nil {
		return err
	}

	name := updateGraph.Name
	event := updateGraph.Event
	period := updateGraph.Period
	length := updateGraph.Length

	if name != "" {
		// Add validation if needed
	}

	if event != "" {
		exists, _ := IsValidEvent(event)
		if !exists {
			return errors.New("Invalid event value")
		}

	}

	if period != "" {
		if period != "DAILY" && period != "HOURLY" && period != "MINUTELY" {
			return errors.New("Invalid period value")
		}
	}

	if length == 0 {
		return errors.New("Invalid length value")
	}

	_, err = db.Exec(`
		UPDATE graphs
			set name = coalesce(NULLIF(?, ''), name),
				event = coalesce(NULLIF(?, ''), event),
				period = coalesce(NULLIF(?, ''), period),
				length = coalesce(NULLIF(?, 0), length)
			where id = ?`,
		name, event, period, length, graphId)

	return err
}

func DeleteGraph(graphId int64) error {
	exists, err := IsValidGraphId(graphId)
	if !exists {
		return errors.New("Invalid graphId")
	}

	if err != nil {
		return err
	}

	_, err = db.Exec(
		`
		DELETE FROM graphs where id = ?
		`,
		graphId)

	return err
}

func CreateGraph(createGraph GraphCreate) (Graph, error) {
	var graph Graph
	dashboardId := createGraph.DashboardId
	name := createGraph.Name
	event := createGraph.Event
	period := createGraph.Period
	length := createGraph.Length

	if dashboardId <= 0 {
		return graph, errors.New("Invalid dashboardId")
	} else {
		exists, _ := IsValidDashboard(dashboardId)
		if !exists {
			return graph, errors.New("Invalid dashboardId")
		}

	}

	if name == "" {
		return graph, errors.New("Invalid name")
	}

	if event != "" {
		exists, _ := IsValidEvent(event)
		if !exists {
			return graph, errors.New("Invalid event value")
		}

	} else {
		return graph, errors.New("Event value cannot be empty")

	}

	if period != "" {
		if period != "DAILY" && period != "HOURLY" && period != "MINUTELY" {
			return graph, errors.New("Invalid period value")
		}

	} else {
		return graph, errors.New("Period cannot be empty")

	}

	if length <= 0 {
		return graph, errors.New("Invalid length value")
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	result, err := db.Exec(
		`
		INSERT INTO graphs (dashboardId, name, event, period, length, createdOn)
		values (?, ?, ?, ?, ?, ?)
		`,
		dashboardId, name, event, period, length, formattedTime)

	if err != nil {
		return graph, err
	}

	graphId, err := result.LastInsertId()
	if err != nil {
		return graph, err
	}

	graph, err = GetGraph(graphId)
	return graph, err

}

func GetGraphData(graphId int64) ([]TimeStat, error) {
	var statsArray []TimeStat
	graph, err := GetGraph(graphId)

	if err != nil {
		return statsArray, err
	}

	event := graph.Event
	period := graph.Period
	length := graph.Length

	return GetEventData(event, period, length)

	// currentTime := time.Now()
	// var startTime time.Time
	// var fromTime time.Time
	// var toTimestamp int64
	// var periodPrefix string

	// var intLength int = int(length)

	// if period == "DAILY" {
	// 	periodPrefix = "daily"
	// 	startTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	// 	fromTime = startTime.AddDate(0, 0, -1*int(length))

	// } else if period == "HOURLY" {
	// 	periodPrefix = "hourly"
	// 	startTime = currentTime.Truncate(time.Hour)
	// 	fromTime = startTime.Add(-time.Duration(intLength) * time.Hour)

	// } else {
	// 	periodPrefix = "minutely"
	// 	startTime = currentTime.Truncate(time.Minute)
	// 	fromTime = startTime.Add(-time.Duration(intLength) * time.Minute)

	// }

	// toTimestamp = startTime.Unix()
	// fromTimestamp := time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(), 0, 0, 0, 0, fromTime.Location()).Unix()

	// query := fmt.Sprintf("select * from %s_%s where time between ? and ?", periodPrefix, event)
	// rows, err := db.Query(query, fromTimestamp, toTimestamp)
	// if err != nil {
	// 	panic(err)
	// }

	// countMap := make(map[int64]int64)
	// for rows.Next() {
	// 	var eventRow EventRow
	// 	err := rows.Scan(&eventRow.Time, &eventRow.Count)

	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	countMap[eventRow.Time] = eventRow.Count
	// }

	// statsArray := make([]TimeStat, intLength)
	// for i := 0; i < intLength; i++ {
	// 	iTime := startTime.Add(time.Duration(-i) * time.Minute)
	// 	iTimestamp := iTime.Unix()
	// 	iCount := int64(0)

	// 	foundCount, ok := countMap[iTimestamp]
	// 	if ok {
	// 		iCount = foundCount
	// 	}

	// 	iStatItem := TimeStat{
	// 		Time:  iTimestamp,
	// 		Count: iCount,
	// 	}

	// 	statsArray[i] = iStatItem
	// }

	// return statsArray
}
