package controllers

import (
	. "2022Q2GO-BOOTCAMP/models"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

var mu sync.Mutex

//External API URL
const (
	urlPunkapi = "https://api.punkapi.com/v2/beers"
)

// Http Client to get Beer data from External APIs
func RunClient(c *gin.Context) {

	resp, getErr := http.Get(urlPunkapi)
	if getErr != nil {
		log.Fatal(getErr)
	}
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	apiBeersData := new(MyBeers)
	if err := json.Unmarshal(body, &apiBeersData.Beers); err != nil {
		fmt.Printf("error unmarshaling JSON: %v\n", err)
	}
	//Save data to CSV file
	csvBreweriesWriter(*apiBeersData)
	//JSON Response data from external API adapted to Beer struct
	c.IndentedJSON(http.StatusOK, apiBeersData.Beers)

}

// getBeers responds with the list of all beer as JSON.
func GetBeers(c *gin.Context) {
	beersData := csvBeersReader("./beersFromAPI.csv")
	c.IndentedJSON(http.StatusOK, beersData.Beers)
}

// getBeersConcurrently responds with the list of all beer read concurrently as JSON.
func GetBeersConcurrently(c *gin.Context) {
	beersData := csvBeersReaderConcurrently(c)
	c.IndentedJSON(http.StatusOK, beersData.Beers)
}

// getBeerById responds with the beer by Id as JSON.
func GetBeerById(c *gin.Context) {
	myBeers := csvBeersReader("./beersFromAPI.csv")
	idParam, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fmt.Println(err)
	}
	var beer Beer
	for _, beerItem := range myBeers.Beers {
		if beerItem.Id == idParam {
			beer = beerItem
		}
	}
	if (Beer{}) == beer {
		c.JSON(http.StatusNotFound, "Resource Not Found: 404")
	} else {
		c.IndentedJSON(http.StatusOK, beer)
	}
}

// Reads from csv beers file and returns a slice of beers
func csvBeersReader(filename string) MyBeers {
	// 1. Open the file
	recordFile, err := os.Open(filename)
	if err != nil {
		fmt.Println("CSV not valid: ", err)
	}
	// 2. Initialize the reader
	reader := csv.NewReader(recordFile)
	// 3. Read all the records
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("An error ocurred while reading the file: ", err)
	}
	// 4. Add records to MyFish struct
	data := new(MyBeers)
	for _, row := range records {
		id, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			fmt.Println(err)
		}
		abv, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			fmt.Println(err)
		}
		ibu, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			fmt.Println(err)
		}
		item := Beer{Id: id, Name: row[1], Tagline: row[2], Description: row[3], ABV: abv, IBU: ibu}
		data.Beers = append(data.Beers, item)
	}
	//5. Close csv file
	err = recordFile.Close()
	if err != nil {
		fmt.Println("An error encountered closing the csv file ", err)
	}
	return *data
}

//Reads beers csv file concurrently, it can filter by query param type for odd and even beer Ids
func csvBeersReaderConcurrently(c *gin.Context) MyBeers {
	qtype := c.Query("type")
	//qitems, _ := strconv.ParseInt(c.Query("items"), 10, 64)
	//qitems_per_workers, _ := strconv.ParseInt(c.Query("items_per_workers"), 10, 64)

	// Open the file
	recordFile, err := os.Open("./beersFromAPI.csv")
	if err != nil {
		fmt.Println("CSV not valid: ", err)
	}
	defer recordFile.Close()
	// Initialize the reader
	reader := csv.NewReader(recordFile)

	// Add records to MyBeers struct
	beers := new(MyBeers)

	// Read lines concurrently
	var wg sync.WaitGroup
	for {
		rStr, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("ERROR: ", err.Error())
			break
		}
		wg.Add(1)
		go func(pwg *sync.WaitGroup) {
			defer pwg.Done()
			beer := parseBeerStruct(rStr)
			mu.Lock()
			if qtype == "odd" && beer.Id%2 != 0 {
				beers.Beers = append(beers.Beers, *beer)
			} else if qtype == "even" && beer.Id%2 == 0 {
				beers.Beers = append(beers.Beers, *beer)
			} else if qtype == "" {
				beers.Beers = append(beers.Beers, *beer)
			}
			mu.Unlock()
		}(&wg)
	}
	wg.Wait()

	return *beers
}

//parses necesary int and float values for id,abv and ibu from csv string format
func parseBeerStruct(data []string) *Beer {
	id, _ := strconv.ParseInt(data[0], 10, 64)
	abv, _ := strconv.ParseFloat(data[4], 64)
	ibu, _ := strconv.ParseInt(data[5], 10, 64)

	return &Beer{
		Id:          id,
		Name:        data[1],
		Tagline:     data[2],
		Description: data[3],
		ABV:         abv,
		IBU:         ibu,
	}
}

//Writes to csvfile beersFromAPI.csv the data(MyBeers) from external Beer API
func csvBreweriesWriter(myBeers MyBeers) {
	// 1. Open the file
	recordFile, err := os.Create("./beersFromAPI.csv")
	if err != nil {
		fmt.Println("An error encountered:", err)
	}

	// 2. Initialize the writer
	writer := csv.NewWriter(recordFile)

	// 3. Write all the records from myBeers
	var data [][]string
	for _, record := range myBeers.Beers {
		row := []string{strconv.FormatInt(record.Id, 10), record.Name, record.Tagline, record.Description, strconv.FormatFloat(record.ABV, 'f', 1, 64), strconv.FormatInt(record.IBU, 10)}
		data = append(data, row)
	}
	writer.WriteAll(data)

	recordFile.Close()
}
