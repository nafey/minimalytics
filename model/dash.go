package model

import (
	"database/sql"
	"errors"
	"time"
)

type Dashboard struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedOn string `json:"createdOn"`
}

type DashboardGet struct {
	Id        int64   `json:"id"`
	Name      string  `json:"name"`
	CreatedOn string  `json:"createdOn"`
	Graphs    []Graph `json:"graphs"`
}

type DashboardUpdate struct {
	Name string `json:"name"`
}

type DashboardCreate struct {
	Name string `json:"name"`
}

func InitDashboards() error {
	query := `
		CREATE TABLE IF NOT EXISTS dashboards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			createdOn TEXT
		);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	_, err = CreateDashboard(DashboardCreate{
		Name: "Example Dashboard",
	})

	return err
}

func IsValidDashboardId(dashboardId int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (
		SELECT 1
		FROM dashboards
		WHERE id = ?
	);`

	err := db.QueryRow(query, dashboardId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GetDashboards() []Dashboard {
	rows, err := db.Query("select * from dashboards")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var dashboards []Dashboard
	for rows.Next() {
		var dash Dashboard
		err := rows.Scan(&dash.Id, &dash.Name, &dash.CreatedOn) // Replace with actual fields

		if err != nil {
			// Handle scan error
			panic(err)
		}
		dashboards = append(dashboards, dash)
	}

	return dashboards
}

func GetDashboard(dashboardId int64) (DashboardGet, error) {
	row := db.QueryRow("select * from dashboards where id = ?", dashboardId)

	var dashboard DashboardGet
	err := row.Scan(&dashboard.Id, &dashboard.Name, &dashboard.CreatedOn)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dashboard, errors.New("Invalid dashboardId")
		}
	}

	var graphs []Graph
	graphs, err = GetDashboardGraphs(dashboardId)

	dashboard.Graphs = graphs

	return dashboard, err
}

func UpdateDashboard(dashboardId int64, updateDashboard DashboardUpdate) error {
	_, err := GetDashboard(dashboardId)
	if err != nil {
		return err
	}

	name := updateDashboard.Name

	if name != "" {
		// Add validation if needed
	}

	_, err = db.Exec(`
		UPDATE dashboards
		set name = coalesce(NULLIF(?, ''), name)
		where id = ?`,
		name, dashboardId)

	if err != nil {
		return (err)
	}

	return nil
}

func CreateDashboard(createDashboard DashboardCreate) (DashboardGet, error) {
	name := createDashboard.Name
	var dash DashboardGet

	if name == "" {
		return dash, errors.New("Invalid name for Dashboard")
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	result, err := db.Exec(
		`
		INSERT INTO dashboards (name, createdOn)
		values (?, ?)
		`,
		name, formattedTime)
	if err != nil {
		return dash, err
	}

	dashboardId, err := result.LastInsertId()
	if err != nil {
		return dash, err
	}

	dash, err = GetDashboard(dashboardId)

	return dash, err
}

func DeleteDashboard(dashboardId int64) error {
	_, err := GetDashboard(dashboardId)
	if err != nil {
		return err
	}

	graphs, err := GetDashboardGraphs(dashboardId)
	if err != nil {
		return err
	}

	for _, graph := range graphs {
		graphId := graph.Id
		DeleteGraph(graphId)
	}

	_, err = db.Exec(
		`
		DELETE FROM dashboards where id = ?
		`,
		dashboardId)

	return err
}

func IsValidDashboard(dashboardId int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (
		SELECT 1
		FROM dashboards
		WHERE id = ?
	);`

	err := db.QueryRow(query, dashboardId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
