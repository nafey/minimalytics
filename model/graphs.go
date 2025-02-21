package model

import (
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
	CreatedOn   string `json:"createdOn"`
}

type GraphUpdate struct {
	Name   string `json:"name"`
	Event  string `json:"event"`
	Period string `json:"period"`
}

type GraphCreate struct {
	DashboardId int64  `json:"dashboardId"`
	Name        string `json:"name"`
	Event       string `json:"event"`
	Period      string `json:"period"`
}

func InitGraphs() {
	query := `
		CREATE TABLE IF NOT EXISTS graphs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dashboardId INTEGER NOT NULL,
			name TEXT,
			event TEXT,
			period TEXT,
			createdOn TEXT
		);`
	_, err := db.Exec(query)
	if err != nil {
		log.Println("failed to create table: %w", err)
		return
	}
	return

}

func GetDashboardGraphs(dashboardId int64) []Graph {
	rows, err := db.Query("select * from graphs where dashboardId = ?", dashboardId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var graphs []Graph
	for rows.Next() {
		var graph Graph
		err := rows.Scan(&graph.Id, &graph.DashboardId, &graph.Name, &graph.Event, &graph.Period, &graph.CreatedOn) // Replace with actual fields
		if err != nil {
			// Handle scan error
			panic(err)
		}
		graphs = append(graphs, graph)
	}

	return graphs
}

func GetGraph(graphId int64) Graph {
	row := db.QueryRow("select * from graphs where id = ?", graphId)

	var graph Graph
	err := row.Scan(&graph.Id, &graph.DashboardId, &graph.Name, &graph.Event, &graph.Period, &graph.CreatedOn)
	if err != nil {
		panic("Dashboard not found")
	}

	return graph
}

func UpdateGraph(graphId int64, updateGraph GraphUpdate) error {
	name := updateGraph.Name
	event := updateGraph.Event
	period := updateGraph.Period

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

	_, err := db.Exec(`
		UPDATE graphs
			set name = coalesce(NULLIF(?, ''), name),
				event = coalesce(NULLIF(?, ''), event),
				period = coalesce(NULLIF(?, ''), period)
			where id = ?`,
		name, event, period, graphId)

	if err != nil {
		panic(err)
	}

	return nil
}

func DeleteGraph(graphId int64) error {
	_, err := db.Exec(
		`
		DELETE FROM graphs where id = ?
		`,
		graphId)

	if err != nil {
		panic(err)
	}

	return nil
}

func CreateGraph(createGraph GraphCreate) error {
	dashboardId := createGraph.DashboardId
	name := createGraph.Name
	event := createGraph.Event
	period := createGraph.Period

	if dashboardId <= 0 {
		return errors.New("Invalid dashboardId")
	} else {
		exists, _ := IsValidDashboard(dashboardId)
		if !exists {
			return errors.New("Invalid dashboardId")
		}

	}

	if name == "" {
		return errors.New("Invalid name")
	}

	if event != "" {
		exists, _ := IsValidEvent(event)
		if !exists {
			return errors.New("Invalid event value")
		}

	} else {
		return errors.New("Event value cannot be empty")

	}

	if period != "" {
		if period != "DAILY" && period != "HOURLY" && period != "MINUTELY" {
			return errors.New("Invalid period value")
		}

	} else {
		return errors.New("Period cannot be empty")

	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	_, err := db.Exec(
		`
		INSERT INTO graphs (dashboardId, name, event, period, createdOn)
		values (?, ?, ?, ?, ?)
		`,
		dashboardId, name, event, period, formattedTime)

	if err != nil {
		panic(err)
	}

	return nil
}
