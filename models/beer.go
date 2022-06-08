package models

type Beer struct {
	Id          int64   `json:"id" uri:"id"`
	Name        string  `json:"name"`
	Tagline     string  `json:"tagline"`
	Description string  `json:"description"`
	ABV         float64 `json:"abv"`
	IBU         int64   `json:"ibu"`
}
type MyBeers struct {
	Beers []Beer
}
