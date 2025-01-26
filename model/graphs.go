package model

type Graph struct {
	Id int64
	DashboardId int64
	Name string
	Event string
	Period string
	CreatedOn string
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