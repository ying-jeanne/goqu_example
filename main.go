package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/doug-martin/goqu"
	_ "github.com/mattn/go-sqlite3"
)

const file string = "grafana.db"

type Team struct {
	ID        int       `db:"id" goqu:"skipinsert"`
	Name      string    `db:"name"`
	OrgID     int       `db:"org_id"`
	CreatedAt time.Time `db:"created"`
	UpdatedAt time.Time `db:"updated"`
	Email     string    `db:"email"`
}

func insertTeam(db *goqu.Database, team Team) error {
	// here for team structure, when it is autoincrement, we need to use keyword skipinsert, otherwise the default value is
	// used, we would endup with teamId = 0 hmmmm it makes me feel the same as xorm, convinient but with a lot of mistery default
	// behavior
	_, err := db.Insert("team").Rows(team).Executor().Exec()
	return err
}

func getTeam(db *goqu.Database, name string) Team {
	var team Team
	// for get we need to precise the where with column name and eq function, which is more precised
	ds := db.From("team").Where(goqu.C("name").Eq(name))
	found, err := ds.ScanStruct(&team)
	switch {
	case err != nil:
		fmt.Println(err.Error())
	case !found:
		fmt.Printf("No team found for name %s\n", name)
	default:
		fmt.Printf("found team: %+v\n", team)
	}
	return team
}

func updateTeam(db *goqu.Database, team Team) error {
	// here we set the entire object team into the record, it doesn't work well
	// the correct way to set the value is to pass goqu.Record with map value, so it
	// overwrite only the field that is necessary
	ds := db.Update("team").Set(goqu.Record{"name": team.Name}).Where(goqu.C("id").Eq(team.ID))
	_, err := ds.Executor().Exec()
	// this is the way to set only one field, if want to set struct, an example:
	// ds := db.Update("team").Set(team).Where(goqu.C("id").Eq(team.ID)) then it is using the default value to set the field
	// if we want absolutely no update on the field, we can use the tag goqu:"skipupdate" to omit the field all the time
	return err
}

func deleteTeam(db *goqu.Database, name string) (int, error) {
	de := db.Delete("team").Where(goqu.C("name").Eq(name)).Returning(goqu.C("id")).Executor()
	var ids []int
	err := de.ScanVals(&ids)
	if len(ids) > 0 {
		return ids[0], err
	}
	return 0, err
}

func main() {
	sqldb, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal(err)
	}
	// It is really easy to create a goqu database based on sql.DB
	db := goqu.New("sqlite3", sqldb)
	team1 := Team{Name: "myname5", OrgID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err = insertTeam(db, team1)
	if err != nil {
		log.Fatal(err)
	}

	// the difference of goqu is that we can build sql query and use the standard library
	// or build the query then scan into go struct
	team3 := getTeam(db, team1.Name)

	team2 := Team{ID: team3.ID, OrgID: 0, Name: "princess"}
	err = updateTeam(db, team2)
	if err != nil {
		log.Fatal(err)
	}
	team4 := getTeam(db, team2.Name)
	fmt.Printf("the team4 is %v after update \n", team4)
	deleteTeam(db, team2.Name)
	if err != nil {
		log.Fatal(err)
	}
}
