package lpjdbapi

import	(
        "gopkg.in/mgo.v2"
        "gopkg.in/mgo.v2/bson"
	"strconv"
	"regexp"
)

type Predator struct {
	Type	string
	Id	string
	Realname	string
	Points int64
}

type Animal struct {
        Type    string
        Id      string
        Realname        string
}


type Hunter struct {
	Name string
} 

type Pass struct {
	Pass string
	Lat string
	Lon string
}

type Kill struct	{
	Animal Predator
	Date string
	Hunter string
	Location Pass
	Q string
	Udate int64
}

type Scorer struct	{
	Name string
	Score int64
	Q int64
}

type Total struct	{
	Animal string
	Q int64
}

type TimeLineNode struct	{
	Name string
	Data []int64
}

type Year struct {

	Name string
	Start int64
	End int64
}

func getDB(name string) (db *mgo.Database, session *mgo.Session){

	session, err := mgo.Dial("localhost")
	db = session.DB(name)
        if err != nil {
                panic(err)
        }
	return db, session

}

func GetKills(hunter string, species string, limit int, year string)	(*[]Kill) {

	y := GetYear(year)
	
	udatemin := y.Start
	udatemax := y.End

	db, session := getDB("lpj")
	defer session.Close()
	k := []Kill{}

	search := bson.M{"udate":bson.M{"$gt":udatemin, "$lt":udatemax}}

	if hunter != ""	&& hunter != "all" {
		search["hunter"] = hunter
	}

	if species != "" && species != "all" {
		search["animal.id"] = species
	}

	if limit == 0	{
		limit = 99999999
	}

	db.C("kills").Find(search).Sort("-udate").Limit(limit).All(&k)
	return &k
}


func GetTimeLine(hunter string, species string, limit int, year string) (*[]TimeLineNode) {

	y := GetYear(year)

        udatemin := y.Start
        udatemax := y.End

        db, session := getDB("lpj")
        defer session.Close()

        k := []Kill{}

	search := bson.M{"udate":bson.M{"$gt":udatemin, "$lt":udatemax}}

        if hunter != "" && hunter != "all" {
                search["hunter"] = hunter
        }

        if species != "" && species != "all" {
                search["animal.id"] = species
        }

        if limit == 0   {
                limit = 99999999
        }

        db.C("kills").Find(search).Sort("-udate").Limit(limit).All(&k)

	months := []string{"[0-9]{4}-07-[0-9]{2}","[0-9]{4}-08-[0-9]{2}","[0-9]{4}-09-[0-9]{2}","[0-9]{4}-10-[0-9]{2}","[0-9]{4}-11-[0-9]{2}","[0-9]{4}-12-[0-9]{2}","[0-9]{4}-01-[0-9]{2}","[0-9]{4}-02-[0-9]{2}","[0-9]{4}-03-[0-9]{2}","[0-9]{4}-04-[0-9]{2}","[0-9]{4}-05-[0-9]{2}","[0-9]{4}-06-[0-9]{2}"}
	r := []TimeLineNode{}

	var m map[string]TimeLineNode
	m = make(map[string]TimeLineNode)
	var i int = 0
	var month string = ""

	for _,month = range months	{
		for _,kill := range k	{
			if _, ok := m[kill.Animal.Realname]; !ok {
    				m[kill.Animal.Realname] = TimeLineNode{Name:kill.Animal.Realname, Data:[]int64{0,0,0,0,0,0,0,0,0,0,0,0}}
			}

			match,_ := regexp.MatchString(month, kill.Date)
			if match == true	{
				q,_ := strconv.Atoi(kill.Q)
				a :=  m[kill.Animal.Realname].Data[i] + int64(q)
				m[kill.Animal.Realname].Data[i] = a
			}
		}
		i++
	}

	i = 0
	for _,node := range m	{
		r = append(r, node);
		i++
	}

        return &r
}


func GetTotals(hunter string, species string, year string)(*[]Total)	{


        y := GetYear(year)

        udatemin := y.Start
        udatemax := y.End

        db, session := getDB("lpj")
        defer session.Close()

        t := []Total{}
        k := []Kill{}

        var m = make(map[string]int64)
        search := bson.M{"udate":bson.M{"$gt":udatemin, "$lt":udatemax}}

        if hunter != "" && hunter != "all" {
                search["hunter"] = hunter
        }

        if species != "" && species != "all"       {
                search["animal.id"] = species
        }

        db.C("kills").Find(search).All(&k)

        for _, kill := range k  {

                q,_ := strconv.Atoi(kill.Q)
                m[kill.Animal.Realname] += int64(q)
        }

	for name, quantity := range m	{
		t = append(t, Total{Animal:name, Q:quantity})
	}

	return &t

}

func GetTopTenForSpecies(year string, species string)(*[10]Scorer) {

        y := GetYear(year)

        udatemin := y.Start
        udatemax := y.End

        db, session := getDB("lpj")
        defer session.Close()

        s := [10]Scorer{}
        k := []Kill{}

        var m = make(map[string]int64)

        search := bson.M{"udate":bson.M{"$gt":udatemin, "$lt":udatemax}, "animal.id": species}

        db.C("kills").Find(search).All(&k)

        for _, kill := range k  {
                q,_ := strconv.Atoi(kill.Q)
                m[kill.Hunter] += (int64(q) * kill.Animal.Points)
        }

        var highest int64 = 0

        for i := 0; i < 10; i++ {

                if len(m)> 0    {
                        t := Scorer{}
                        highest = 0
                        for name, score := range m      {
                                if score != 0   {
                                        if score > highest      {
                                                t.Name = name
                                                t.Score = score
                                                highest = score
                                        }
                                }
                        }
                        s[i] = t
                        delete(m, t.Name)
                }
        }

        return &s
}


func GetTopTen(year string)(*[10]Scorer) {

        y := GetYear(year)

        udatemin := y.Start
        udatemax := y.End

        db, session := getDB("lpj")
        defer session.Close()

        s := [10]Scorer{}
        k := []Kill{}

	var m = make(map[string]int64)

	search := bson.M{"udate":bson.M{"$gt":udatemin, "$lt":udatemax}}

        db.C("kills").Find(search).All(&k)

	for _, kill := range k	{
		q,_ := strconv.Atoi(kill.Q)
		m[kill.Hunter] += (int64(q) * kill.Animal.Points)
	}

	var highest int64 = 0

	for i := 0; i < 10; i++	{

		if len(m)> 0	{
			t := Scorer{}
			highest = 0
			for name, score := range m	{
				if score != 0	{
					if score > highest	{
						t.Name = name
						t.Score = score
						highest = score
					}
				}
			}
			s[i] = t
			delete(m, t.Name)
		}
	}

        return &s
}


func GetHunters() (*[]Hunter) {

 
        db, session := getDB("lpj")
        defer session.Close()
	h := []Hunter{}

        db.C("hunters").Find(nil).Sort("name").All(&h)
        return &h
}

func GetPredators() (*[]Predator) {

        db, session := getDB("lpj")
        defer session.Close()

        p := []Predator{}

        db.C("predator").Find(nil).Sort("name").All(&p)
        return &p
}

func GetAnimals() (*[]Animal) {

        db, session := getDB("lpj")
        defer session.Close()

        p := []Animal{}

        db.C("animal").Find(nil).Sort("name").All(&p)
        return &p
}

func GetYears()	(*[]Year)	{
       
        db, session := getDB("lpj")
        defer session.Close()
	y := []Year{}

        db.C("year").Find(nil).All(&y)
        return &y
}

func getActiveYear()	(*Year)	{

        db, session := getDB("lpj")
        defer session.Close()

        y := Year{}

        db.C("year").Find(bson.M{"current":1}).One(&y)
        return &y

}

func GetYear(name string)	(*Year)	{
        db, session := getDB("lpj")
        defer session.Close()

        y := Year{}

	if(name == "")	{
		return getActiveYear()
	}

        db.C("year").Find(bson.M{"name":name}).One(&y)
        return &y
}
