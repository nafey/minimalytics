package model

type Dashboard struct {
	Id int64
	Name string
	CreatedOn string
}

func GetDashboard(dasbhoardId int64) Dashboard {
	row := db.QueryRow("select * from Dashboards where id = ?", dasbhoardId)	

	var dashboardItem Dashboard
	err := row.Scan(&dashboardItem.Id, &dashboardItem.Name, &dashboardItem.CreatedOn)

	if err != nil {
		panic ("Not found conifg item")
	}

	return dashboardItem

}