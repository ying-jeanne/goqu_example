package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/doug-martin/goqu"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/profile"
)

const file string = "grafana.db"
const numrecords = 2000

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
	_, err := db.From("team").Insert(team).Exec()
	return err
}

func getTeam(db *goqu.Database, name string) (bool, Team, error) {
	var team Team
	// for get we need to precise the where with column name and eq function, which is more precised
	ds := db.From("team").Where(goqu.I("name").Eq(name))
	found, err := ds.ScanStruct(&team)
	// switch {
	// case err != nil:
	// 	fmt.Println(err.Error())
	// case !found:
	// 	fmt.Printf("No team found for name %s\n", name)
	// default:
	// 	fmt.Printf("found team: %+v\n", team)
	// }
	return found, team, err
}

func updateTeam(db *goqu.Database, team Team) error {
	// here we set the entire object team into the record, it doesn't work well
	// the correct way to set the value is to pass goqu.Record with map value, so it
	// overwrite only the field that is necessary
	ds := db.From("team").Where(goqu.I("id").Eq(team.ID)).Update(goqu.Record{"name": team.Name})
	_, err := ds.Exec()
	// this is the way to set only one field, if want to set struct, an example:
	// ds := db.Update("team").Set(team).Where(goqu.C("id").Eq(team.ID)) then it is using the default value to set the field
	// if we want absolutely no update on the field, we can use the tag goqu:"skipupdate" to omit the field all the time
	return err
}

func deleteTeam(db *goqu.Database, name string) error {
	_, err := db.From("team").Where(goqu.I("name").Eq(name)).Delete().Exec()
	return err
}

func main() {
	// create a sql.DB sqlite3 driver
	sqldb, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal(err)
	}

	// It is really easy to create a goqu database based on sql.DB, create a db from the sql.DB object
	db := goqu.New("sqlite3", sqldb)

	// The test senarios:
	// 1. Create 1000 teams with go struct
	// 2. Verify after each creation so 1000 get with go struct casting
	// 3. Update 1000 teams with only name field
	// 4. Verify after update so 1000 get with go struct casting
	// 5. Delete 1000 teams with name field, return the id deleted
	// 6. Get 1000 teams which is nolonger existing

	// start timer
	start := time.Now()

	// CPU profiling by default related doc here: https://flaviocopes.com/golang-profiling/
	// defer profile.Start().Stop()

	// Memory profiling
	defer profile.Start(profile.MemProfile).Stop()

	for i := 0; i < numrecords; i++ {
		num := strconv.Itoa(i)
		team1 := Team{Name: "mynamee" + num, OrgID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		err = insertTeam(db, team1)
		// fmt.Printf("the inserted team %+v\n", team1)

		if err != nil {
			log.Fatal(err)
		}

		// The difference of goqu is that we can build sql query and use the standard library
		// or build the query then scan into go struct
		_, team3, err := getTeam(db, team1.Name)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("the get team %+v\n", team3)
		team2 := Team{ID: team3.ID, OrgID: 0, Name: "princess" + num}
		err = updateTeam(db, team2)
		if err != nil {
			log.Fatal(err)
		}

		_, team4, err := getTeam(db, team2.Name)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("the updated team %+v\n", team4)

		err = deleteTeam(db, team4.Name)
		if err != nil {
			log.Fatal(err)
		}

		_, _, err = getTeam(db, team4.Name)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("the deleted team %+v\n", f)
	}

	fmt.Println("the result is: ", time.Since(start))
}
