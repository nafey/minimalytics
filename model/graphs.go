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
		err := rows.Scan(&graph.Id, &graph.DashboardId, &graph.Name, &graph.Event, &graph.Period, &graph.CreatedOn) // Replace with actual fields
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
	err := row.Scan(&graph.Id, &graph.DashboardId, &graph.Name, &graph.Event, &graph.Period, &graph.CreatedOn)

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

	_, err = db.Exec(`
		UPDATE graphs
			set name = coalesce(NULLIF(?, ''), name),
				event = coalesce(NULLIF(?, ''), event),
				period = coalesce(NULLIF(?, ''), period)
			where id = ?`,
		name, event, period, graphId)

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

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	result, err := db.Exec(
		`
		INSERT INTO graphs (dashboardId, name, event, period, createdOn)
		values (?, ?, ?, ?, ?)
		`,
		dashboardId, name, event, period, formattedTime)

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
